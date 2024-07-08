package main

import (
    "encoding/json"
    "net/http"
	"strconv"
	"sync"
)

type ApiResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"msg"`
    Data    interface{} `json:"data"`
}

type Comment struct {
    ID      int    `json:"id"`
    Name    string `json:"name"`
    Comment string `json:"content"`
}

var mutex = &sync.RWMutex{}
var comments []Comment
var incrementID int = 1

func addCommentHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        // 如果不是POST请求，返回405 Method Not Allowed
        sendResponse(w, http.StatusMethodNotAllowed, "只支持POST方法", nil)
        return
    }

    // 解析请求体中的name和content
    var newComment Comment
    err := json.NewDecoder(r.Body).Decode(&newComment)
    if err != nil {
        // 如果解析失败，返回400 Bad Request
        sendResponse(w, http.StatusBadRequest, "解析请求体失败", nil)
        return
    }

    // 添加评论到comments切片中
    mutex.Lock()
    newComment.ID = incrementID
    incrementID++
    comments = append(comments, newComment)
    mutex.Unlock()

    // 返回成功的响应
    sendResponse(w, http.StatusOK, "评论添加成功", newComment)
}

func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        sendResponse(w, http.StatusMethodNotAllowed, "只支持POST方法", nil)
        return
    }

    // 从查询参数中获取id
    queryValues := r.URL.Query()
    idStr := queryValues.Get("id")
    if idStr == "" {
        sendResponse(w, http.StatusBadRequest, "缺少id参数", nil)
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        sendResponse(w, http.StatusBadRequest, "id参数无效", nil)
        return
    }

    // 删除具有给定ID的评论
    deleted := deleteCommentByID(id)
    if !deleted {
        sendResponse(w, http.StatusNotFound, "未找到指定ID的评论", nil)
        return
    }

    // 发送成功响应
    sendResponse(w, http.StatusOK, "评论删除成功", nil)
}

// deleteCommentByID 在comments切片中查找并删除具有给定ID的评论
// 如果找到并删除了评论，则返回true；否则返回false
func deleteCommentByID(id int) bool {
    mutex.Lock()
    defer mutex.Unlock()
    for i, comment := range comments {
        if comment.ID == id {
            comments = append(comments[:i], comments[i+1:]...)
            return true
        }
    }
    return false
}

func getCommentHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        sendResponse(w, http.StatusMethodNotAllowed, "只支持GET方法", nil)
        return
    }

    // 从查询参数中获取page和size
    queryValues := r.URL.Query()
    pageStr := queryValues.Get("page")
    sizeStr := queryValues.Get("size")

    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        sendResponse(w, http.StatusBadRequest, "page参数无效", nil)
        return
    }

    size, err := strconv.Atoi(sizeStr)
    if err != nil || (size < -1 || size == 0) { // 允许size为-1，表示获取所有评论
        sendResponse(w, http.StatusBadRequest, "size参数无效", nil)
        return
    }

	mutex.Lock()
    defer mutex.Unlock()

	if size == -1 {
		responseData := map[string]interface{}{
			"total":    len(comments),
			"comments": comments,
		}
        sendResponse(w, http.StatusOK, "获取评论成功", responseData)
        return
    }

    // 计算当前页的评论范围
    start := (page - 1) * size
    end := start + size
    if start >= len(comments) {
        sendResponse(w, http.StatusOK, "超出评论范围", map[string]interface{}{"total": len(comments), "comments": []Comment{}})
        return
    }
    if end > len(comments) {
        end = len(comments)
    }

    // 构造响应体
    responseData := map[string]interface{}{
        "total":    len(comments),
        "comments": comments[start:end],
    }

    sendResponse(w, http.StatusOK, "获取评论成功", responseData)
}

func sendResponse(w http.ResponseWriter, code int, message string, data interface{}) {
    response := ApiResponse{
        Code:    code,
        Message: message,
        Data:    data,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(response)
}

func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 允许来自任何源的请求
        w.Header().Set("Access-Control-Allow-Origin", "*")
        // 允许预检请求中的方法
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        // 允许预检请求中的头信息
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

        // 如果是预检请求，直接返回
        if r.Method == "OPTIONS" {
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/comment/get", getCommentHandler)
    mux.HandleFunc("/comment/add", addCommentHandler)
	mux.HandleFunc("/comment/delete", deleteCommentHandler)

    handler := enableCORS(mux)

    http.ListenAndServe(":8080", handler)
}