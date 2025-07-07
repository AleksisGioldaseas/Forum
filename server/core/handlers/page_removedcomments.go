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

func RemovedCommentsPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("RemovedComments handler called!")

	activeUserIsLoggedIn := *activeUser != config.User{}

	tx, _ := db.ExportTx()
	defer tx.Rollback()

	removedComments, errs := db.GetComments(tx, 0, 15, 0, activeUser.ID, activeUser.Role, "new", false, true)
	if len(errs) != 0 {
		fmt.Println("Error with fetching removed comments: ", errs)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexRemovedComments.html")
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
		fmt.Println("removed comments handler: Tx commit failed: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	jsonRemovedComments, err := json.Marshal(struct {
		Comments []*config.Comment `json:"comments"`
	}{Comments: removedComments})
	if err != nil {
		fmt.Println("Error with marshaling Removed comment: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type removedCommentsPageDataT struct {
		//all pages
		GenericResponse

		JsonRemovedComments string
	}

	data := removedCommentsPageDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},
		JsonRemovedComments: string(jsonRemovedComments),
	}

	fmt.Println(data.NotificationAlert)

	if err := tmpl.ExecuteTemplate(writer, "indexRemovedComments.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
