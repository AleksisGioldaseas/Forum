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
)

func SuperReportsViewPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("SuperReportsViewPageHandler handler called!")

	activeUserIsLoggedIn := *activeUser != config.User{}

	searchArgs := &database.SearchArgs{
		ActiveUserId:     activeUser.ID,
		Sorting:          "new",
		Filtering:        "all",
		Limit:            5,
		IsModPlus:        activeUser.Role > 1,
		OnlySuperReports: true,
	}

	tx, err := db.ExportTx()
	if err != nil {
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	defer tx.Rollback()
	superReports, err := db.Search(tx, searchArgs)
	if err != nil {
		fmt.Println("Error with fetching super reports: ", err)
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
		log.Println("SuperReportsViewPageHandler: Tx commit failed", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	if notificationCount != 0 {
		hasUnseen = true
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexSuperReportsView.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	fmt.Println(superReports)

	jsonReports, err := json.Marshal(struct {
		Posts []*config.Post `json:"posts"`
	}{Posts: superReports})
	if err != nil {
		fmt.Println("Error with marshaling posts: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type superReportsViewDataT struct {
		//all pages
		GenericResponse

		SuperReports string
	}

	data := superReportsViewDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},
		SuperReports: string(jsonReports),
	}

	if err := tmpl.ExecuteTemplate(writer, "indexSuperReportsView.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}

}
