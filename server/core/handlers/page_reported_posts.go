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

func ReportedPostsPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("ReportedPosts handler called!")

	activeUserIsLoggedIn := *activeUser != config.User{}

	tx, err := db.ExportTx()
	if err != nil {
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}
	defer tx.Rollback()

	searchArgs := &database.SearchArgs{
		ActiveUserId: activeUser.ID,
		Sorting:      "hot",
		Filtering:    "all",
		Limit:        5,
		IsModPlus:    activeUser.Role > 1,
		OnlyReported: true,
	}

	ReportedPosts, err := db.Search(tx, searchArgs)
	if err != nil {
		fmt.Println("Error with fetching Reported posts: ", err)
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
		log.Print("reported posts page handler: Transaction Failure", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexReportedPosts.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	jsonReportedPosts, err := json.Marshal(struct {
		Posts []*config.Post `json:"posts"`
	}{Posts: ReportedPosts})
	if err != nil {
		fmt.Println("Error with marshaling Reported posts: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type ReportedPageDataT struct {
		//all pages
		GenericResponse

		ReportedPostsJson string
	}

	data := ReportedPageDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},
		ReportedPostsJson: string(jsonReportedPosts),
	}

	if err := tmpl.ExecuteTemplate(writer, "indexReportedPosts.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
