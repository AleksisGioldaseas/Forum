package handlers_test

import (
	"bytes"
	"forum/persistence/database"
	"forum/server/core/config"
	"forum/server/core/handlers"
	"forum/utils"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

/* Instructions:
1) run `make db-test`
2) from database_test.go run `TestAddUser`
3) add dir `persistence/data/testdata` and put a photo test.png that is less than 20Mb
4) run `go test ./server/core/handlers/` or click run test
*/

func startDb(t *testing.T) *database.DataBase {
	var dbConfig = &database.DBConfig{
		Path:     []string{"../", "../", "../", "persistence", "data", "test_forum.db"},
		Wal:      database.Wal{AutoTruncate: true, TruncateInterval: utils.Duration(1 * time.Hour)},
		UseCache: false,
	}
	db, err := database.Open(dbConfig)
	if err != nil {
		t.Fatal("Failed to open test database:", err)
	}
	return db

}

func TestCreatePostHandler(t *testing.T) {
	db := startDb(t)

	// Create a buffer to hold multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	_ = writer.WriteField("title", "Test Post")
	_ = writer.WriteField("body", "This is a test body.")
	_ = writer.WriteField("categories", "music")
	_ = writer.WriteField("categories", "coding")

	// Add an image file field
	fileWriter, err := writer.CreateFormFile("image", "test.jpg")
	if err != nil {
		t.Fatal(err)
	}

	// Open an image file (or you can use a dummy byte stream)
	testImagePath := filepath.Join("../", "../", "../", "persistence", "data", "testdata", "test.png")
	file, err := os.Open(testImagePath)
	if err != nil {
		t.Fatalf("Cannot open test image: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		t.Fatalf("Cannot copy image data: %v", err)
	}

	writer.Close()

	// Create a request
	req := httptest.NewRequest(http.MethodPost, "/create-post", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Limit body size to simulate the real middleware
	req.Body = http.MaxBytesReader(nil, req.Body, 10<<20) // 10MB

	// Create a response recorder
	rec := httptest.NewRecorder()

	// Call the handler
	activeUser := &config.User{ID: 1}
	handlers.CreatePostHandler(rec, req, db, activeUser)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", res.StatusCode)
	}

	respBody, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(respBody), `"success":true`) {
		t.Errorf("Expected response to contain post_id 123, got %s", string(respBody))
	}
}
