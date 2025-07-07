

window.stopLoadingReportedComments = false;
let lastTriggerTimeReportedComments = 0;

function handleScrollReportedComment() {
    const now = Date.now();

    if (window.stopLoadingReportedComments == false) {
        if ((window.innerHeight + window.scrollY) >= document.body.offsetHeight - 200) {
            if (now >= lastTriggerTimeReportedComments + 1000) {
                lastTriggerTimeReportedComments = now;
                fetchMoreReportedComments()
                    .then(moreReportedComments => {
                        const activeUserRole = document.getElementById('active-user-role').dataset.value
                        const activeUserName = document.getElementById('active-user-name').dataset.value


                        if (moreReportedComments.length == 0) {

                            window.stopLoadingReportedComments = true;
                        } else {
                            if (moreReportedComments) {
                                appendReportedComments(moreReportedComments, activeUserRole, activeUserName);
                            }
                        }
                    })
                    .catch(error => {

                        window.stopLoadingReportedComments = true;
                        console.log("Error fetching more ReportedComments:", error);
                    });
            }

        }
    }

}




window.addEventListener("scroll", handleScrollReportedComment);






async function fetchReportedComments(count, page, postId, sortType) {


    const activeFilter = document.querySelector('#filtering .active');


    const data = {
        post_id: parseInt(postId),
        count: count,
        page: page,
        sort_type: sortType,
        reported_only: true
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

window.nextReportedCommentPage = 1

async function fetchMoreReportedComments() {

    try {
        const ReportedComments = await fetchReportedComments(5, window.nextReportedCommentPage, 0, "new");
        window.nextReportedCommentPage++;
        return ReportedComments;
    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}









function appendReportedComments(comments, activeUserRole, activeUserName, isSuper = false) {
    let commentsSection = document.getElementById("comments-section");


    comments.forEach(comment => {
        c = createCommentElement(comment, activeUserRole, activeUserName, isSuper)
        commentsSection.appendChild(c)
    });


}