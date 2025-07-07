package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func ApproveCommentHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("approveComment Handler called!")

	type approveCommentRequest struct {
		CommentId int64 `json:"comment_id"`
	}

	approveCommentReq := &approveCommentRequest{}
	err := jsonRequestExtractor(request, approveCommentReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "approve comment: "+err.Error())
		return
	}

	err = db.ModAssessment("comment", approveCommentReq.CommentId, activeUser.ID, "", activeUser.UserName, 0)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to approve comment", "Failed to approve comment: "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("ApproveCommentHandler: Failed to send final response!")
	}

}
