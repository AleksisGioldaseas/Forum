package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func CreateSuperReportHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("EVENT Create super report handler called!")

	type SuperReportRequest struct {
		Title                string `json:"title"`
		Body                 string `json:"body"`
		SuperReportCommentId int64  `json:"super_report_comment_id"`
		SuperReportPostId    int64  `json:"super_report_post_id"`
		SuperReportUserId    int64  `json:"super_report_user_id"`
	}

	superReportReq := &SuperReportRequest{}
	err := jsonRequestExtractor(request, superReportReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "create super report: "+err.Error())
		return
	}

	sanitizedTitle := database.Sanitize(superReportReq.Title)
	sanitizedBody := string(database.BasicHTMLSanitize(superReportReq.Body))

	superReport := &config.Post{
		UserID:               activeUser.ID, //we add it ourselves because we can't trust the user!
		Title:                sanitizedTitle,
		IsSuperReport:        true,
		Body:                 sanitizedBody,
		SuperReportCommentId: superReportReq.SuperReportCommentId,
		SuperReportPostId:    superReportReq.SuperReportPostId,
		SuperReportUserId:    superReportReq.SuperReportUserId,
		Categories:           []string{"cats"},
	}

	err = db.ValidateTitle(superReport.Title)

	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, err.Error(), "create super report: "+err.Error())
		return
	}

	err = db.ValidatePostBody(superReport.Body)

	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, err.Error(), "create super report: "+err.Error())
		return
	}

	superReportId, err := db.AddPost(superReport)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "CreateSuperReport: Post creation failed. "+err.Error())
		return
	}

	data := struct { //we send the id so website can redirect to the super report id's url
		SuperReportId int64 `json:"super_report_id"`
	}{
		SuperReportId: superReportId,
	}

	err = jsonOkResponder(writer, data, "Super report received")
	if err != nil {
		fmt.Println("CreateSuperReport: Failed to send final response!")
	}

}
