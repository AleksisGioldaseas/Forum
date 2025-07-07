document.addEventListener("DOMContentLoaded", function () {
    let postData = document.getElementById("post-data").dataset.json;
    let commentsData = document.getElementById("comments-data").dataset.json;


    try {
        let post = JSON.parse(postData);

        const activeUserRole = document.getElementById("active-user-role").dataset.value
        const activeUsername = document.getElementById("active-user-name").dataset.value


        const postContainer = document.querySelector(".post-comments-container");

        let postElement = createPostElement(post, activeUserRole, activeUsername)

        postContainer.prepend(postElement)

        //post button events
        document.querySelectorAll('.report-post-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const postEditDiv = document.getElementById('post-report-things')
                postEditDiv.style.display = 'block';
            });
        });

        document.querySelectorAll('.sendButtonReportPost').forEach(btn => {
            btn.addEventListener('click', () => {
                const postEditDiv = document.getElementById('post-report-things')
                postEditDiv.style.display = 'none';
            });
        });

        refreshComments(JSON.parse(commentsData).comments, post.is_super_report)



    } catch (error) {
        console.error("Error parsing JSON:", error);
    }


});



function refreshComments(comments, isSuper = false) {

    let commentsSection = document.getElementById("comments-section");
    commentsSection.innerHTML = ``

    if (!comments || !Array.isArray(comments) || comments.length === 0) {
        commentsSection.innerHTML = "<p>No comments yet.</p>";
    } else {
        comments.forEach(comment => {
            const activeUserRole = document.getElementById("active-user-role").dataset.value
            const activeUsername = document.getElementById("active-user-name").dataset.value
            let commentElement = createCommentElement(comment, activeUserRole, activeUsername, isSuper)

            commentsSection.appendChild(commentElement);
        });
    }

    // hookCommentEvents()
}
