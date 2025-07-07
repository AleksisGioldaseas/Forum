document.addEventListener('DOMContentLoaded', setupCommentEventDelegation);

function setupCommentEventDelegation() {
    // Use the closest stable parent container that holds all comments
    const commentsContainer = document.querySelector('.comments-container') || document.body;
    commentsContainer.addEventListener('click', (event) => {
        
        const target = event.target;
        
        const commentWrapper = target.closest('.outer-comment-wrapper');
        if (!commentWrapper){
            console.log("unrecognized button"); 
            return; // Exit if click wasn't in a comment
        }

        
        // Report Comment Button
        else if (target.matches('.report-comment-btn') || target.closest('.report-comment-btn')) {
           
            commentWrapper.querySelector('.report-controls').style.display = 'block';
        }
        // Send Report Comment Button
        else if (target.matches('.sendButtonReportComment') || target.closest('.sendButtonReportComment')) {
            
            const clickedButton = target.matches('.sendButtonReportComment') ? 
                target : target.closest('.sendButtonReportComment');
            const container = clickedButton.closest('.report-controls');
            const newBody = container.querySelector('.newText').value;
            

            const data = {
            comment_id: parseInt(clickedButton.dataset.value),
                message: newBody
            };
            commentWrapper.querySelector('.report-controls').style.display = 'none';

            sendJsonRequest("/commentreport", data, document.getElementById("editCommentResponse"), false); 
        }

        else if (target.matches('.com-vote-up')) {
            const isLoggedIn = document.getElementById('auth-status').getAttribute('data-logged-in') === 'true';
        

            // Check if the click was on a like/dislike button
            const likeButton = target.closest('.like-button');

            if ((likeButton) && !isLoggedIn) {
                const authButton = document.getElementById("auth-btn");
                if (authButton) {
                    authButton.click(); // Simulate clicking the login button
                }
                return;
            }

            if (likeButton) {
                
                handleReactionClick(likeButton, 'like', false);
            }
            
        }

        else if (target.matches('.com-vote-down')) {
            const isLoggedIn = document.getElementById('auth-status').getAttribute('data-logged-in') === 'true';
          

            // Check if the click was on a like/dislike button
         
            const dislikeButton = target.closest('.dislike-button');

            if ((dislikeButton) && !isLoggedIn) {
                const authButton = document.getElementById("auth-btn");
                if (authButton) {
                    authButton.click(); // Simulate clicking the login button
                }
                return;
            }

            if (dislikeButton) {
                handleReactionClick(dislikeButton, 'dislike', false);
            }
        }


        // Edit Comment Button
        else if (target.matches('.edit-comment-btn') || target.closest('.edit-comment-btn')) {
            commentWrapper.querySelector('.edit-controls').style.display = 'block';
            commentWrapper.querySelector('.comment-body').classList.add('hidden');
        }
        // Send Edit Comment Button
        else if (target.matches('.sendButtonEditComment') || target.closest('.sendButtonEditComment')) {
            const clickedButton = target.matches('.sendButtonEditComment') ? 
                target : target.closest('.sendButtonEditComment');
            const container = clickedButton.closest('.edit-controls');
            const newBody = container.querySelector('.newText').value;

            const data = {
                comment_id: parseInt(clickedButton.dataset.value),
                body: newBody
            };
            sendJsonRequest("/commentedit", data, document.getElementById("editCommentResponse"), true);
        }



        // Delete Comment Button
        else if (target.matches('.delete-comment-btn') || target.closest('.delete-comment-btn')) {
        const confirmed = confirm("Are you sure you want to delete comment?");
        if (!confirmed) return;
            const clickedButton = target.matches('.delete-comment-btn') ? 
                target : target.closest('.delete-comment-btn');
            const data = {
                comment_id: parseInt(clickedButton.dataset.value)
            };
            sendJsonRequest("/commentdelete", data, document.getElementById("deleteCommentResponse"), true);
        }



        // Approve Comment Button (Moderator)
        else if (target.matches('.approve-comment-btn') || target.closest('.approve-comment-btn')) {
            const clickedButton = target.matches('.approve-comment-btn') ? 
                target : target.closest('.approve-comment-btn');
            const confirmed = confirm("Are you sure you want to approve this comment?");
            if (!confirmed) return;
            const data = {
                comment_id: parseInt(clickedButton.dataset.value)
            };
            sendJsonRequest("/commentapprove", data, document.getElementById("approveCommentResponse"), true);
        }


        
        // Remove Comment Button
        if (target.matches('.remove-comment-btn') || target.closest('.remove-comment-btn')) {
            const removalReasons = commentWrapper.querySelector('.removal-reasons');
            removalReasons.style.display = "block";
        }
        // Final Remove Comment Button (Moderator)
        else if (target.matches('.final-remove-comment-btn') || target.closest('.final-remove-comment-btn')) {
            const clickedButton = target.matches('.final-remove-comment-btn') ? 
                target : target.closest('.final-remove-comment-btn');
            
            const data = {
                comment_id: parseInt(clickedButton.dataset.value),
                reason: clickedButton.textContent
            };
            sendJsonRequest("/commentremove", data, document.getElementById("removeCommentResponse"), true);
        }else{
            console.log("unrecognized button 2")
        }
    });
}
