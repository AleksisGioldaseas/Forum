package handlers

import (
	"encoding/json"
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
)

func SuperReportPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Super report page handler called")
	activeUserIsLoggedIn := *activeUser != config.User{}

	temp := strings.ToLower(strings.TrimPrefix(request.URL.Path, "/superreport/"))
	temp = strings.TrimSuffix(temp, "/")
	parts := strings.Split(temp, "/")
	superReportIdstr := parts[0]

	superReportId, err := strconv.Atoi(superReportIdstr)
	if err != nil {
		fmt.Println("non number super report id given by user")
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusNotFound, "404 - Page Not Found")
		return
	}

	var Comment *config.Comment
	var Post *config.Post
	var User *config.User

	tx, _ := db.ExportTx()
	defer tx.Rollback()

	superReport, err := db.GetPostById(tx, activeUser.ID, int64(superReportId), activeUser.Role > 1) // added user id here to and changed postId to int64 to run db func
	if err != nil || !superReport.IsSuperReport {
		fmt.Println("unable to find post using id: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusNotFound, "404 - Page Not Found")
		return
	}

	switch {
	case superReport.SuperReportCommentId != 0:
		Comment, err = db.GetCommentById(tx, activeUser.Role, superReport.SuperReportCommentId)
	case superReport.SuperReportPostId != 0:
		Post, err = db.GetPostById(tx, activeUser.ID, superReport.SuperReportPostId, true)
	case superReport.SuperReportUserId != 0:
		User, err = db.GetUserById(tx, superReport.SuperReportUserId)
	}
	if err != nil {
		fmt.Println("Can't find report target: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusNotFound, "404 - Report Target Not Found")
		return
	}

	mainSuperReportJson, err := json.Marshal(superReport)
	if err != nil {
		fmt.Println("unable to marshan main superReport: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	commentsStruct := struct { //commentStruct
		Comments []*config.Comment `json:"comments"`
	}{}
	var errs []error
	commentsStruct.Comments, errs = db.GetComments(tx, int(superReport.ID), 20, 0, activeUser.ID, activeUser.Role, "old", false, false)
	if len(errs) > 0 {
		fmt.Println("errors fetching comments by super report id: ", errs)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	commentsJson, err := json.Marshal(commentsStruct)
	if err != nil {
		fmt.Println("unable to marshan main super report: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	hasUnseen, notificationCount, err := db.CountNotifications(nil, tx, int(activeUser.ID))
	if err != nil {
		fmt.Println("Error unable to count notifiations: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Print("SuperReportPageHandler: Transaction Failure", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	postJson, err := json.Marshal(Post)
	if err != nil {
		fmt.Println("unable to marshan target postJson: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}
	commentJson, err := json.Marshal(Comment)
	if err != nil {
		fmt.Println("unable to marshan target commentJson: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}
	userJson, err := json.Marshal(User)
	if err != nil {
		fmt.Println("unable to marshan target userJson: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type SuperReportDataT struct {
		//all pages
		GenericResponse

		//specific
		JsonComment string
		JsonPost    string
		JsonUser    string

		SuperReportId   int
		JsonSuperReport string
		SuperReport     *config.Post
		JsonComments    string
		Comments        []*config.Comment
	}

	data := SuperReportDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},

		JsonComment: string(commentJson),
		JsonPost:    string(postJson),
		JsonUser:    string(userJson),

		SuperReportId:   int(superReport.ID),
		JsonSuperReport: string(mainSuperReportJson),
		SuperReport:     superReport,
		JsonComments:    string(commentsJson),
		Comments:        commentsStruct.Comments,
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexSuperReport.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		// SendErrorPage(writer, 500, "500 - Internal Server Error")
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	if err := tmpl.ExecuteTemplate(writer, "indexSuperReport.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}

}
