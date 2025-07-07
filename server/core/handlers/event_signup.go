package handlers

//handles signup requests from the front

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"forum/utils"
	"net/http"
)

func SignupHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Signup Handler called!")

	type SignupRequest struct {
		Username       string `json:"username"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		PasswordRepeat string `json:"passwordRepeat"`
	}

	signReq := &SignupRequest{}
	err := jsonRequestExtractor(request, signReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "signup: "+err.Error())
		return
	}

	// Call signup
	_, err = signup(db, signReq.Username, signReq.Email, signReq.Password, signReq.PasswordRepeat)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, err.Error(), "Signup: Signup failed. "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil, "Registration succeeded")
	if err != nil {
		fmt.Println("Signup: Failed to send final response!")
	}
}

// executes the signup procedure
func signup(db *database.DataBase, username, email, password, passwordRepeat string) (*config.User, error) {
	// Sanitize and validate inputs
	username = database.Sanitize(username)
	email = database.Sanitize(email)

	if err := db.ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := database.ValidateEmail(email); err != nil {
		return nil, err
	}
	if err := db.ValidatePassword(password, passwordRepeat); err != nil {
		return nil, err
	}

	tx, err := db.ExportTx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check for dupes
	uniqueName, err := db.IsUsernameUnique(tx, username)
	if err != nil {
		return nil, errors.New("failed to check for username dupe: " + err.Error())
	}
	if !uniqueName {
		return nil, custom_errs.ErrUsernameNotUnique
	}
	uniqueMail, err := db.IsEmailUnique(tx, email)
	if err != nil {
		return nil, errors.New("failed to check for email dupe: " + err.Error())
	}
	if !uniqueMail {
		return nil, custom_errs.ErrEmailNotUnique
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	userSalt, err := utils.Salt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	hashedPass := utils.HashPass(password, userSalt, Configuration.XorKey)

	defaultPfp := "default_pfp.jpg"

	user := &config.User{
		UserName:     username,
		Email:        email,
		PasswordHash: hashedPass,
		Salt:         userSalt,
		ProfilePic:   &defaultPfp,
	}

	_, err = db.AddUser(user)
	if err != nil {
		fmt.Println("Failed to add user: ", err)
		return user, err
	}

	return user, nil
}
