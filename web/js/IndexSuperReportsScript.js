

window.stopLoadingSuperReports = false;
let lastTriggerTimeSuperReports = 0;

function handleScrollSuperReport() {
    const now = Date.now();
    
    if (window.stopLoadingSuperReports == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTimeSuperReports + 1000) {
                lastTriggerTimeSuperReports = now;
                fetchMoreSuperReports()
                    .then(moreSuperReports => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                        const activeUserName = document.getElementById('active-user-name').dataset.value
                        
                        
                        if (moreSuperReports.length == 0) {
                            
                            window.stopLoadingSuperReports = true;
                        }else{
                            if (moreSuperReports){
                                appendPosts(moreSuperReports, activeUserRole, activeUserName);
                            }
                        }
                    })
                    .catch(error => {
                        
                        window.stopLoadingSuperReports = true;
                        console.log("Error fetching more SuperReports:", error);
                    });
            }

        }
    }

}




window.addEventListener("scroll", handleScrollSuperReport);