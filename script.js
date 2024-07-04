document.addEventListener('DOMContentLoaded', function() {
    const comments = []; // 存储所有评论
    let currentPage = 1; // 当前页码
    const commentsPerPage = 5; // 每页显示的评论数

    // 监听提交按钮的点击事件
    document.getElementById('submitBtn').addEventListener('click', function() {
        const username = document.getElementById('usernameInput').value;
        const comment = document.getElementById('commentInput').value;
        if (username && comment) {
            comments.push({ username, comment }); // 将评论添加到数组
            renderComments(); // 渲染评论
            const maxPage = Math.ceil(comments.length / commentsPerPage);
            document.getElementById('pageInfo').textContent = `${currentPage}/${maxPage}`;
            usernameInput.value = '';
            commentInput.value = '';
        }

    });

    // 渲染评论到页面
    function renderComments() {
        const commentSection = document.querySelector('.commentSection');
        commentSection.innerHTML = ''; // 清空当前评论
        const startIndex = (currentPage - 1) * commentsPerPage;
        const endIndex = Math.min(startIndex + commentsPerPage, comments.length);

        for (let i = startIndex; i < endIndex; i++) {
            const commentDiv = document.createElement('div');
            commentDiv.classList.add('comment');
            commentDiv.innerHTML = `<h3>${comments[i].username}</h3><p>${comments[i].comment}</p>`;
            const deleteBtn = document.createElement('button');
            deleteBtn.textContent = '删除';
            deleteBtn.classList.add('del');
            deleteBtn.onclick = function() {
                comments.splice(i, 1);
                renderComments();
                let maxPage = Math.ceil(comments.length / commentsPerPage);
                maxPage = maxPage === 0 ? 1 : maxPage;
                document.getElementById('pageInfo').textContent = `${currentPage}/${maxPage}`;
            };
            commentDiv.appendChild(deleteBtn);
            commentSection.appendChild(commentDiv);
        }
    }

    document.body.addEventListener('click', function(next) {
        if (next.target.id === 'nextBtn') {
            const maxPage = Math.ceil(comments.length / commentsPerPage);
            if (currentPage < maxPage) {
                currentPage++;
                renderComments();
                document.getElementById('pageInfo').textContent = `${currentPage}/${maxPage}`;
            }
        }
    });

    document.body.addEventListener('click', function(prev) {
        if (prev.target.id === 'prevBtn') {
            if (currentPage > 1) {
                currentPage--;
                renderComments();
                const maxPage = Math.ceil(comments.length / commentsPerPage);
                document.getElementById('pageInfo').textContent = `${currentPage}/${maxPage}`;
            }
        }
    });
});