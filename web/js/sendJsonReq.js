// Vag: Modified this to send raw json and multipart json requests
// Original func bellow
async function sendJsonRequest(url, data, responseElement, refresh) {
    try {
        const isFormData = data instanceof FormData;
        const response = await fetch(url, {
            method: "POST",
            headers: isFormData ? {} : { "Content-Type": "application/json" },
            body: isFormData ? data : JSON.stringify(data)
        });
        const requestBody = isFormData ? data : JSON.stringify(data);


        if (response.ok) {
            const responseData = await response.json(); // Assuming the server returns JSON
            if (responseElement) {
                responseElement.textContent = `${JSON.stringify(responseData)}`;
            }

            if (responseData.message != "") {
                showSuccPopup(`${responseData.message}`)
            }

            if (refresh == true) {
                location.reload()
            }
            if (url == "/createpost"){
                // console.log(responseData.data.post_id)
                window.location.assign(`/post/${responseData.data.post_id}`);
            }
            
        } else {
            const responseData = await response.json();
            // if (responseElement) {
            //     responseElement.textContent = `Error: ${response.status} - ${response.statusText} - ${JSON.stringify(responseData)}`;
            // }

            showErrorPopup(`${responseData.message}`)
            console.log(responseData.console_log)
            
            
        }
        
    } catch (error) {
  
        showErrorPopup(`Sorry something unexpected happen!`)
        console.log("Unexpcted error: ", error)
    }

    
}

function showErrorPopup(message, duration = 4000) {
    // Get the template
    const template = document.getElementById('error-popup-template');

    // Clone the template content
    const popup = template.content.cloneNode(true).querySelector('.error-popup');

    // Set the message text
    popup.querySelector('.error-popup-text').textContent = message;



    // Append to body
    document.body.appendChild(popup);

    // Trigger the show animation
    setTimeout(() => {
        popup.classList.add('show');
    }, 10);



    // Start fade out and remove after duration
    setTimeout(() => {
        popup.classList.add('fade-out');

        // Remove from DOM after animation completes
        setTimeout(() => {
            popup.remove();
        }, 500);
    }, duration);
}

function showSuccPopup(message, duration = 3000) {
    // Get the template
    const template = document.getElementById('succ-popup-template');

    // Clone the template content
    const popup = template.content.cloneNode(true).querySelector('.succ-popup');

    // Set the message text
    popup.querySelector('.succ-popup-text').textContent = message;

  

    // Append to body
    document.body.appendChild(popup);

    // Trigger the show animation
    setTimeout(() => {
        popup.classList.add('show');
    }, 10);



    // Start fade out and remove after duration
    setTimeout(() => {
        popup.classList.add('fade-out');

        // Remove from DOM after animation completes
        setTimeout(() => {
            popup.remove();
        }, 500);
    }, duration);
}

function addEventListenerIfExists(elementId, eventType, callback) {
    const element = document.getElementById(elementId);
    if (element) {
        element.addEventListener(eventType, (event) => {
            callback(event); // Vag: returning callback event to the callerhere
        });
    } else {
        //IF BY ID IT FAILS, THEN TRY CLASS NAME

        // Select all elements with class
        const elements = document.querySelectorAll(elementId);

        // Add event listener to each
        elements.forEach(element => {
            element.addEventListener(eventType, (event) => {
                callback(event); // Vag: returning callback event to the callerhere
            });
        });


     
    }
}





