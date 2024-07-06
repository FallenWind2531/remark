document.addEventListener('DOMContentLoaded', function() {
    const comments = []; // 存储所有评论
    let currentPage = 1; // 当前页码
    const commentsPerPage = 5; // 每页显示的评论数
    const maxPage = 1;

    /*setInterval(() => {
        fetchComments();
    }, 5000); // 每5秒更新一次评论*/

    // 获取评论
    function fetchComments() {
        fetch(`http://localhost:8080/comment/get?page${currentPage}=&size=${commentsPerPage}`) // 假设每次获取100条评论
            .then(response => response.json())
            .then(data => {
                comments = data.data.comments; // 更新评论数组
                maxPage = Math.ceil(data.data.total / commentsPerPage);
                document.getElementById('pageInfo').textContent = `${currentPage}/${maxPage}`;
                renderComments(); // 重新渲染评论
            });
    }

    // 监听提交按钮的点击事件
    document.getElementById('submitBtn').addEventListener('click', function() {
        const username = document.getElementById('usernameInput').value;
        const comment = document.getElementById('commentInput').value;
        if (username && comment) {
            fetch('http://localhost:8080/comment/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: username, content: comment }),
            })
            .then(response => response.json())
            .then(data => {
                fetchComments(); // 重新获取评论
            });
            usernameInput.value = '';
            commentInput.value = '';
        }

    });

    // 渲染评论到页面
    function renderComments() {
        const commentSection = document.querySelector('.commentSection');
        commentSection.innerHTML = ''; // 清空当前评论
        const startIndex = 0
        const endIndex = comments.length;

        for (let i = startIndex; i < endIndex; i++) {
            const commentDiv = document.createElement('div');
            commentDiv.classList.add('comment');
            commentDiv.innerHTML = `<h3>${comments[i].username}</h3><p>${comments[i].comment}</p>`;
            const deleteBtn = document.createElement('button');
            deleteBtn.textContent = '删除';
            deleteBtn.classList.add('del');
            deleteBtn.onclick = function() {
                fetch(`http://localhost:8080/comment/delete?id=${comments[i].id}`, {
                    method: 'POST',
                })
                .then(response => response.json())
                .then(data => {
                    fetchComments(); // 重新获取评论
                });
            };
            commentDiv.appendChild(deleteBtn);
            commentSection.appendChild(commentDiv);
        }
    }

    document.body.addEventListener('click', function(next) {
        if (next.target.id === 'nextBtn') {
            if (currentPage < maxPage) {
                currentPage++;
                fetchComments();
            }
        }
    });

    document.body.addEventListener('click', function(prev) {
        if (prev.target.id === 'prevBtn') {
            if (currentPage > 1) {
                currentPage--;
                fetchComments();
            }
        }
    });

    fetchComments();
});