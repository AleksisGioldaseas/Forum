package handlers

//handler for a users profile
import (
	"encoding/json"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

func ProfilePageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("profile handler called!")

	activeUserIsLoggedIn := *activeUser != config.User{}

	username := strings.TrimPrefix(request.URL.Path, "/profile/")

	user, err := db.GetUserByUserName(nil, username, false)
	if err != nil {
		fmt.Println("error fetching user: ", err)
		if errors.Is(err, custom_errs.ErrNameNotFound) {
			showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusNotFound, "404 - User Not Found")
			return
		} else {
			showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
			return
		}
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexProfile.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	jsonText, err := json.Marshal(user)
	if err != nil {
		fmt.Println("error marshaling user struct:", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type ProfileDataT struct {
		//all pages
		GenericResponse

		//specific
		User         config.User
		JsonUserData string
		ActiveUserId int64
		Bio          template.HTML
	}
	var bio string
	if user.Bio != nil {
		bio = *user.Bio
	}

	data := ProfileDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:       activeUserIsLoggedIn,
			ActiveUsername:   activeUser.UserName,
			ActiveProfilePic: activeUser.ProfilePic,
			ActiveUserRole:   activeUser.Role,
		},

		User:         *user,
		Bio:          template.HTML(bio),
		JsonUserData: string(jsonText),
	}

	if err := tmpl.ExecuteTemplate(writer, "indexProfile.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