// voting testing
function handleVote(button, action, ispost = true) {

    let wrapper = button.closest(".post-wrapper");
    
    if (!wrapper) {
        wrapper = button.closest(".outer-comment-wrapper")
        if (!wrapper) {
            console.error("wrapper not found!");
            return;
        }
    }
    let Id = 0
    if (ispost) {
        Id = parseInt(wrapper.getAttribute("data-post-id"));
        if (Id == 0) {
            return;
        }
    } else {
        Id = parseInt(wrapper.getAttribute("data-comment-id"));
        if (Id == 0) {
            return;
        }
    }

    // Get the current user reaction from the data attribute
    const activeUserReaction = parseInt(wrapper.getAttribute("data-user-reaction"));
    let newAction = action;
    // Determine the new action: like, dislike, or neutral (removing vote)
    if ((newAction === "like" && activeUserReaction === 1) || (newAction === "dislike" && activeUserReaction === -1)) {

        newAction = "neutral"; // Reset vote if the same button is clicked
    }
    // Update the data-user-reaction after changing the reaction
    wrapper.setAttribute('data-user-reaction', newAction === "like" ? 1 : (newAction === "dislike" ? -1 : 0));


    const data = {
        id: Id,
        action: newAction
    };

    if (ispost) {
        sendJsonRequest("/votepost", data, null, false);
    } else {
        sendJsonRequest("/votecomment", data, null, false);
    }
}

