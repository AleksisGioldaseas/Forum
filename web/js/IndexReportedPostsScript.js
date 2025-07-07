

window.stopLoadingReportedPosts = false;
let lastTriggerTimeReportedPosts = 0;

function handleScrollReportedPost() {
    const now = Date.now();
    
    if (window.stopLoadingReportedPosts == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTimeReportedPosts + 1000) {
                lastTriggerTimeReportedPosts = now;
                fetchMoreReportedPosts()
                    .then(moreReportedPosts => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                        const activeUserName = document.getElementById('active-user-name').dataset.value
                        
                        
                        if (moreReportedPosts.length == 0) {
                            
                            window.stopLoadingReportedPosts = true;
                        }else{
                            if (moreReportedPosts){
                                appendPosts(moreReportedPosts, activeUserRole, activeUserName);
                            }
                        }
                    })
                    .catch(error => {
                        
                        window.stopLoadingReportedPosts = true;
                        console.log("Error fetching more ReportedPosts:", error);
                    });
            }

        }
    }

}




window.addEventListener("scroll", handleScrollReportedPost);