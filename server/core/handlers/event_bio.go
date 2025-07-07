package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func BioHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("BioHandler called")
	type Bio struct {
		Bio string `json:"bio"`
	}

	newBio := &Bio{}
	err := jsonRequestExtractor(request, newBio)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "biohandler: "+err.Error())
		return
	}

	sanitizedBio := string(database.BasicHTMLSanitize(newBio.Bio))

	err = db.ValidateBio(sanitizedBio)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "biohandler: "+err.Error())
		return
	}

	activeUser.Bio = &sanitizedBio

	activeUser.ProfilePic = nil
	err = db.UpdateUserProfile(nil, activeUser)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Update Bio: Database error:. "+err.Error())
		return
	}
	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("Update Bio:: Failed to send final response!")
	}
}
