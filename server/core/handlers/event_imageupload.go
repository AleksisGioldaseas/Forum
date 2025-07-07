package handlers

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/utils"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"slices"

	"github.com/google/uuid"
)

type imgConfig struct {
	MaxSize    utils.FileSize `json:"max_size"`
	Types      []string       `json:"file_types"`
	PathPrefix []string       `json:"path_prefix"`
}

// Parses the image file and stores it to the configured path. Returns a uuid as filename
func UploadImage(r *http.Request) (filename string, status int, err error) {
	fmt.Println("UploadImage called")
	file, header, err := r.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", 200, http.ErrMissingFile
		}
		return "", http.StatusBadRequest, custom_errs.ErrInvalidImageFile
	}
	defer file.Close()

	cfg := Configuration.Images

	if header.Size > cfg.MaxSize.ToInt64() {
		return "", http.StatusBadRequest, custom_errs.ErrImageTooBig
	}

	if !isImage(file, cfg.Types) {
		return "", http.StatusBadRequest, custom_errs.ErrInvalidImageFile
	}

	filename = uuid.New().String() + filepath.Ext(header.Filename)
	var pathAndFile []string
	pathAndFile = append(pathAndFile, cfg.PathPrefix...)
	pathAndFile = append(pathAndFile, filename)
	path := filepath.Join(pathAndFile...)

	dst, err := os.Create(path)
	if err != nil {
		return "", http.StatusInternalServerError, errors.New("unable to save file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", http.StatusInternalServerError, errors.New("failed to save file")
	}

	return filename, 200, nil
}

func isImage(file multipart.File, imageTypes []string) bool {
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return false
	}
	filetype := http.DetectContentType(buf)
	file.Seek(0, io.SeekStart)

	return slices.Contains(imageTypes, filetype)
}
