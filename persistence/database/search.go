package database

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/server/core/config"
	"strings"
)

type SearchArgs struct {
	ActiveUserId                                           int64
	Sorting, Filtering                                     string
	Limit, Offset                                          int
	SearchQry                                              string
	Categories                                             []string
	IsModPlus, OnlyRemoved, OnlySuperReports, OnlyReported bool
}

// Dynamicaly constructs query to search in posts by title
func (db *DataBase) Search(tx *sql.Tx, sa *SearchArgs) ([]*config.Post, error) {
	errorMsg := "search : %w"

	if sa.OnlyRemoved && !sa.IsModPlus || sa.OnlySuperReports && !sa.IsModPlus || sa.OnlyReported && !sa.IsModPlus {
		return nil, fmt.Errorf(errorMsg, "invalid call. only moderators can see removed posts")
	}

	posts := []*config.Post{}
	validSortColumns := map[string]string{
		"hot": "p.RankScore",
		"top": "p.TotalKarma",
		"new": "p.Created",
	}

	sorting, ok := validSortColumns[sa.Sorting]
	if !ok {
		return posts, fmt.Errorf(errorMsg, "unknown sorting value given")
	}

	// Values that will be given to query execution
	args := []any{sa.IsModPlus, sa.IsModPlus, sa.IsModPlus, sa.IsModPlus, sa.IsModPlus, sa.ActiveUserId}

	categoryIds := []int64{}
	for _, catName := range sa.Categories {
		id, err := db.GetCategoryIdByName(catName)
		if err != nil {
			fmt.Printf("can't find cat id from cat name: %v\n", err)
		}
		categoryIds = append(categoryIds, id)
	}

	extraJoins := []string{}
	for i, catId := range categoryIds {
		extraJoins = append(extraJoins, fmt.Sprintf("JOIN PostCategory pc%d ON p.Id = pc%d.PostId AND pc%d.CategoryId = %d", i, i, i, catId))
	}

	// WHERE conditions
	var whereClauses []string

	whereClauses = append(whereClauses, "p.Deleted = 0")

	//WHERE Removed = 0
	switch sa.Filtering {
	case "all":
	case "my-posts":
		whereClauses = append(whereClauses, "p.UserId = ?")
		args = append(args, sa.ActiveUserId)

	case "liked":
		whereClauses = append(whereClauses, "ur.Reaction = 1")
	default:
		return posts, fmt.Errorf(errorMsg, "unknown filtering value given")
	}

	if sa.SearchQry != "" {
		whereClauses = append(whereClauses, "p.Title LIKE ?")
		args = append(args, "%"+sa.SearchQry+"%")
	}

	if !sa.IsModPlus {
		whereClauses = append(whereClauses, "p.Removed = 0")
	}

	if sa.OnlyRemoved {
		whereClauses = append(whereClauses, "p.Removed = 1")
	}

	if sa.OnlyReported {
		whereClauses = append(whereClauses, "r.PostId IS NOT NULL")
		sorting = "MAX(r.Created)"
	}

	if sa.OnlySuperReports {
		whereClauses = append(whereClauses, "p.IsSuperReport = 1")
	} else {
		whereClauses = append(whereClauses, "p.IsSuperReport = 0")
	}

	// Construct the WHERE part of the query
	var whereQueryPart strings.Builder
	if len(whereClauses) > 0 {
		whereQueryPart.WriteString("WHERE ")
		whereQueryPart.WriteString(strings.Join(whereClauses, " AND "))
	}

	query := fmt.Sprintf(GENERIC_POST_SEARCH, strings.Join(extraJoins, "\n"), whereQueryPart.String(), sorting)
	args = append(args, sa.Limit, sa.Offset*sa.Limit)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	rows, err := exor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}
	defer rows.Close()
	var allErrors []error = nil
	firstRow := true
	for rows.Next() {
		post := config.NewPost()
		var categoriesStr string
		var reports string
		err := rows.Scan(
			&post.ID, &post.UserID, &post.UserName, &post.Title,
			&post.Body, &post.PostImg, &post.Likes, &post.Dislikes,
			&post.CreationDate, &post.UserReaction,
			&categoriesStr, &post.CommentCount, &post.Removed, &post.Edited,
			&post.Deleted, &post.UserRole, &reports, &post.RemovalReason, &post.ModeratorName, &post.IsSuperReport,
		)
		if err != nil {
			// Vag: Changed this to log errors and continue fetching posts
			allErrors = append(allErrors, err)
			continue
			// return nil, fmt.Errorf(errorMsg, err)
		}

		if reports != "" {
			m := make(map[string]struct{})
			for _, rep := range strings.Split(reports, "|!|!|") {
				m[rep] = struct{}{}
			}

			for uniqueKey := range m {
				post.Reports = append(post.Reports, uniqueKey)
			}
		}

		post.Categories = strings.Split(categoriesStr, ", ")

		posts = append(posts, post)

		if firstRow && db.UseCache {
			firstRow = false
			go func(post *config.Post) {
				if err := db.cache.Posts.Put(post); err != nil {
					wrappedErr := fmt.Errorf(errorMsg, err)
					fmt.Println(wrappedErr)
				}
			}(post)
		}
	}
	if allErrors != nil {
		err = fmt.Errorf(errorMsg, errors.Join(allErrors...))
	}

	return posts, err
}
