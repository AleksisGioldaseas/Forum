package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func ApprovePostHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("approvePost Handler called!")

	type approvePostRequest struct {
		PostId int64 `json:"post_id"`
	}

	approvePostReq := &approvePostRequest{}
	err := jsonRequestExtractor(request, approvePostReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "approve post: "+err.Error())
		return
	}

	err = db.ModAssessment("post", approvePostReq.PostId, activeUser.ID, "", activeUser.UserName, 0)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to approve post", "failed to approve post. "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("ApprovePostHandler: Failed to send final response!")
	}

}
