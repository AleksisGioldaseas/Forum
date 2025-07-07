package database

import (
	"context"
	"errors"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"forum/utils"
	"testing"
	"time"
)

var dbConfig = &DBConfig{
	Path:     []string{"../", "../", "data", "test_forum.db"},
	Wal:      Wal{AutoTruncate: true, TruncateInterval: utils.Duration(1 * time.Hour)},
	UseCache: false,
}

func setupTestDB(t *testing.T) (*DataBase, context.Context, func()) {

	db, err := Open(dbConfig)
	if err != nil {
		t.Fatal("Failed to open test database:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db.Ctx = ctx

	// Clean up previous test data
	_, _ = db.conn.Exec("DELETE FROM User")
	_, err = db.conn.Exec("DELETE FROM Post")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = db.conn.Exec("DELETE FROM Comment")
	_, _ = db.conn.Exec("DELETE FROM Category")
	_, _ = db.conn.Exec("DELETE FROM PostCategory")
	_, _ = db.conn.Exec("DELETE FROM UserReactions")
	_, _ = db.conn.Exec("DELETE FROM Notifications")

	_, _ = db.conn.Exec("INSERT INTO User (Id, UserName, Email, PasswordHash, ProfilePic, Bio, TotalKarma, Role) VALUEs (1, 'vagelis', 'vag@vagelis.com', 'pwvag', 'pic', 'Hi its me, Vag', 1000, 1)")
	_, _ = db.conn.Exec("INSERT INTO User (Id, UserName, Email, PasswordHash, ProfilePic, Bio, TotalKarma, Role) VALUEs (2, 'Jim', 'jim@jim.com', 'pwjim', 'pic', 'Hi its me Jim', 500, 1)")

	_, _ = db.conn.Exec("INSERT INTO Post (Id, UserId, Title, Body, Img, RankScore) VALUES (1, 1, 'Post', 'Testing','pic', 0)")
	_, _ = db.conn.Exec("INSERT INTO Post (Id, UserId, Title, Body, Img, RankScore) VALUES (2, 1, 'Post 2', 'Testing', 'pic', 1000)")
	// _, _ = db.conn.Exec("INSERT INTO Post (Id, UserId, Title, Body, Img, RankScore) VALUES (3, 2, 'Post 2', 'Testing', 'pic', 1000)")

	_, _ = db.conn.Exec("INSERT INTO Comment (Id, UserId, PostId, Body) VALUES (1, 1, 1, 'Comment on post 1')")

	_, _ = db.conn.Exec("INSERT INTO Category (Id, Name) VALUES (1, 'music')")
	_, _ = db.conn.Exec("INSERT INTO Category (Id, Name) VALUES (2, 'coding')")

	_, _ = db.conn.Exec("INSERT INTO PostCategory (PostId, CategoryId) VALUES (1, 1)")
	_, _ = db.conn.Exec("INSERT INTO PostCategory (PostId, CategoryId) VALUES (1, 2)")
	_, _ = db.conn.Exec("INSERT INTO PostCategory (PostId, CategoryId) VALUES (2, 1)")
	_, _ = db.conn.Exec("INSERT INTO PostCategory (PostId, CategoryId) VALUES (2, 2)")
	return db, ctx, cancel
}

func TestAddUser(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	user := &config.User{
		UserName:     "George",
		Email:        "g@george.vag",
		PasswordHash: "pwg",
	}
	id, err := db.AddUser(user)
	if err != nil {
		t.Error(err, id)
	}
}

func TestGetUserByID(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	time.Sleep(1 * time.Second)

	user, err := db.GetUserById(nil, 1)
	if err != nil {
		t.Errorf("error getting user %v %v", 1, err)
	}
	if user.UserName != "vagelis" {
		t.Errorf("wrong user name got %v instead of 'vagelis'", user.UserName)
	}
}

func TestCreateCategory(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	_, err := db.AddCategory("bussines")
	if err != nil {
		t.Error(err)
	}
	_, err = db.AddCategory("coding")
	if err != nil {
		t.Error(err)
	}
	c, err := db.GetAllCategories()
	if err != nil {
		t.Error("error on getting all categories: ", err)
	}
	if len(c) != 3 {
		t.Error("exepcted 3 got :", c)
	}
}

func TestUpdateUserProfile(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	t.Run("No new", func(t *testing.T) {
		u := &config.User{
			ID: 1,
		}
		err := db.UpdateUserProfile(nil, u)
		if err == nil {
			t.Errorf("expected no new values error %v", err)
		}
	})

	t.Run("Just bio", func(t *testing.T) {
		bio := "this is me"
		u := &config.User{
			ID:  1,
			Bio: &bio,
		}
		err := db.UpdateUserProfile(nil, u)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("just pic", func(t *testing.T) {
		pic := "this is me"
		u := &config.User{
			ID:         1,
			ProfilePic: &pic,
		}
		err := db.UpdateUserProfile(nil, u)
		if err != nil {
			t.Error(err)
		}

	})

	t.Run("Both bio and pic", func(t *testing.T) {
		pic := "this is me"
		bio := "this is me"
		u := &config.User{
			ID:         1,
			ProfilePic: &pic,
			Bio:        &bio,
		}
		err := db.UpdateUserProfile(nil, u)
		if err != nil {
			t.Error(err)
		}

	})
}

func TestCreatePost(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	t.Run("Valid add post", func(t *testing.T) {
		post := &config.Post{
			UserID:     1,
			Title:      "Post",
			Body:       "Testing",
			Categories: []string{"music", "coding"},
		}
		id, err := db.AddPost(post)
		if err != nil || id == 0 {
			t.Error(err, id)
		}
	})

	t.Run("Invalid category on post", func(t *testing.T) {
		post := &config.Post{
			UserID:     1,
			Title:      "Post",
			Body:       "Testing",
			Categories: []string{"invalid"},
		}
		_, err := db.AddPost(post)
		if err == nil {
			t.Errorf("expected error 'add post: GetCategoryIdByName: sql: no rows in result' but got %v", err)
		}
	})
}

func TestSearchPost(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	t.Run("Search with keyword", func(t *testing.T) {
		args := &SearchArgs{
			ActiveUserId: 1,
			Sorting:      "hot",
			Filtering:    "all",
			Limit:        10,
			SearchQry:    "Post",
			IsModPlus:    true,
		}
		posts, err := db.Search(nil, args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(posts) != 2 {
			t.Errorf("Expected 2 posts, got %d", len(posts))
		}
	})

	t.Run("Search with category filter", func(t *testing.T) {
		args := &SearchArgs{
			ActiveUserId: 1,
			Sorting:      "hot",
			Filtering:    "all",
			Limit:        10,
			Categories:   []string{"music"},
			IsModPlus:    true,
		}
		posts, err := db.Search(nil, args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(posts) != 2 {
			t.Errorf("Expected 2 posts, got %d", len(posts))
		}
	})

	t.Run("Search with non-matching keyword", func(t *testing.T) {
		args := &SearchArgs{
			ActiveUserId: 1,
			Sorting:      "hot",
			Filtering:    "all",
			Limit:        10,
			SearchQry:    "NonExistent",
			IsModPlus:    true,
		}
		posts, err := db.Search(nil, args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(posts) != 0 {
			t.Errorf("Expected 0 posts, got %d", len(posts))
		}
	})

	t.Run("Search with category and keyword", func(t *testing.T) {
		args := &SearchArgs{
			ActiveUserId: 1,
			Sorting:      "hot",
			Filtering:    "all",
			Limit:        10,
			SearchQry:    "Post",
			Categories:   []string{"coding"},
			IsModPlus:    true,
		}
		posts, err := db.Search(nil, args)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(posts) != 2 {
			t.Errorf("Expected 2 posts, got %d", len(posts))
		}
	})
}

func TestGetUserByName(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	user, err := db.GetUserByUserName(nil, "vagelis", true)
	if err != nil {
		t.Errorf("error getting user %v %v", 1, err)
	}

	if user.ID != 1 {
		t.Errorf("wrong id got %v instead of 1", user.ID)
	}
}

func TestGetAuthUserByName(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	user, err := db.GetAuthUserByUserName("Vagelis")
	if err != nil {
		t.Fatalf("error getting user %v %v", 1, err)
	}

	if user.ID != 1 {
		t.Errorf("wrong id got %v instead of 1", user.ID)
	}
}

func TestPostsGetById(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	_, errSlice := db.GetPostById(nil, 0, 1, false)
	if errSlice != nil {
		t.Error("no posts")
	}
}

func TestGetPostsByCategory(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	_, err := db.GetPostsByCategories([]string{"music"}, 10, 0, 1, "ranking")
	if err != nil {
		t.Error(err)
	}

	_, err = db.GetPostsByCategories([]string{"music"}, 10, 0, 1, "created")
	if err != nil {
		t.Error(err)
	}

	_, err = db.GetPostsByCategories([]string{"music"}, 10, 0, 1, "karma")
	if err != nil {
		t.Error(err)
	}
}

func TestEdit(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	t.Run("Edit Comment", func(*testing.T) {
		err := db.EditComment(1, "New Body", 1)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Edit Post", func(*testing.T) {
		err := db.EditPost(1, "New Body", 1)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestToggleDeleteStatus(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)

	if err != nil {
		t.Error(err)
	}
	defer tx.Rollback()
	err = db.ToggleDeleteStatus(tx, 1, "post", 1, 1)
	if err != nil {
		t.Error(err)
	}
}

func TestTogglePostRemove(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()
	err := db.ModAssessment("post", 1, 1, "something", "god", 1)
	if err != nil {
		t.Error(err)
	}

}

func TestCreateComment(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()
	c := &config.Comment{
		UserID: 1,
		PostID: 1,
		Body:   "SomeComment",
	}
	_, err := db.AddComment(c)
	if err != nil && !errors.Is(err, custom_errs.ErrNotificationFailed) {
		t.Error(err)
	}
}

func TestGetComById(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()

	_, errs := db.GetComments(nil, 1, 10, 0, 0, 0, "top", false, false)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
}

func TestLike(t *testing.T) {
	db, ctx, cancel := setupTestDB(t)
	defer db.Close()
	defer cancel()

	t.Run("react on post", func(*testing.T) {
		err, _ := db.React(int64(1), int64(1), "post", "like")
		if err != nil {
			t.Error(err)
		}

		tx, err := db.conn.BeginTx(ctx, nil)
		if err != nil {
			t.Error(err)
		}
		defer tx.Rollback()

		num, _, err := db.getReaction(tx, int64(1), int64(1), "postId")
		if num != 1 || err != nil {
			t.Fatal(err)
		}
		tx.Commit()

		user, err := db.GetUserById(nil, 1)
		if user.TotalKarma != 1001 || err != nil {
			t.Error(err)
		}

		post, err := db.GetPostById(nil, int64(1), int64(1), false)
		if post.TotalKarma != 1 || err != nil {
			t.Error(err)
		}
	})

	t.Run("react on comment", func(*testing.T) {
		err, _ := db.React(int64(1), int64(1), "comment", "like")
		if err != nil {
			t.Error(err)
		}

		tx, err := db.conn.BeginTx(ctx, nil)
		if err != nil {
			t.Error(err)
		}
		defer tx.Rollback()

		num, _, err := db.getReaction(tx, int64(1), int64(1), "commentId")
		if num != 1 || err != nil {
			t.Fatal(err)
		}
		tx.Commit()

		user, err := db.GetUserById(nil, 1)
		if user.TotalKarma != 1001 || err != nil {
			t.Error(err)
		}

		comment, err := db.GetPostById(nil, int64(1), int64(1), false)
		if comment.TotalKarma != 1 || err != nil {
			t.Error(err)
		}
	})
}

func TestReport(t *testing.T) {
	db, _, _ := setupTestDB(t)
	defer db.Close()
	t.Run("place report", func(*testing.T) {
		err := db.Report(1, 1, 0, "spam")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get all reports", func(*testing.T) {
		_, err := db.GetReports(1, 0)
		if err != nil {
			t.Error(err)
		}
	})
}