function hookAllButtons() {

    addEventListenerIfExists("sendButtonLogin", "click", function () {
        const data = {
            username: document.getElementById("loginUsername").value,
            password: document.getElementById("loginPassword").value
        };
        if (data.username == "" || data.password == ""){
            alert("Please fill missing fields")
            return
        }
        sendJsonRequest("/login", data, document.getElementById("loginResponse"), true);
    });

    // Register
    addEventListenerIfExists("sendButtonRegister", "click", function () {
        const data = {
            username: document.getElementById("registerUsername").value,
            password: document.getElementById("registerPassword").value,
            passwordRepeat: document.getElementById("registerPasswordRepeat").value,
            email: document.getElementById("registerEmail").value
        };

        if (data.username == "" || data.password == "" || data.passwordRepeat == "" || data.email == ""){
            alert("Please fill missing fields")
            return
        }

        sendJsonRequest("/signup", data, document.getElementById("registerResponse"), false);
        // quoting the master
    });

    // Logout
    addEventListenerIfExists("sendButtonLogout", "click", function () {
        const confirmed = confirm("Are you sure you want to log out?");
        if (!confirmed) return;
        sendJsonRequest("/logout", {}, document.getElementById("logoutResponse"), true);
    });
    // Logout
    addEventListenerIfExists("sendButtonLogoutAllElse", "click", function () {
        sendJsonRequest("/logoutall", {}, document.getElementById("logoutResponse"), true);
    });



    //Create Super Report
    addEventListenerIfExists("sendButtonCreateSuperReport", "click", function () {

        const data = {
            title: document.getElementById("postTitle").value,
            body: document.getElementById("postBody").value,
            super_report_comment_id: parseInt(document.getElementById("comment-id").dataset.value),
            super_report_post_id: parseInt(document.getElementById("post-id").dataset.value),
            super_report_user_id: parseInt(document.getElementById("user-id").dataset.value)
        };
        sendJsonRequest("/createsuperreport", data, document.getElementById("createPostResponse"), false);
    });


    // Create Post Multipart parsing
    addEventListenerIfExists("sendButtonCreatePost", "click", (event) => {
        event.preventDefault();

        const MAX_IMAGE_SIZE = 20 * 1024 * 1024;
        const formData = new FormData();

        const title = document.getElementById("postTitle").value.trim();
        const body = document.getElementById("postBody").value.trim();
        const imageInput = document.getElementById("postImage");
        const image = imageInput.files[0];

        if (!title && !body && !image) {
            alert("Please provide at least a title, a body, or an image before submitting.");
            return;
        }

        formData.append("title", title);
        formData.append("body", body);

        if (image) {
            if (image.size > MAX_IMAGE_SIZE) {
                alert("Image is too large. Maximum size is 20MB.");
                return;
            }
            formData.append("image", image);
        } else {
            console.warn("No image selected. Submitting without image.");
        }

        const categoriesInput = document.getElementById("postCategories").value;
        const categories = categoriesInput
            .split(",")
            .map(c => c.trim())
            .filter(c => c !== "");

        for (const category of categories) {
            formData.append("categories", category);
        }

     

        sendJsonRequest("/createpost", formData, document.getElementById("createPostResponse"), false);

        

    });

    // Create Comment
    addEventListenerIfExists("sendButtonCreateComment", "click", function () {
        const data = {
            post_id: parseInt(document.getElementById("commentPostId").getAttribute("data-post-id")),
            body: document.getElementById("commentBody").value
        };
        sendJsonRequest("/createcomment", data, document.getElementById("createCommentResponse"), true);
    });


    // UNVERIFIED Edit Post
    addEventListenerIfExists("sendButtonEditpost", "click", function (event) {
        const clickedButton = event.currentTarget;
        const data = {
            post_id: parseInt(clickedButton.dataset.value),
            body: document.getElementById("postNewText").value
        };
        sendJsonRequest("/postedit", data, document.getElementById("editPostResponse"), true);
    });

    document.querySelectorAll('.edit-post-btn').forEach(btn => {
        btn.addEventListener('click', () => {

            const body = document.getElementById('post-body')
            body.classList.add('hidden');

            const postEditDiv = document.getElementById('post-edit-things')
            postEditDiv.style.display = 'block';

            const newText = document.getElementById('postNewText')
            newText.value = body.innerHTML
        });
    });




    //  Delete Post
    addEventListenerIfExists(".delete-post-btn", "click", function (event) {
        const clickedButton = event.currentTarget;
        const confirmed = confirm("Are you sure you want to delete this post?");
        if (!confirmed) return;
        const data = {
            post_id: parseInt(clickedButton.dataset.value)
        };
        sendJsonRequest("/postdelete", data, document.getElementById("deletePostResponse"), true);
    });





    // UNVERIFIED Report Post
    addEventListenerIfExists(".sendButtonReportPost", "click", function (event) {

        const clickedButton = event.currentTarget;
        const data = {
            post_id: parseInt(clickedButton.dataset.value),
            message: document.getElementById("postReportText").value
        };
        sendJsonRequest("/postreport", data, document.getElementById("reportPostResponse"), false);
    });



    //  Approve Post (Moderator)
    addEventListenerIfExists(".approve-post-btn", "click", function (event) {
        const clickedButton = event.currentTarget;
        const confirmed = confirm("Are you sure you want to approve this post?");
        if (!confirmed) return;
        const data = {
            post_id: parseInt(clickedButton.dataset.value)
        };
        sendJsonRequest("/postapprove", data, document.getElementById("approvePostResponse"), true);
    });

    document.querySelectorAll('.remove-post-btn').forEach(btn => {
        btn.addEventListener('click', (event) => {
            const container = event.currentTarget.closest('.post-wrapper');
            const removalReasons = container.querySelector('.removal-reasons');
            removalReasons.style.display = "block";
        });
    });


    //  Remove Post (Moderator)
    addEventListenerIfExists(".final-remove-post-btn", "click", function (event) {
        const clickedButton = event.currentTarget;
        const data = {
            post_id: parseInt(clickedButton.dataset.value),
            reason: clickedButton.textContent
        };
        sendJsonRequest("/postremove", data, document.getElementById("removePostResponse"), true);
    });




    //  Mod request (User)
    addEventListenerIfExists("sendButtonModRequest", "click", function () {
        const confirmed = confirm("Are you sure you want to submit mod request?");
        if (!confirmed) return;
        const data = {
        }
        sendJsonRequest("/modrequest", data, document.getElementById("ModRequestResponse"), true);
    });

    //  Promote User
    addEventListenerIfExists("promoteUser", "click", function (event) {
        const clickedButton = event.currentTarget;
        const confirmed = confirm("Are you sure you want to promote this user?");
        if (!confirmed) return;
        const data = {
            username: clickedButton.dataset.value
        }
        sendJsonRequest("/promoteuser", data, document.getElementById("ModRequestResponse"), true);
    });

    //  Demote User
    addEventListenerIfExists("demoteMod", "click", function (event) {
        const clickedButton = event.currentTarget;
        const confirmed = confirm("Are you sure you want to demote this user?");
        if (!confirmed) return;
        const data = {
            username: clickedButton.dataset.value
        }
        sendJsonRequest("/demotemoderator", data, document.getElementById("ModRequestResponse"), true);
    });

    // Reload Comments
    addEventListenerIfExists("comment-sorting", "change", function (event) {
        const element = event.currentTarget;

        window.stopLoadingComments = false
        window.nextCommentPage = 1

        const data = {
            post_id: parseInt(element.dataset.postid),
            count: 5,
            page: 0,
            sort_type: element.value

        }

        const responseElement = document.createElement('div');

        sendJsonRequest("/commentlist", data, responseElement, false)
            .then(() => {
                refreshComments(JSON.parse(responseElement.textContent).data.comments);
                // hookAllButtons()
            });
    });

    // Add New Category (Admin)
    addEventListenerIfExists("addCategoryButton", "click", function () {
        const categoryName = document.getElementById("add-category-box").value.trim();
        if (!categoryName) return;
        const confirmed = confirm("Are you sure you want to add this category?");
        if (!confirmed) return;

        const data = { name: categoryName };
        const responseElement = document.getElementById("createPostResponse");

        sendJsonRequest("/addcategory", data, responseElement, true);
    });


    // Remove Category (Admin)
    addEventListenerIfExists("removeCategoryButton", "click", function () {
        const categoryName = document.getElementById("remove-category-box").value.trim();
        if (!categoryName) return;
        const confirmed = confirm("Are you sure you want to remove this category?");
        if (!confirmed) return;

        const data = { name: categoryName };
        const responseElement = document.getElementById("createPostResponse");

        sendJsonRequest("/removecategory", data, responseElement, true);
    });


    addEventListenerIfExists("confirmBanBtn", "click", function (event) {
        // const clickedButton = event.currentTarget;
        const username = document.getElementById("banButton")?.dataset.value;

        // Get selected days
        let days = null;
        const selectedBanOption = document.querySelector(".ban-option.selected");
        if (selectedBanOption) {
            const dayValue = selectedBanOption.dataset.days;
            if (dayValue === "custom") {
                const customVal = parseInt(document.getElementById("customBanInput").value, 10);
                if (!isNaN(customVal) && customVal > 0) {
                    days = customVal;
                }
            } else if (dayValue === "0") {
                days = 99999; // permanent
            } else {
                days = parseInt(dayValue, 10);
            }
        }

        // Get selected reason
        const selectedReasonButton = document.querySelector(".ban-reason.selected");
        const reason = selectedReasonButton ? selectedReasonButton.dataset.reason : null;

        if (!username || !days || !reason) {
            console.warn("Missing required ban data");
            return;
        }

        const data = {
            username: username,
            days: days,
            reason: reason
        };

        const responseElement = document.getElementById("createPostResponse");
        sendJsonRequest("/banuser", data, responseElement, true);
    });

    addEventListenerIfExists("unbanButton", "click", function (event) {
        const clickedButton = event.currentTarget;
        const confirmed = confirm("Are you sure you want to unban this user?");
        if (!confirmed) return;
        const data = {
            username: clickedButton.dataset.value,
        }
        const responseElement = document.getElementById("createPostResponse");
        sendJsonRequest("/unbanuser", data, responseElement, true);
    })

    // Update Bio Button
    addEventListenerIfExists("updateBioButton", "click", function () {
        
        const bioEditArea = document.getElementById("bio_edit_area");
        const currentBio = document.getElementById("current-bio");
        const updateBioButton = document.getElementById("updateBioButton");
        const responseElement = document.createElement('div');


        const raw = currentBio.innerHTML 
        currentBio.innerHTML = ''
        currentBio.innerText = raw
        // Avoid adding buttons again
        if (bioEditArea.innerHTML.trim() !== "") return;

        // Make paragraph editable
        currentBio.contentEditable = "true";
        currentBio.focus();
        currentBio.style.border = "1px solid #ccc";
        currentBio.style.padding = "5px";

        // Hide Update Bio button
        updateBioButton.style.display = "none";

        // Create Save button
        const saveButton = document.createElement("button");
        saveButton.innerText = "Save";
        saveButton.onclick = function () {
            const bio = currentBio.innerText;
            const data = {
                bio: bio
            };

            sendJsonRequest("/updatebio", data, responseElement, true);

            // Reset UI
            currentBio.contentEditable = "false";
            currentBio.style.border = "none";
            updateBioButton.style.display = "inline-block";
            bioEditArea.innerHTML = "";
        };

        // Create Cancel button
        const cancelButton = document.createElement("button");
        cancelButton.innerText = "Cancel";
        cancelButton.style.marginLeft = "10px";
        cancelButton.onclick = function () {
            // Just cancel editing, no data sent
            currentBio.contentEditable = "false";
            currentBio.style.border = "none";
            updateBioButton.style.display = "inline-block";
            bioEditArea.innerHTML = "";
        };

        // Append buttons
        bioEditArea.appendChild(saveButton);
        bioEditArea.appendChild(cancelButton);
    });





}



