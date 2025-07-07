package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

// Adding this constant for 25mb limit assuming the image is 20 and allowing 5mb headroom for text. Change this to a larger num or include on Server Global config
// const maxPostSize = 25 << 20

func CreatePostHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("EVENT Create post handler called!")

	type PostRequest struct {
		Title      string   `json:"title"`
		Body       string   `json:"body"`
		PostImg    string   `json:"image"`
		Categories []string `json:"categories"`
	}

	newPost := &PostRequest{}

	// initial check for post size
	request.Body = http.MaxBytesReader(writer, request.Body, Configuration.MaxPostSize.ToInt64())
	defer request.Body.Close()
	if err := request.ParseMultipartForm(Configuration.MaxPostSize.ToInt64()); err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "createpost: "+err.Error())
		return
	}

	// sanitize all elements of the post
	newPost.Title = database.Sanitize(request.FormValue("title"))
	newPost.Body = string(database.BasicHTMLSanitize(request.FormValue("body")))
	rawCategories := request.Form["categories"]
	newPost.Categories = make([]string, len(rawCategories))
	for i, cat := range rawCategories {
		newPost.Categories[i] = database.Sanitize(cat)
	}

	imgFileName, status, err := UploadImage(request)
	if err != nil && status != 200 {
		jsonProblemResponder(writer, status, "Something went wrong with the image", "upload image err: "+err.Error())
		return
	}
	newPost.PostImg = imgFileName

	post := &config.Post{
		UserID:     activeUser.ID, //we add it ourselves because we can't trust the user!
		Title:      newPost.Title,
		Body:       newPost.Body,
		PostImg:    newPost.PostImg,
		Categories: newPost.Categories,
	}

	err = db.ValidateTitle(post.Title)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, err.Error(), "createpost: "+err.Error())
		return
	}

	err = db.ValidatePostBody(post.Body)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, err.Error(), "createpost: "+err.Error())
		return
	}

	if len(post.Categories) == 0 {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Missing category", "CreatePost: Post creation failed. Missing category")
		return
	}

	if len(post.Categories) > 5 {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Too many categories", "CreatePost: Post creation failed. Too many categories")
		return
	}

	postId, err := db.AddPost(post)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "CreatePost: Post creation failed. "+err.Error())
		return
	}

	data := struct { //we send the id so website can redirect to the post's url
		PostId int64 `json:"post_id"`
	}{
		PostId: postId,
	}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("CreatePost: Failed to send final response!")
	}
}
