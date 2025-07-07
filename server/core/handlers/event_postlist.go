package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func GetPostListHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("PostList Handler called!")

	type postListRequest struct {
		SearchQuery       string   `json:"search_query"`
		Count             int      `json:"count"`
		Page              int      `json:"page"` //page 1, 2, 3 and so on
		Sorttype          string   `json:"sort_type"`
		Filtertype        string   `json:"filter_type"`
		Categories        []string `json:"categories"`
		OnlySuperReports  bool     `json:"only_super_reports"`
		OnlyRemovedPosts  bool     `json:"only_removed_posts"`
		OnlyReportedPosts bool     `json:"only_reported_posts"`
	}

	postsRequest := &postListRequest{}
	err := jsonRequestExtractor(request, postsRequest)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "post list:"+err.Error())
		return
	}

	if activeUser.ID == 0 {
		postsRequest.Filtertype = "all"
	}
	fmt.Println("post list request: ", postsRequest.OnlySuperReports)
	args := &database.SearchArgs{
		ActiveUserId:     activeUser.ID,
		Sorting:          postsRequest.Sorttype,
		Filtering:        postsRequest.Filtertype,
		Limit:            postsRequest.Count,
		Offset:           postsRequest.Page,
		SearchQry:        postsRequest.SearchQuery,
		Categories:       postsRequest.Categories,
		IsModPlus:        activeUser.Role > 1,
		OnlyRemoved:      postsRequest.OnlyRemovedPosts,
		OnlySuperReports: postsRequest.OnlySuperReports,
		OnlyReported:     postsRequest.OnlyReportedPosts,
	}
	Posts, err := db.Search(nil, args)

	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "GetPostList: Failed to find posts. "+fmt.Sprint(err))
		return
	}

	err = jsonOkResponder(writer, Posts)
	if err != nil {
		fmt.Println("Postlist: Failed to send final response!")
	}

}
