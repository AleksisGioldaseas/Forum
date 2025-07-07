package database

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"forum/server/core/sse"

	"log"
	"strings"
)

// ----------------------------------------
// ADD FUNCS                               |
// ----------------------------------------

// Updates database with post and links post to given categories
func (db *DataBase) AddPost(p *config.Post) (int64, error) {
	errorMsg := "add post: %w"
	if p.Title == "" && p.Body == "" && p.PostImg == "" {
		return 0, fmt.Errorf(errorMsg, "no title or body or image")
	}

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	defer tx.Rollback()

	err = db.ValidateCategories(p.Categories)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	// TODO: Remove p.Likes, p.Dislikes, p.RankScore for production
	result, err := tx.ExecContext(
		ctx,
		CREATE_POST,
		p.UserID, p.Title, p.Body,
		p.PostImg, p.Likes, p.Dislikes,
		p.RankScore, p.Removed,
		p.IsSuperReport, p.SuperReportCommentId,
		p.SuperReportPostId, p.SuperReportUserId,
	)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	// Adding postImage file name to images as with Hide = 0 (false)
	if p.PostImg != "" {
		if err = db.AddImage(tx, p.PostImg); err != nil {
			return 0, fmt.Errorf(errorMsg, err)
		}
	}

	p.ID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	for _, category := range p.Categories {
		if category == "" {
			continue
		}
		catId, err := db.GetCategoryIdByName(category)
		if err != nil {
			return 0, fmt.Errorf(errorMsg, err)
		}
		result, err := tx.ExecContext(ctx, LINK_POST_CATEGORY, p.ID, catId)
		if err != nil {
			return 0, fmt.Errorf(errorMsg, custom_errs.ErrLinkPostToCategory)
		}
		if err = CatchNoRowsErr(result); err != nil {
			return 0, fmt.Errorf(errorMsg, err)
		}
	}

	if p.IsSuperReport {
		fmt.Println("Sending mod report")
		adminIds, err := db.getAdminIds(tx)
		if err != nil {
			return 0, fmt.Errorf("issue with getting admin id's: %w", err)
		}
		for _, adminId := range adminIds {
			if adminId == p.UserID {
				continue
			}
			err = db.AddNotifSuper(tx, adminId, p.UserID, p.ID)
			if err != nil {
				return 0, fmt.Errorf("failed to add notification to admin: %w", err)
			}

			if err := sse.SendSSENotification(adminId); err != nil {
				if !errors.Is(err, custom_errs.ErrUserNotConnected) {
					fmt.Println("add post: SendSSENotification failed: ", err.Error())
				}
			}

		}
	}

	if !p.IsSuperReport {
		err = db.AddUserActivity(tx, p.UserID, &p.ID, nil, nil)
		if err != nil {
			return 0, fmt.Errorf(errorMsg, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	if db.UseCache && !p.IsSuperReport {
		go func(p *config.Post) {
			if err := db.cache.Posts.Put(p); err == nil {
				db.cache.UpdatePostCount(1)
			}
		}(p)
	}

	return p.ID, nil
}

// ----------------------------------------
// GET FUNCS                               |
// ----------------------------------------

// Updates post struct instance with data from db by post Id
func (db *DataBase) GetPostById(tx *sql.Tx, activeUserId, postId int64, isModPlus bool) (*config.Post, error) {
	errorMsg := "GetPostById: %w"

	if db.UseCache && activeUserId == GUEST {
		cached, err := db.cache.Posts.Get(postId)
		if err == nil {
			cached.UserReaction = nil
			return cached, nil
		}
	}

	p := config.NewPost()
	var postCats string
	var reports string

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	//test that post page can't get to report, and report can't get to post page
	err := exor.QueryRowContext(ctx, GET_POST_BY_ID, isModPlus, isModPlus, isModPlus, isModPlus, isModPlus, isModPlus, activeUserId, postId).
		Scan(
			&p.ID, &p.UserID, &p.UserName, &p.Title,
			&p.Body, &p.PostImg, &p.Likes, &p.Dislikes,
			&p.RankScore, &p.CreationDate, &p.TotalKarma, &p.UserReaction,
			&postCats, &p.Removed, &p.Deleted, &p.Edited,
			&p.IsSuperReport, &p.SuperReportCommentId, &p.SuperReportPostId, &p.SuperReportUserId,
			&reports, &p.RemovalReason, &p.ModeratorName, &p.CommentCount,
		)

	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	if reports != "" {

		m := make(map[string]struct{})
		for _, rep := range strings.Split(reports, "|!|!|") {
			m[rep] = struct{}{}
		}

		for uniqueKey := range m {
			p.Reports = append(p.Reports, uniqueKey)
		}
	}

	p.Categories = strings.Split(postCats, ", ")
	log.Println("Fetched post from db")

	if db.UseCache {
		db.putPostInCache(*p)
	}
	return p, nil
}

func (db *DataBase) GetPostsByCategories(categories []string, limit, offset int, activeUserId int64, sort string) ([]*config.Post, error) {
	errMsg := "GetPostsByCategories: %w"
	if limit > db.Limits.RowsLimit {
		return nil, fmt.Errorf(errMsg, custom_errs.ErrExceededRowsLimit)
	}

	switch sort {
	case "ranking":
		sort = "p.RankScore"
	case "karma":
		sort = "p.TotalKarma"
	case "created":
		sort = "p.Created"
	default:
		return nil, custom_errs.ErrInvalidSortingArg
	}
	placeholders := strings.Repeat("?,", len(categories))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(SORT_POSTS_BY_CATEGORY, placeholders)

	args := make([]any, len(categories)+4)
	args[0] = activeUserId
	for i, category := range categories {
		args[i+1] = category
	}
	args[len(categories)+1] = sort
	args[len(categories)+2] = limit
	args[len(categories)+3] = limit * offset

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	rows, err := exor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var allErrors []error = nil
	var posts []*config.Post
	for rows.Next() {
		p := config.NewPost()
		var postCats string
		err := rows.Scan(
			&p.ID, &p.UserID, &p.UserName, &p.Title,
			&p.Body, &p.PostImg, &p.Likes, &p.Dislikes,
			&p.RankScore, &p.CreationDate, &p.UserReaction,
			&postCats,
		)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}

		p.Categories = strings.Split(postCats, ", ")

		if p.Removed == TRUE || p.Deleted == TRUE {
			removedPostSetup(p)
		}

		if db.UseCache {
			db.putPostInCache(*p)
		}

		posts = append(posts, p)
	}

	if allErrors != nil {
		return posts, fmt.Errorf(errMsg+" multiple errors: %w", errors.Join(allErrors...))
	}
	return posts, err
}

func (db *DataBase) EditPost(postId int64, body string, activeUserId int64) (err error) {
	errMsg := "edit post %w"

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	result, err := exor.ExecContext(ctx, UPDATE_POST, body, postId, activeUserId)
	if err != nil {
		return
	}
	err = CatchNoRowsErr(result)

	if db.UseCache {
		db.deleteOrRestorePostInCache(postId, 0, errMsg)
	}
	return err
}

// COUNT --------------------------------------

func (db *DataBase) CountPosts() (int64, error) {
	if db.UseCache {
		if count, _ := db.cache.GetPostCount(); count != 0 {
			return count, nil
		}
	}
	var totalPosts int64

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	err := exor.QueryRowContext(ctx, COUNT_POSTS).Scan(&totalPosts)
	return totalPosts, err
}

func (db *DataBase) FuzzyPostTimes() {
	_, err := db.conn.Exec(FUZZY_POST_TIME)
	if err != nil {
		log.Fatal(err)
	}
}

func removedPostSetup(p *config.Post) {
	p.Body = "Content has been removed"
	// p.Categories = []string{}
	p.Dislikes = 0
	p.Likes = 0
	p.TotalKarma = 0
	p.PostImg = ""
	p.RankScore = 0
	p.UserImg = ""
	p.UserName = ""
	p.UserReaction = nil
}

func (db *DataBase) DeletePost(activeUserId, postId int64) error {
	errMsg := "delete post %w"

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	defer tx.Rollback()

	if err := db.ToggleDeleteStatus(tx, activeUserId, "post", postId, 1); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	if err = db.DeleteUserActivity(tx, activeUserId, &postId, nil, nil); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if db.UseCache {
		db.deleteOrRestorePostInCache(postId, REMOVE, errMsg)
	}
	return nil
}
