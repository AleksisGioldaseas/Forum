package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func ReportPostHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("reportPost Handler called!")

	type reportPostRequest struct {
		PostId  int64  `json:"post_id"`
		Message string `json:"message"`
	}

	reportPostReq := &reportPostRequest{}
	err := jsonRequestExtractor(request, reportPostReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "report post: "+err.Error())
		return
	}

	err = db.ValidateReport(reportPostReq.Message)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, err.Error(), "failed to report. "+err.Error())
		return
	}

	err = db.Report(activeUser.ID, reportPostReq.PostId, 0, reportPostReq.Message)

	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "failed to report post. "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data, "Your report was received!")
	if err != nil {
		fmt.Println("ReportPostHandler: Failed to send final response!")
	}
}