function searchButtonHandler() {
    fetchPosts(5, 0).then(posts => refreshPosts(posts))
}

function loadFirstSuperReports() {
    fetchPosts(5, 0).then(posts => refreshPosts(posts))
}


window.searchButtonHandler = searchButtonHandler;
// Attach the event listener to the search button
addEventListenerIfExists("search-button", "click", searchButtonHandler);


function refreshPosts(posts) {
    let postContainer = document.getElementById("posts-container");
    
    postContainer.innerHTML = ''
    window.stopLoading = false
    window.nextPage = 1
    window.scrollTo(0, 0);
    posts.forEach(post => {
        let p = renderMiniPost(post)
        postContainer.appendChild(p)
    });
}

function appendPosts(posts) {
    let postContainer = document.getElementById("posts-container");
    posts.forEach(post => {
        p = renderMiniPost(post)
        postContainer.appendChild(p)
    });
}

function appendComments(comments, activeUserRole, activeUserName, isSuper = false) {
    let commentsSection = document.getElementById("comments-section");


    comments.forEach(comment => {
        c = createCommentElement(comment, activeUserRole, activeUserName, isSuper)
        commentsSection.appendChild(c)
    });


}


document.addEventListener("DOMContentLoaded", function () {
    hookAllButtons()
});

// Remove hover functionality
const postContainer = document.querySelector('.post-container');
let canEdit = false 
if (postContainer){
    canEdit = postContainer.getAttribute('data-can-edit-profile') === 'true';
}

