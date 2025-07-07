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

func HomePageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Home page handler called")
	if request.URL.Path != "/" && request.URL.Path != "/index" {
		isJson := request.Header.Get("Content-Type") == "application/json"
		if isJson {
			jsonProblemResponder(writer, http.StatusBadRequest, "Oups! Developer mistake most likely! uwu", "NON EXISTENT ENDPOINT")
			return
		}
		showErrorPage(writer, config.User{}, false, http.StatusNotFound, "404 - Page not found")
		return
	}

	fmt.Println("Home handler called:", request.URL.Path)

	activeUserIsLoggedIn := *activeUser != config.User{}
	searchArgs := &database.SearchArgs{
		ActiveUserId: activeUser.ID,
		Sorting:      "hot",
		Filtering:    "all",
		Limit:        5,
		IsModPlus:    activeUser.Role > 1,
	}

	tx, err := db.ExportTx()
	if err != nil {
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}
	defer tx.Rollback()

	posts, err := db.Search(tx, searchArgs)

	if err != nil {
		fmt.Println("error fetching posts from db: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	//turning posts into json, ready to be sent to front
	jsonText, err := json.Marshal(struct {
		Posts []*config.Post `json:"posts"`
	}{Posts: posts})
	if err != nil {
		fmt.Println("Error with marshaling posts: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	hasUnseen, notificationCount, err := db.CountNotifications(nil, tx, int(activeUser.ID))
	if err != nil {
		fmt.Println("Error unable to count notifiations: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	categories, err := db.GetAllCategories()
	if err != nil {
		log.Printf("unable to load all categories: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("home page handler: Tx failed", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexHome.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	// all data this handler will send to front
	type HomeDataT struct {
		//all pages
		GenericResponse

		//specific
		Categories []string
		Posts      []*config.Post
		JsonPosts  string
	}

	data := HomeDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},

		Categories: categories,
		JsonPosts:  string(jsonText),
		Posts:      posts,
	}

	if err := tmpl.ExecuteTemplate(writer, "indexHome.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
