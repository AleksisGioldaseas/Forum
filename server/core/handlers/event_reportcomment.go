package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func ReportCommentHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("reportComment Handler called!")

	type reportCommentRequest struct {
		CommentId int64  `json:"comment_id"`
		Message   string `json:"message"`
	}

	reportCommentReq := &reportCommentRequest{}
	err := jsonRequestExtractor(request, reportCommentReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "report comment: "+err.Error())
		return
	}

	sanitizedMessage := database.Sanitize(reportCommentReq.Message)

	err = db.ValidateReport(sanitizedMessage)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, err.Error(), "failed to report. "+err.Error())
		return
	}

	err = db.Report(activeUser.ID, 0, reportCommentReq.CommentId, sanitizedMessage)

	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "failed to report comment. "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data, "Your report was received!")
	if err != nil {
		fmt.Println("ReportCommentHandler: Failed to send final response!")
	}

}
