document.addEventListener('DOMContentLoaded', loadSuperReportTarget);


function loadSuperReportTarget(){
    let container = document.getElementById("target-container")
    const activeUserRole = document.getElementById("active-user-role").dataset.value
    const activeUsername = document.getElementById("active-user-name").dataset.value


    let postJson = document.getElementById("target-post-json").dataset.value;
    let post = JSON.parse(postJson)
    let commentJson = document.getElementById("target-comment-json").dataset.value;
    let comment = JSON.parse(commentJson)
    let userJson = document.getElementById("target-user-json").dataset.value;
    let user = JSON.parse(userJson)

    if (post){
        postElem = createPostElement(post, activeUserRole, activeUsername)
        container.appendChild(postElem)
    }

    if (comment){
        commentElem = createCommentElement(comment, activeUserRole, activeUsername, true)
        container.appendChild(commentElem)
    }

    if (user){
        userElem = createUserElement(user)
        container.appendChild(userElem)
    }

}