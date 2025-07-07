
window.stopLoadingComments = false;
let lastTriggerTimeComments = 0;

function handleScrollComment() {
    const now = Date.now();
    if (window.stopLoadingComments == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTimeComments + 1000) {
                lastTriggerTimeComments = now;
                fetchMoreComments()
                    .then(moreComments => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                        const activeUserName = document.getElementById('active-user-name').dataset.value
                        
                        if (moreComments.length == 0) {
                            window.stopLoadingComments = true;
                        }else{
                            if (moreComments){
                                appendComments(moreComments, activeUserRole, activeUserName);
                            }
                        }
                    })
                    .catch(error => {
                        window.stopLoadingComments = true;
                        console.log("Error fetching more Comments:", error);
                    });
            }

        }
    }

}




window.addEventListener("scroll", handleScrollComment);