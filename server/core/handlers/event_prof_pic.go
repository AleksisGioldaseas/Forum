package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"log"
	"net/http"
)

func ProfilePicHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("ProfilePic Handler called!")

	imgFileName, status, err := UploadImage(request)
	if err != nil {
		jsonProblemResponder(writer, status, "", "profile pic: "+err.Error())
		return
	}

	tx, _ := db.ExportTx()
	defer tx.Rollback()

	activeUser.ProfilePic = &imgFileName
	err = db.UpdateUserProfile(tx, activeUser)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Profile Pic: Database fail: "+err.Error())
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Print("profile pic handler handler: Transaction Failure", err)
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Profile Pic: Database fail: "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("ProfilePic: Failed to send final response!")
	}
}
