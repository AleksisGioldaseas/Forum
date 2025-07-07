package handlers

//handler for home page

import (
	"encoding/json"
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"html/template"
	"log"
	"net/http"
	"path"
)

func ReportedCommentsPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("ReportedComments handler called!")

	activeUserIsLoggedIn := *activeUser != config.User{}

	tx, _ := db.ExportTx()
	defer tx.Rollback()

	ReportedComments, errs := db.GetComments(tx, 0, 15, 0, activeUser.ID, activeUser.Role, "new", true, false)
	if len(errs) != 0 {
		fmt.Println("Error with fetching Reported comments: ", errs)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexReportedComments.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
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
		fmt.Println("ReportedCommentsPageHandler: Tx commit failed: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	jsonReportedComments, err := json.Marshal(struct {
		Comments []*config.Comment `json:"comments"`
	}{Comments: ReportedComments})
	if err != nil {
		fmt.Println("Error with marshaling Reported comment: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type reportedCommentsPageDataT struct {
		//all pages
		GenericResponse

		JsonReportedComments string
	}

	data := reportedCommentsPageDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},
		JsonReportedComments: string(jsonReportedComments),
	}

	for _, c := range ReportedComments {
		fmt.Println("comment: ", c)
	}

	// fmt.Println(data.NotificationAlert)

	if err := tmpl.ExecuteTemplate(writer, "indexReportedComments.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
