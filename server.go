package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

type Comment struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	Comment string `json:"content"`
}

var db *gorm.DB
var err error

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("no such config file")
		} else {
			log.Println("read config error")
		}
		log.Fatal(err)
	}
}

func initDB() {
	db, err = gorm.Open(sqlite.Open(viper.GetString(`database.path`)), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Comment{})
}

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

	result := db.Create(&newComment)
	if result.Error != nil {
		sendResponse(w, http.StatusInternalServerError, "添加评论失败", nil)
		return
	}

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
	var comment Comment
	db.First(&comment, id)
	if comment.ID == 0 {
		sendResponse(w, http.StatusNotFound, "未找到指定ID的评论", nil)
		return
	}

	result := db.Delete(&comment)
	if result.Error != nil {
		sendResponse(w, http.StatusInternalServerError, "删除评论失败", nil)
		return
	}

	sendResponse(w, http.StatusOK, "评论删除成功", nil)
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

	var total int64
	db.Model(&Comment{}).Count(&total)

	var comments []Comment

	if size == -1 {
		// 如果size为-1，获取所有评论，不使用分页
		result := db.Find(&comments)
		if result.Error != nil {
			sendResponse(w, http.StatusInternalServerError, "获取评论失败", nil)
			return
		}
	} else {
		offset := (page - 1) * size
		result := db.Offset(offset).Limit(size).Find(&comments)
		if result.Error != nil {
			sendResponse(w, http.StatusInternalServerError, "获取评论失败", nil)
			return
		}
	}

	responseData := map[string]interface{}{
		"total":    total,
		"comments": comments,
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
	initDB()

	mux := http.NewServeMux()

	mux.HandleFunc("/comment/get", getCommentHandler)
	mux.HandleFunc("/comment/add", addCommentHandler)
	mux.HandleFunc("/comment/delete", deleteCommentHandler)

	handler := enableCORS(mux)

	http.ListenAndServe(viper.GetString(`server.port`), handler)
}
