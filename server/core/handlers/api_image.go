package handlers

import (
	"forum/persistence/database"
	"forum/server/core/config"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// Serves image on /image/<img name> from configured imgCon.PathPrefix
func serveImageAPI(w http.ResponseWriter, r *http.Request, db *database.DataBase, activeUser *config.User) {
	imageName := strings.TrimPrefix(r.URL.Path, "/image/")

	// Optional: sanitize input to avoid path traversal
	if strings.Contains(imageName, "..") || imageName == "" {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	hide := db.IsImageHidden(imageName)
	if hide && !(activeUser.Role == MOD || activeUser.Role == ADMIN) {
		http.Error(w, "Invalid image path", http.StatusNotFound)
		return
	}

	var pathAndFile []string
	pathAndFile = append(pathAndFile, Configuration.Images.PathPrefix...)
	pathAndFile = append(pathAndFile, imageName)
	imagePath := path.Join(pathAndFile...)
	file, err := os.Open(imagePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(buffer)

	file.Seek(0, 0)

	w.Header().Set("Content-Type", contentType)
	io.Copy(w, file)
}
