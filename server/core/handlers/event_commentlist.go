package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func GetCommentListHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("CommentList Handler called!")

	type CommentListRequest struct {
		PostId       int    `json:"post_id"`
		Count        int    `json:"count"`
		Page         int    `json:"page"` //page 1, 2, 3 and so on
		Sorttype     string `json:"sort_type"`
		ReportedOnly bool   `json:"reported_only"`
		RemovedOnly  bool   `json:"removed_only"`
	}

	commentsRequest := &CommentListRequest{}
	err := jsonRequestExtractor(request, commentsRequest)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "get commetn list: "+err.Error())
		return
	}

	commentsStruct := struct { //commentStruct
		Comments []*config.Comment `json:"comments"`
	}{}

	var errs []error
	commentsStruct.Comments, errs = db.GetComments(nil, int(commentsRequest.PostId), commentsRequest.Count, commentsRequest.Page, activeUser.ID, activeUser.Role, commentsRequest.Sorttype, commentsRequest.ReportedOnly, commentsRequest.RemovedOnly)
	if len(errs) > 0 {
		fmt.Println("errors fetching comments by post id: ", errs)
		showErrorPage(writer, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	fmt.Println("CommentListRequest:", commentsRequest)
	fmt.Println("comments sent: ", len(commentsStruct.Comments))

	err = jsonOkResponder(writer, commentsStruct)
	if err != nil {
		fmt.Println("CommentList: Failed to send final response!")
	}

}
