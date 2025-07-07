
window.stopLoadingRemovedComments = false;
let lastTriggerTimeRemovedComments = 0;

function handleScrollRemovedComment() {
    const now = Date.now();
    if (window.stopLoadingRemovedComments == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTimeRemovedComments + 1000) {
                lastTriggerTimeRemovedComments = now;
                fetchMoreRemovedComments()
                    .then(moreRemovedComments => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                        const activeUserName = document.getElementById('active-user-name').dataset.value
                        
                        if (moreRemovedComments.length == 0) {
                            window.stopLoadingRemovedComments = true;
                        }else{
                            if (moreRemovedComments){
                                appendRemovedComments(moreRemovedComments, activeUserRole, activeUserName);
                            }
                        }
                    })
                    .catch(error => {
                        window.stopLoadingRemovedComments = true;
                        console.log("Error fetching more RemovedComments:", error);
                    });
            }

        }
    }

}




window.addEventListener("scroll", handleScrollRemovedComment);






async function fetchRemovedComments(count, page, postId, sortType) {

    const activeFilter = document.querySelector('#filtering .active');
    

    const data = {
        post_id: parseInt(postId),
        count: count,
        page: page,
        sort_type: sortType,
        removed_only: true
    };

    const responseElement = document.createElement('div');
    
    try {
        await sendJsonRequest("/commentlist", data, responseElement, false)
        return JSON.parse(responseElement.textContent).data.comments
    } catch (err) {
        console.error("Error:", err);
        throw err; // Re-throw to let caller handle it
    }


}

window.nextRemovedCommentPage = 1

async function fetchMoreRemovedComments() {

    try {
    
      

        const RemovedComments = await fetchRemovedComments(5, window.nextRemovedCommentPage, 0, "new");
        window.nextRemovedCommentPage++;
        return RemovedComments;
    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}










function appendRemovedComments(comments, activeUserRole, activeUserName, isSuper = false) {
    let commentsSection = document.getElementById("comments-section");

   
    comments.forEach(comment => {
        c = createCommentElement(comment, activeUserRole, activeUserName, isSuper)
        commentsSection.appendChild(c)
    });

  
}