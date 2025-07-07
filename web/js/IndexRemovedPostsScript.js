

window.stopLoadingRemovedPosts = false;
let lastTriggerTimeRemovedPosts = 0;

function handleScrollRemovedPost() {
    const now = Date.now();
    if (window.stopLoadingRemovedPosts == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTimeRemovedPosts + 1000) {
                lastTriggerTimeRemovedPosts = now;
                fetchMoreRemovedPosts()
                    .then(moreRemovedPosts => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                        const activeUserName = document.getElementById('active-user-name').dataset.value
                        
                        
                        if (moreRemovedPosts.length == 0) {
                            
                            window.stopLoadingRemovedPosts = true;
                        }else{
                            if (moreRemovedPosts){
                                appendPosts(moreRemovedPosts, activeUserRole, activeUserName);
                            }
                        }
                    })
                    .catch(error => {
                        
                        window.stopLoadingRemovedPosts = true;
                        console.log("Error fetching more RemovedPosts:", error);
                    });
            }

        }
    }

}




window.addEventListener("scroll", handleScrollRemovedPost);