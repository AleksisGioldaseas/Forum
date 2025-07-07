package handlers

import (
	"encoding/json"
	"forum/persistence/database"
	"forum/server/core/config"
	"log"
	"net/http"
	"strconv"
)

func userActivityAPI(w http.ResponseWriter, r *http.Request, db *database.DataBase, activeUser *config.User) {
	activeUserIsLoggedIn := activeUser.ID != 0
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")
	userName := r.URL.Query().Get("username")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		http.Error(w, "Invalid limit", http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, "Invalid page", http.StatusBadRequest)
		return
	}

	userActivity, err := db.GetUserActivitiesByUname(nil, userName, limit, page)
	if err != nil {
		log.Printf("Error getting user activities: %v", err)
		showErrorPage(w, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "500 - Internal Server Error")
		return
	}

	err = json.NewEncoder(w).Encode(userActivity)
	if err != nil {
		showErrorPage(w, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "500 - Internal Server Error")
		return
	}
}
