



//################################## BIG POST
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################

function dh(html) {
    const entities = {
        '&#39;': "'",
        '&lt;': '<',
        '&gt;': '>',
        '&amp;': '&',
        // add more if needed
    };
    return html.replace(/&#39;|&lt;|&gt;|&amp;/g, match => entities[match]);
}

function createPostElement(post, activeUserRole, activeUsername) {
    // Create main wrapper element
    const postWrapper = document.createElement('div');
    postWrapper.className = 'post-wrapper';
    postWrapper.dataset.postId = post.id;
    postWrapper.dataset.userReaction = post.user_reaction || 0;




    // Build removal notice if needed
    let removalNotice = ``
    if (post.removed == true) {
        if (activeUserRole >= 2) {
            removalNotice = `
        <p id="remove-tag">REMOVED</p>
        <p id="remove-reason">by: ${post.mod_name} for reason: ${post.removal_reason}</p>`;
        }
    }


    // Build post image if exists
    const postImage = (post.post_img && !["<nil>", "null", "undefined"].includes(post.post_img)) ?
        `<img id="post-image" class="post-image" src="/image/${post.post_img}">` :
        '';

    // Build post actions based on user permissions
    const postActions = buildPostActions(post, activeUserRole, activeUsername);

    const voteAndCommentWrapper = document.createElement("div");
    voteAndCommentWrapper.classList.add("vote-comment-wrapper");

    const reactionBox = ReactionElement(post)
    if (post.is_super_report) {
        reactionBox.innerHTML = ''
    }
    voteAndCommentWrapper.appendChild(reactionBox)

    const commentCounter = createCommentCountElement(post)
    voteAndCommentWrapper.appendChild(commentCounter)

    postWrapper.appendChild(voteAndCommentWrapper)

    // Main post HTML structure
    postWrapper.innerHTML += `
        <div class="post-comments-container">
            <div class="post-container">
                ${removalNotice}
                <h1 id="post-title">${post.title}</h1>
                <div class="post-details">
                    <p><strong>Posted by:</strong> <a ${(post.user_name !== "(deleted)" && post.user_name !== "(removed)") ? `href="/profile/${post.user_name}"` : ``}>${post.user_name}</a></p>
                    <p><strong>Category:</strong> ${post.categories ? post.categories.join(", ") : "None"}</p>
                    <div id="post-body">${post.body}</div>
                    <div style="display:none;" id="post-edit-things">
                        <textarea id="postNewText" rows="10" cols="100">${post.body}</textarea>
                        <button style="display: block;" data-value="${post.id}" id="sendButtonEditpost">Save</button>
                    </div>
                    ${postImage}
                    <p id="creation-date"><strong>Posted on:</strong> ${new Date(post.creation_date).toLocaleString()}</p>
                </div>
            </div>
            <div id="post-actions">${postActions}</div>
            <div style="display:none;" id="post-report-things">
                <textarea id="postReportText" rows="3" cols="20"></textarea>
                <button style="display: block;" data-value="${post.id}" class="sendButtonReportPost">Save</button>
            </div>
        </div>
    `;

    // Set up interactive elements
    // setupPostInteractivity(postWrapper, post);

    return postWrapper;
}


function createCommentCountElement(post) {
    const commentButton = document.createElement("div");
    commentButton.classList.add("comments-count");
    const commentIcon = document.createElement("i");
    commentIcon.classList.add("fa-solid", "fa-comment");
    const commentCount = document.createElement("span");
    commentCount.textContent = post.comment_count || 0;
    commentButton.appendChild(commentIcon);
    commentButton.appendChild(commentCount);
    commentButton.addEventListener("click", () => {
        window.location.href = `/post/${post.id}#comments`;
    });
    return commentButton
}




/**
 * Sets up global event delegation for all like/dislike buttons.
 * Call this once when the page loads.
 */
function setupReactionHandlers(ispost = true) {
    // Listen for clicks on the document (or a closer static parent)
    document.addEventListener('click', function (e) {

        const isLoggedIn = document.getElementById('auth-status').getAttribute('data-logged-in') === 'true';

        if (!e.target.matches('.post-vote')) {
            return
        }

        // Check if the click was on a like/dislike button
        const likeButton = e.target.closest('.like-button');
        const dislikeButton = e.target.closest('.dislike-button');

        if ((likeButton || dislikeButton) && !isLoggedIn) {
            const authButton = document.getElementById("auth-btn");
            if (authButton) {
                authButton.click(); // Simulate clicking the login button
            }
            return;
        }


        if (likeButton) {
            handleReactionClick(likeButton, 'like', ispost);
        }
        else if (dislikeButton) {
            handleReactionClick(dislikeButton, 'dislike', ispost);
        }
    });
}

/**
 * Handles the actual vote logic
 * @param {HTMLElement} button - The clicked button
 * @param {string} reactionType - 'like' or 'dislike'
 */
function handleReactionClick(button, reactionType, ispost = true) {


    // Find the parent post wrapper
    let wrapper = button.closest('.post-wrapper');
    if (!wrapper) {
        wrapper = button.closest('.outer-comment-wrapper');
    }

    // const postId = wrapper.dataset.postId;



    // Call your vote handler
    handleVote(button, reactionType, ispost);

    // Update UI immediately (optimistic update)
    updateButtonStyles(button, wrapper, reactionType);
}


/* Updates button styles after click (visual feedback)  */

function updateButtonStyles(clickedButton, wrapper, reactionType) {

    // Update UI counts correctly
    const likeButton = wrapper.querySelector(".like-button");
    const dislikeButton = wrapper.querySelector(".dislike-button");
    const likeCount = wrapper.querySelector(".like-count");
    const dislikeCount = wrapper.querySelector(".dislike-count");
    const scoreDisplay = wrapper.querySelector(".score-display"); // Select the score display element

    let newAction = reactionType;

    if (clickedButton.classList.contains("active")) {
        newAction = "neutral"; // User clicked the same active button → remove vote
    }

    let newLikeCount = parseInt(likeCount.textContent);

    let newDislikeCount = parseInt(dislikeCount.textContent);

    if (newAction === "neutral") {
        if (likeButton.classList.contains("active")) {
            newLikeCount -= 1;
            likeButton.classList.remove("active");
        } else if (dislikeButton.classList.contains("active")) {
            newDislikeCount -= 1;
            dislikeButton.classList.remove("active");
        }
    } else if (newAction === "like") {
        if (dislikeButton.classList.contains("active")) {
            newDislikeCount -= 1;
            dislikeButton.classList.remove("active");
        }
        if (!likeButton.classList.contains("active")) {
            newLikeCount += 1;
        }
        likeButton.classList.add("active");
    } else if (newAction === "dislike") {
        if (likeButton.classList.contains("active")) {
            newLikeCount -= 1;
            likeButton.classList.remove("active");
        }
        if (!dislikeButton.classList.contains("active")) {
            newDislikeCount += 1;
        }
        dislikeButton.classList.add("active");
    }

    // Update the like and dislike counts in the UI
    likeCount.textContent = newLikeCount;
    dislikeCount.textContent = newDislikeCount;

    // Update the score dynamically
    const score = newLikeCount - newDislikeCount;
    scoreDisplay.textContent = Math.abs(score);  // Update score display

    // Apply background color based on score
    if (score < 0) {
        scoreDisplay.classList.remove("zero");
        scoreDisplay.classList.add("negative");
    } else if (score === 0) {
        scoreDisplay.classList.remove("negative");
        scoreDisplay.classList.add("zero");
    } else {
        scoreDisplay.classList.remove("negative", "zero");
    }

    // Update or recreate the tooltip after voting
    updateTooltip(scoreDisplay, wrapper);
}


// Function to update or recreate the tooltip
function updateTooltip(scoreDisplay, postWrapper) {
    // If tooltip already exists, just update it
    let tooltip = scoreDisplay.querySelector(".tooltip");
    if (!tooltip) {
        // Create the tooltip if it doesn't exist
        tooltip = document.createElement("div");
        tooltip.classList.add("tooltip");
        scoreDisplay.appendChild(tooltip);
    }

    // Get updated likes and dislikes
    const likes = postWrapper.querySelector(".like-count").textContent;
    const dislikes = postWrapper.querySelector(".dislike-count").textContent;

    // Update the tooltip content
    tooltip.textContent = `Likes: ${likes} \nDislikes: ${dislikes}`;
    tooltip.style.whiteSpace = 'pre-line';  // This allows the text to respect new lines
}

document.addEventListener("DOMContentLoaded", function () {
    setupReactionHandlers()
});





















//##########################   POSTS
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################


function renderMiniPost(post) {
    const postBox = document.createElement("div");
    postBox.classList.add("post-box");

    let commentButton = createCommentCountElement(post)

    const postWrapper = document.createElement("div");
    postWrapper.classList.add("post-wrapper");
    postWrapper.setAttribute("data-post-id", post.id);
    postWrapper.setAttribute("data-user-reaction", post.user_reaction || 0);

    const likeDislikeContainer = ReactionElement(post)

    if (post.is_super_report) {
        likeDislikeContainer.innerHTML = ``
    }

    const voteAndCommentWrapper = document.createElement("div");
    voteAndCommentWrapper.classList.add("vote-comment-wrapper");
    voteAndCommentWrapper.appendChild(likeDislikeContainer);
    voteAndCommentWrapper.appendChild(commentButton);
    postWrapper.appendChild(voteAndCommentWrapper);

    const categoriesArray = Array.isArray(post.categories) ? post.categories : [];
    const categoriesList = categoriesArray.length > 0
        ? categoriesArray.map(cat => `<span class="category">${cat}</span>`).join(", ")
        : "<em>No categories</em>";



    if (post.is_super_report == true) {
        postBox.innerHTML = `<a id="post-title-url" href="/superreport/${post.id}"><h2 class="post-title">${post.title}</h2>
        </a>`
    } else {
        postBox.innerHTML = `<a id="post-title-url" href="/post/${post.id}"><h2 class="post-title">${post.title}</h2>
        </a>`
    }
    postBox.innerHTML += `

        
            
        ${post.post_img ? `<img src="/image/${post.post_img}" alt="${post.title} image" />` : ''}
        <div class="post-user"><strong>Posted by:</strong> <a id="post-title-url" ${(post.user_name !== "(deleted)" && post.user_name !== "(removed)") ? `href="/profile/${post.user_name}"` : ``}>${post.user_name}</a></div>
        <p class="post-categories"><strong>Categories:</strong> ${categoriesList}</p>
        <div class="post-details">
            <p>${post.body}</p>
            <p style="display:none;"><strong>Likes:</strong> ${post.likes} | <strong>Dislikes:</strong> ${post.dislikes}  | <strong>ActiveuserReaction:</strong> ${post.user_reaction} </p>
            <p><strong>Posted on:</strong> ${new Date(post.creation_date).toLocaleString()}</p>
        </div>`;


    if (post.reports) {
        let split = post.reports.join('<br>');
        if (split.length > 0) {
            postBox.innerHTML += `<p>${post.reports.length} reports:</p>` + split
        }
    }




    postWrapper.appendChild(postBox);

    return postWrapper
}




// Helper function to build post action buttons
function buildPostActions(post, activeUserRole, activeUsername) {
    let actions = '';

    // Mod actions
    if (activeUserRole > 1) {
        if (post.removed === 0) {
            actions += `
                <div class="removal-reasons" style="display: none;">
                    <p>Pick removal reason:</p>
                    ${['Obscene', 'Irrelevant', 'Insulting', 'Illegal', 'Other']
                    .map(reason => `<button data-value="${post.id}" class="final-remove-post-btn">${reason}</button>`)
                    .join('')}
                </div>
                <button data-value="${post.id}" class="remove-post-btn">Remove</button>
            `;
        } else {
            actions += `<button data-value="${post.id}" class="approve-post-btn">Approve</button>`;
        }
    }

    // Owner actions
    if (activeUsername === post.user_name) {
        actions += `
            <button class="edit-post-btn">Edit</button>
            <button data-value="${post.id}" class="delete-post-btn">Delete</button>
        `;
    }

    // Reporting
    if (activeUserRole > 0) {
        actions += `<button class="report-post-btn">Report</button>`;

        if (activeUserRole > 1) {
            actions += `<a class="super-report-link" href="/superreportform?post_id=${post.id}">
                        <button>Super Report</button>
                    </a>`
        }

        if (post.reports?.length) {
            actions += `<p>${post.reports.length} reports:</p>${post.reports.join('<br>')}`;
        }
    }

    return actions;
}


function ReactionElement(dataObject, horizontal = false) {
    const likeDislikeContainer = document.createElement("div");



    likeDislikeContainer.classList.add("like-dislike-container");



    if (horizontal) {
        likeDislikeContainer.style.display = 'flex';
        likeDislikeContainer.style.width = 'auto';
        if (dataObject.removed == true || dataObject.deleted == true) {
            likeDislikeContainer.style.width = '160px';
        }
        likeDislikeContainer.style.height = '40px';
        likeDislikeContainer.style.flexDirection = 'row';
    }


    const likeButton = document.createElement("button");
    likeButton.classList.add("like-button");
    if (horizontal) {
        likeButton.innerHTML = '<i class="fa-solid com-vote-up fa-thumbs-up"></i>';
    } else {
        likeButton.innerHTML = '<i class="fa-solid post-vote fa-thumbs-up"></i>';
    }

    const dislikeButton = document.createElement("button");
    dislikeButton.classList.add("dislike-button");
    if (horizontal) {
        dislikeButton.innerHTML = '<i class="fa-solid com-vote-down fa-thumbs-down"></i>';
    } else {
        dislikeButton.innerHTML = '<i class="fa-solid post-vote fa-thumbs-down"></i>';
    }
    const activeUserReaction = parseInt(dataObject.user_reaction);
    if (activeUserReaction === 1) {
        likeButton.classList.add("active");
    } else if (activeUserReaction === -1) {
        dislikeButton.classList.add("active");
    }




    const likeCount = document.createElement("div");
    likeCount.classList.add("like-count");
    likeCount.innerHTML = dataObject.likes;
    likeCount.style.display = "none";

    const dislikeCount = document.createElement("div");
    dislikeCount.classList.add("dislike-count");
    dislikeCount.innerHTML = dataObject.dislikes;
    dislikeCount.style.display = "none";

    const scoreDisplay = document.createElement("div");
    scoreDisplay.classList.add("score-display");
    if (dataObject.removed == true || dataObject.deleted == true) {
        scoreDisplay.style.marginLeft = '40px';
    }
    const initialScore = Math.abs(dataObject.likes - dataObject.dislikes);
    scoreDisplay.textContent = `${initialScore}`;
    if (dataObject.likes - dataObject.dislikes < 0) {
        scoreDisplay.classList.add("negative");
    } else if (dataObject.likes - dataObject.dislikes === 0) {
        scoreDisplay.classList.add("zero");
    }
    const tooltip = document.createElement("div");
    tooltip.classList.add("tooltip");
    tooltip.innerHTML = `Likes: ${dataObject.likes} <br> Dislikes: ${dataObject.dislikes}`;
    scoreDisplay.appendChild(tooltip);


    likeDislikeContainer.appendChild(likeCount);

    if (dataObject.removed == false && dataObject.deleted == false) {
        likeDislikeContainer.appendChild(likeButton);
    }
    likeDislikeContainer.appendChild(scoreDisplay);

    if (dataObject.removed == false && dataObject.deleted == false) {
        likeDislikeContainer.appendChild(dislikeButton);
    }

    likeDislikeContainer.appendChild(dislikeCount);

    return likeDislikeContainer
}






async function fetchPosts(count, page) {

    const activeFilter = document.querySelector('#filtering .active');
    const activeSort = document.querySelector('#sorting .active');

    const data = {
        count: count,
        page: page,
        filter_type: activeFilter?.dataset.value || 'all',
        sort_type: activeSort?.dataset.value || 'hot',
        search_query: document.getElementById("search-query")?.value || '',
        categories: selectedCategories
    };

    const responseElement = document.createElement('div');

    try {
        await sendJsonRequest("/postlist", data, responseElement, false);

        const jsonData = JSON.parse(responseElement.textContent);

        if (!jsonData?.data || !Array.isArray(jsonData.data)) {
            throw new Error("Invalid posts data received");
        }

        return jsonData.data; // <-- This now returns ACTUAL data
    } catch (err) {
        console.error("Error:", err);
        throw err; // Re-throw to let caller handle it
    }
}



window.nextPage = 1

async function fetchMorePosts() {

    try {

        const posts = await fetchPosts(5, window.nextPage);
        window.nextPage++;
        return posts;
    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}





















//############################### COMMENTS
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################



function createCommentElement(comment, activeUserRole, activeUsername, isSuper = false) {
    let commentDiv = document.createElement("div");
    commentDiv.classList.add("comment");

    // Create left-side like/dislike and count
    let reactionDiv = ReactionElement(comment, true)
    if (isSuper) {
        reactionDiv.innerHTML = ''
    }

    // Main comment body
    let bodyDiv = document.createElement("div");

    if (comment.removed == 1 && activeUserRole >= 2) {
        bodyDiv.innerHTML += `
                <p>REMOVED by: ${comment.mod_name} for reason: ${comment.removal_reason}</p> 
        `;
    }
    bodyDiv.innerHTML += `
            
            <p><strong><a class="comment-user" ${(comment.user_name !== "(deleted)" && comment.user_name !== "(removed)") ? `href="/profile/${comment.user_name}"` : ``}>${comment.user_name}</a>:</strong></p>
            <div class="comment-body" >${comment.body}</div>
        `;

    bodyDiv.innerHTML += `
        <div class="comment-actions">`


    if (activeUserRole > 1) {
        if (comment.removed == 0) {
            bodyDiv.innerHTML += `
                <div style="display: none; data-value="${comment.id}" class="data-comment-id">
                </div>
                <div class="removal-reasons" style="display: none;">
                    <p>Pick removal reason:</p>
                    <button data-value="${comment.id}" class="final-remove-comment-btn">Obscene</button>
                    <button data-value="${comment.id}" class="final-remove-comment-btn">Irrelevant</button>
                    <button data-value="${comment.id}" class="final-remove-comment-btn">Insulting</button>
                    <button data-value="${comment.id}" class="final-remove-comment-btn">Illegal</button>
                    <button data-value="${comment.id}" class="final-remove-comment-btn">Other</button>
                </div>

                <button class="remove-comment-btn">Remove</button>
                `

        } else {
            bodyDiv.innerHTML += `
                <button data-value="${comment.id}" class="approve-comment-btn">Approve</button>`
        }
    }
    if (activeUsername == comment.user_name && comment.removed == false && comment.deleted == false) {
        //Reminder, edit button should spawn text area and the submit button should actually send connect to the edit comment endpoint
        bodyDiv.innerHTML += `
            <div style="display: none;" class="edit-controls">
                <textarea class="newText" rows="3" cols="30">${comment.body}</textarea>
                <button data-value="${comment.id}" class="sendButtonEditComment">Save</button>
            </div>
            <button class="edit-comment-btn">Edit</button>
            <button data-value="${comment.id}" class="delete-comment-btn">Delete</button>`
    }

    if (activeUserRole > 0 && comment.removed == false && comment.deleted == 0) {
        //Reminder, report button should spawn text area and the submit button should actually send connect to the edit comment endpoint
        bodyDiv.innerHTML += `
            
            <div style="display: none;" class="report-controls">
                <textarea class="newText" rows="3" cols="30"></textarea>
                <button data-value="${comment.id}" class="sendButtonReportComment">Save</button>
            </div>
            <button class="report-comment-btn">Report</button>
            `
    }



    if (activeUserRole > 1) {
        //Reminder, report button should spawn text area and the submit button should actually send connect to the edit comment endpoint
        bodyDiv.innerHTML += `
            <a href="/superreportform?comment_id=${comment.id}"> <button>Super Report</button> </a>
            `

        let split = comment.reports.join('<br>');
        if (split.length > 0) {
            bodyDiv.innerHTML += `<p>${comment.reports.length} reports:</p>` + split
        }



    }

    bodyDiv.innerHTML += `
        </div>`

    // Layout wrapper
    let outerwrapper = document.createElement("div");
    outerwrapper.classList.add("outer-comment-wrapper");
    outerwrapper.setAttribute("data-user-reaction", comment.user_reaction) // <- added this

    outerwrapper.appendChild(reactionDiv);
    outerwrapper.style.display = "flex"
    outerwrapper.style.flexDirection = "row"

    let wrapperDiv = document.createElement("div");
    wrapperDiv.classList.add("comment-border");
    wrapperDiv.classList.add("comment-wrapper");

    wrapperDiv.appendChild(bodyDiv);

    outerwrapper.appendChild(wrapperDiv)
    outerwrapper.setAttribute("data-comment-id", comment.id)
    commentDiv.appendChild(outerwrapper);


    return commentDiv


}



async function fetchComments(count, page, postId, sortType) {

    const activeFilter = document.querySelector('#filtering .active');


    const data = {
        post_id: parseInt(postId),
        count: count,
        page: page,
        sort_type: sortType
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

window.nextCommentPage = 1

async function fetchMoreComments() {

    try {
        const postIdElem = document.getElementById('commentPostId');
        const postId = postIdElem.getAttribute('data-post-id');

        const commentSortElem = document.getElementById('comment-sorting')

        const comments = await fetchComments(5, window.nextCommentPage, postId, commentSortElem.value);
        window.nextCommentPage++;
        return comments;
    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}















//########################### SUPER REPORT
//########################### SUPER REPORT
//########################### SUPER REPORT
//########################### SUPER REPORT
//########################### SUPER REPORT
//########################### SUPER REPORT


window.nextSuperReportPage = 1

async function fetchMoreSuperReports() {

    try {


        const superReports = await fetchSuperReports(5, window.nextSuperReportPage);
        window.nextSuperReportPage++;
        return superReports;
    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}



async function fetchSuperReports(count, page) {

    const activeFilter = document.querySelector('#filtering .active');
    const activeSort = document.querySelector('#sorting .active');

    const data = {
        count: count,
        page: page,
        filter_type: 'all',
        sort_type: 'new',
        only_super_reports: true,
    };

    const responseElement = document.createElement('div');

    try {
        await sendJsonRequest("/postlist", data, responseElement, false);

        const jsonData = JSON.parse(responseElement.textContent);

        if (!jsonData?.data || !Array.isArray(jsonData.data)) {
            throw new Error("Invalid superreports data received");
        }

        return jsonData.data; // <-- This now returns ACTUAL data
    } catch (err) {
        console.error("Error:", err);
        throw err; // Re-throw to let caller handle it
    }
}








//##################################### reported posts
//##################################### reported posts
//##################################### reported posts
//##################################### reported posts
//##################################### reported posts
//##################################### reported posts
//##################################### reported posts
//##################################### reported posts
//##################################### reported posts




window.nextReportedPostPage = 1

async function fetchMoreReportedPosts() {

    try {


        const ReportedPosts = await fetchReportedPosts(5, window.nextReportedPostPage);
        window.nextReportedPostPage++;
        return ReportedPosts;
    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}



async function fetchReportedPosts(count, page) {

    const activeFilter = document.querySelector('#filtering .active');
    const activeSort = document.querySelector('#sorting .active');

    const data = {
        count: count,
        page: page,
        filter_type: 'all',
        sort_type: 'new',
        only_reported_posts: true,
    };

    const responseElement = document.createElement('div');

    try {

        const jsonData = JSON.parse(responseElement.textContent);

        if (!jsonData?.data || !Array.isArray(jsonData.data)) {
            throw new Error("Invalid ReportedPosts data received");
        }

        return jsonData.data; // <-- This now returns ACTUAL data
    } catch (err) {
        console.error("Error:", err);
        throw err; // Re-throw to let caller handle it
    }
}



//################################### REMOVED POSTS
//################################### REMOVED POSTS
//################################### REMOVED POSTS
//################################### REMOVED POSTS
//################################### REMOVED POSTS
//################################### REMOVED POSTS
//################################### REMOVED POSTS
//################################### REMOVED POSTS





window.nextRemovedPostPage = 1

async function fetchMoreRemovedPosts() {

    try {


        const RemovedPosts = await fetchRemovedPosts(5, window.nextRemovedPostPage);
        window.nextRemovedPostPage++;
        return RemovedPosts;
    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}



async function fetchRemovedPosts(count, page) {

    const activeFilter = document.querySelector('#filtering .active');
    const activeSort = document.querySelector('#sorting .active');

    const data = {
        count: count,
        page: page,
        filter_type: 'all',
        sort_type: 'new',
        only_removed_posts: true,
    };

    const responseElement = document.createElement('div');

    try {
        await sendJsonRequest("/postlist", data, responseElement, false);

        const jsonData = JSON.parse(responseElement.textContent);

        if (!jsonData?.data || !Array.isArray(jsonData.data)) {
            throw new Error("Invalid RemovedPosts data received");
        }

        return jsonData.data; // <-- This now returns ACTUAL data
    } catch (err) {
        console.error("Error:", err);
        throw err; // Re-throw to let caller handle it
    }
}









//################################# NOTIFICATIONS
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################



async function fetchNotifs(count, page) {

    const data = {
        count: count,
        page: page,
    };

    const responseElement = document.createElement('div');

    try {
        await sendJsonRequest("/notificationlist", data, responseElement, false)
        return JSON.parse(responseElement.textContent).data.Notifications
    } catch (err) {
        console.error("Error:", err);
        throw err; // Re-throw to let caller handle it
    }

}


window.nextNotifPage = 1

async function fetchMoreNotifs() {

    try {
        const notifs = await fetchNotifs(15, window.nextNotifPage);
        window.nextNotifPage++;
        return notifs;

    } catch (error) {
        console.error("Failed to fetch more posts:", error);
        throw error; // Re-throw to let caller handle it
    }
}


function createNotificationElement(notification) {
    const notificationDiv = document.createElement('div');
    notificationDiv.className = 'notification';

    // Main content paragraph
    const contentP = document.createElement('p');

    // Sender username as a link
    const senderLink = document.createElement('a');
    if (notification.SenderUserName != "System Administration") {
        senderLink.href = `/profile/${notification.SenderUserName}`;
        senderLink.style.color = 'blue';
    } else {
        senderLink.style.color = 'red';
    }
    senderLink.textContent = notification.SenderUserName;
    senderLink.style.textDecoration = 'none';

    contentP.appendChild(senderLink);
    contentP.appendChild(document.createTextNode(' '));

    // Sender username (always present)
    // const senderStrong = document.createElement('strong');
    // senderStrong.textContent = notification.SenderUserName;
    // contentP.appendChild(senderStrong);
    // contentP.appendChild(document.createTextNode(' '));

    // Add data attributes for potential interaction
    notificationDiv.dataset.notificationId = notification.Id;
    notificationDiv.dataset.targetId = notification.TargetId;



    // Handle all possible action types
    let actionText = '';
    switch (notification.ActionType) {
        case 'like':
            actionText = 'liked your ';
            break;
        case 'dislike':
            actionText = 'disliked your ';
            break;
        case 'comment':
            actionText = 'commented on your ';
            break;
        case 'mod action':
            actionText = 'performed a moderation action on your ';
            break;
        case 'mod request':
            actionText = 'requested moderation on your ';
            break;
        case 'ban':
            actionText = "banned you"
            break
        case 'unban':
            actionText = "unbanned you"
            break
        case 'modrequest':
            actionText = 'requested to become a moderator'
            break
        case 'mod-demotion':
            actionText = 'has demoted you to user'
            break
        case 'user-promotion':
            actionText = 'has promoted you to moderator'
            break
        case 'super-report':
            actionText = 'created a super report titled '
            break
        default:
            actionText = 'interacted with your ';
    }
    contentP.appendChild(document.createTextNode(actionText));

    // Handle all possible target types with appropriate wording
    console.log(notification.TargetType)
    switch (notification.TargetType) {
        case 'super-report':
            const bonusLink = document.createElement('a');
            bonusLink.className = 'bonus-text';
            const formattedBonusText = dh(notification.BonusText.replace(/\s+/g, '_'));
            bonusLink.textContent = dh(notification.BonusText);
            if (notification.TargetParentId === null) {
                if (notification.BonusText) {
                    bonusLink.href = `/superreport/${notification.TargetId}/${formattedBonusText}`;
                }
            } else {
                contentP.appendChild(document.createTextNode(' super report titled '))
                bonusLink.href = `/superreport/${notification.TargetParentId}/${formattedBonusText}`;
            }
            contentP.appendChild(bonusLink);
            break;
        case 'post':
            if (notification.BonusText) {
                contentP.appendChild(document.createTextNode('post titled '));

                const bonusLink = document.createElement('a');
                bonusLink.className = 'bonus-text';

                // Replace spaces with underscores
                const formattedBonusText = dh(notification.BonusText.replace(/\s+/g, '_'));
                bonusLink.href = `/post/${notification.TargetId}/${formattedBonusText}`;
                if (notification.ActionType === 'comment') {
                    bonusLink.href = `/post/${notification.TargetParentId}/${formattedBonusText}`;
                }
                bonusLink.textContent = dh(notification.BonusText);

                contentP.appendChild(bonusLink);
            } else {
                contentP.appendChild(document.createTextNode('post'));
            }
            break;

        case 'comment':
            if (notification.ActionType === 'comment') {
                contentP.appendChild(document.createTextNode('comment thread'));
            } else {

                contentP.appendChild(document.createTextNode('comment on post titled '));

                const commentPostLink = document.createElement('a');
                commentPostLink.className = 'bonus-text';

                // This should now be the actual post title, not the comment
                const formattedPostTitle = dh(notification.BonusText)
                    ? dh(notification.BonusText.replace(/\s+/g, '_'))
                    : 'view';

                // Use post ID (TargetId) and actual title        
                commentPostLink.href = `/post/${notification.TargetParentId}/${formattedPostTitle}`;
                commentPostLink.textContent = dh(notification.BonusText) || 'View post';


                console.log(formattedPostTitle)


                contentP.appendChild(commentPostLink);

            }
            break;

        case 'user':
            contentP.appendChild(document.createTextNode('profile'));
            break;

        default:
            contentP.appendChild(document.createTextNode('.'));
    }

    notificationDiv.appendChild(contentP);

    // Timestamp (simple ISO format display)
    const timestampP = document.createElement('p');
    const timestampSmall = document.createElement('small');
    const timestampSpan = document.createElement('span');
    timestampSpan.className = 'timestamp';

    try {
        const date = new Date(notification.Created);
        timestampSpan.textContent = date.toLocaleString();
    } catch (e) {
        timestampSpan.textContent = 'just now';
    }

    timestampSmall.appendChild(timestampSpan);
    timestampP.appendChild(timestampSmall);
    notificationDiv.appendChild(timestampP);

    // "New" badge for unseen notifications
    if (!notification.Seen) {
        const newSpan = document.createElement('span');
        newSpan.className = 'new-notification';
        newSpan.textContent = 'New';
        notificationDiv.appendChild(newSpan);
    }

    // Additional bonus text (for comments or mod actions)
    if (notification.BonusText &&

        notification.ActionType === 'ban' ||
        notification.ActionType === 'unban') {
        const bonusP = document.createElement('p');
        bonusP.className = 'bonus-text';
        bonusP.textContent = dh(notification.BonusText);
        notificationDiv.appendChild(bonusP);
    }



    if (notification.TargetParentId) {
        notificationDiv.dataset.targetParentId = notification.TargetParentId;
    }

    return notificationDiv;
}


//################################# USER
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################
//#############################################


function createUserElement(user) {
    const container = document.createElement('div');
    container.style.border = '1px solid #ccc';
    container.style.padding = '1em';
    container.style.margin = '1em 0';
    container.style.borderRadius = '8px';
    container.style.maxWidth = '400px';
    container.style.fontFamily = 'sans-serif';
    container.classList.add('post-box')
    // container.classList.add("fa-solid")

    const addField = (label, value) => {
        const row = document.createElement('div');
        row.style.marginBottom = '0.5em';
        const labelEl = document.createElement('strong');
        labelEl.textContent = `${label}: `;
        row.appendChild(labelEl);
        row.appendChild(document.createTextNode(value ?? 'None'));
        container.appendChild(row);
    };

    addField('ID', user.id);
    addField('Username', user.user_name);
    addField('Email', user.email);
    addField('Profile Pic', user.profile_pic);
    addField('Description', user.description);
    addField('Bio', user.bio);
    addField('Total Karma', user.total_karma);
    addField('Created', new Date(user.created).toLocaleString());
    addField('Role', user.role);

    return container;
}


