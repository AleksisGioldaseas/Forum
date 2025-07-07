package populate

import (
	"context"
	"errors"
	"forum/common/custom_errs"
	g "forum/global"
	"forum/persistence/database"

	"forum/server/core/config"
	"forum/utils"
	"log"
	"math/rand"
	"path/filepath"

	"encoding/json"
	"fmt"
	"os"
)

type SeedData struct {
	Categories []string          `json:"categories"`
	Users      []*config.User    `json:"users"`
	Posts      []*config.Post    `json:"posts"`
	Comments   []*config.Comment `json:"comments"`
}

func StartProcedure() {
	dbCtx, shutDownDb := context.WithCancel(context.Background())
	g.Configs.Database.Ctx = dbCtx
	defer shutDownDb()

	db, err := database.Open(g.Configs.Database)
	if err != nil {
		log.Println("Database error", err)
		panic(1)
	}
	defer db.Close()

	data, err := LoadSeedData(filepath.Join("populate_data", "seeds.json"))
	if err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}

	if err := PopulateDatabase(db, data); err != nil {
		log.Fatalf("Failed to populate database: %v", err)
	}

	src := filepath.Join("populate_data", "testimages")
	images, _ := os.ReadDir(src)
	for _, img := range images {

		src := filepath.Join(src, img.Name())
		dest := filepath.Join("data", "images", img.Name())

		if err := CopyFile(src, dest); err != nil {
			log.Fatalf("Failed to copy %s: %v", img, err)
		}
	}

	log.Println("Database populated successfully.")
}

func LoadSeedData(filePath string) (*SeedData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data SeedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func PopulateDatabase(db *database.DataBase, data *SeedData) error {
	fmt.Println("start populate")
	// Insert users
	for _, user := range data.Users {
		user.PasswordHash = utils.HashPass("mp", "", g.Configs.Handlers.XorKey)
		user.Email = fmt.Sprint(rand.Intn(10000)) + "@hotmail.com"

		if _, err := db.AddUser(user); err != nil {
			return fmt.Errorf("failed to insert user %s: %v", user.UserName, err)
		}

		if user.UserName == "admin" {
			ADMIN := 3
			err := db.UpdateRole(user.UserName, 14, ADMIN)
			if err != nil {
				fmt.Println(err)
				return fmt.Errorf("failed to promote existing user to administrator")
			}
		}

		if user.UserName == "mod" {
			MOD := 2
			err := db.UpdateRole(user.UserName, 14, MOD)
			if err != nil {
				return fmt.Errorf("failed to promote existing user to moderator")
			}
		}
	}

	for _, category := range data.Categories {
		if _, err := db.AddCategory(category); err != nil {
			return fmt.Errorf("failed to insert ccategory: %v", err)
		}
	}

	// Insert posts
	for _, post := range data.Posts {
		if _, err := db.AddPost(post); err != nil {
			return fmt.Errorf("failed to insert post %s: %v", post.Title, err)
		}
	}

	for _, comment := range data.Comments {
		if _, err := db.AddComment(comment); err != nil && !errors.Is(err, custom_errs.ErrNotificationFailed) {
			// return fmt.Errorf("failed to insert comment: %v", err)
			fmt.Printf("failed to insert comment: %v\n", err)
		}
	}

	err := db.UpdateUserKarma()
	if err != nil {
		fmt.Println(err.Error())
	}

	db.FuzzyPostTimes()

	fmt.Println("end populate")
	return nil
}
