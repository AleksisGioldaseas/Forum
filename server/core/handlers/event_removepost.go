package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func RemovePostHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("removePost Handler called!")

	type removePostRequest struct {
		PostId int64  `json:"post_id"`
		Reason string `json:"reason"`
	}

	removePostReq := &removePostRequest{}
	err := jsonRequestExtractor(request, removePostReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "remove post: "+err.Error())
		return
	}
	// Vag: Assuming activeUser is the moderator and the moderator is verified
	err = db.ModAssessment("post", removePostReq.PostId, activeUser.ID, removePostReq.Reason, activeUser.UserName, 1)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to remove post", "failed to remove post. "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("RemovePostHandler: Failed to send final response!")
	}

}