// Add Profile image
if (canEdit) {
    addEventListenerIfExists("profileImage", "click", () => {
        const input = document.createElement("input");
        input.type = "file";
        input.accept = "image/*";

        input.onchange = () => {
            const file = input.files[0];

            if (!file) return; // User cancelled

            if (file.size > 20 * 1024 * 1024) { // 20MB
                alert("File size must be 20MB or less.");
                return;
            }

            const reader = new FileReader();
            reader.onload = () => {
                const img = document.getElementById("profileImage");
                img.src = reader.result;
                img.style.display = "block";

                document.getElementById("actionButtons").style.display = "block";
            };
            reader.readAsDataURL(file);

            // Save the file reference if you need to upload it later
            window.selectedProfileImageFile = file;
        };

        input.click(); // Trigger file picker
    });
}

document.getElementById("cancelUploadButton")?.addEventListener("click", () => {
    document.getElementById("profileImage").style.display = "none";
    document.getElementById("profileImage").src = "";
    document.getElementById("actionButtons").style.display = "none";
    window.selectedProfileImageFile = null;
    location.reload();
});

document.getElementById("uploadProfilePicButton")?.addEventListener("click", () => {
    if (window.selectedProfileImageFile) {
        const MAX_IMAGE_SIZE = 20 * 1024 * 1024;
        const formData = new FormData();
        formData.append("image", window.selectedProfileImageFile);
        sendJsonRequest("/profilepic", formData, document.getElementById("createImageResponse"), true);
    } else {
        alert("No file selected.");
    }
});




