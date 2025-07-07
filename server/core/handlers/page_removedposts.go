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

func RemovedPostsPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("RemovedPosts handler called!")

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
		OnlyRemoved:  true,
	}

	removedPosts, err := db.Search(tx, searchArgs)
	if err != nil {
		fmt.Println("Error with fetching removed posts: ", err)
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
		log.Print("removed posts page handler: Transaction Failure", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexRemovedPosts.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	jsonRemovedPosts, err := json.Marshal(struct {
		Posts []*config.Post `json:"posts"`
	}{Posts: removedPosts})
	if err != nil {
		fmt.Println("Error with marshaling removed posts: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type removedPageDataT struct {
		//all pages
		GenericResponse

		RemovedPostsJson string
	}

	data := removedPageDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},
		RemovedPostsJson: string(jsonRemovedPosts),
	}

	fmt.Println(data.NotificationAlert)

	if err := tmpl.ExecuteTemplate(writer, "indexRemovedPosts.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
