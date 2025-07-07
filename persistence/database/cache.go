package database

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/cache"
	"forum/server/core/config"
	"log"
	"sync"
)

// Loads all categories, user count and post count to cache
func (db *DataBase) loadCache() error {
	fmt.Println("*** CACHE ***")
	if db.cacheCfg == nil {
		return errors.New("error undefined cache configuration parameters")
	}

	if db.cache == nil {
		db.UseCache = false
		return custom_errs.ErrNilCache
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	wg.Add(4)

	log.Println("Updating cache...")
	go func() {
		defer wg.Done()
		if err := db.loadCategoryMap(); err != nil {
			errCh <- err
		}
		log.Println("Updated categories")
	}()

	go func() {
		defer wg.Done()
		var err error
		usersCount, err := db.CountUsers()
		if err == nil {
			db.cache.NewUserCount(usersCount)
			log.Println("Updated User Count")
			return
		}

		errCh <- err
	}()

	go func() {
		defer wg.Done()
		var err error
		postsCount, err := db.CountPosts()
		if err == nil {
			db.cache.NewPostCount(postsCount)
			log.Println("Updated Post Count")
			return
		}

		errCh <- err
	}()

	go func() {
		defer wg.Done()
		var err error
		searchArgs := &SearchArgs{
			Sorting:   "hot",
			Filtering: "all",
			Limit:     db.cacheCfg.PostsLimit,
		}
		posts, err := db.Search(nil, searchArgs)

		if err != nil {
			fmt.Println("search err on populate: ", err)
		}

		for i := len(posts) - 1; i >= 0; i-- {
			if err := db.cache.Posts.Put(posts[i]); err != nil {
				fmt.Println(err)
			}
		}

		errCh <- err
	}()

	wg.Wait()
	close(errCh)
	errs := []error{}
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors loading cache: %w", errors.Join(errs...))
	}

	log.Print("Cache updated successfuly")

	return nil
}

// Creates a new instace of db.cache and reloads from db
func (db *DataBase) RefreshCache() error {
	log.Println("Reloading Cache...")
	db.mu.Lock()
	defer db.mu.Unlock()

	db.cache = cache.NewCache(db.cacheCfg)
	return db.loadCache()
}

// Creates a cache map of [category name]categoryId
func (db *DataBase) loadCategoryMap() error {
	errorMsg := "load category map: %w"

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	rows, err := exor.QueryContext(ctx, `SELECT Name, Id FROM Category`)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var id int64
		if err := rows.Scan(&name, &id); err != nil {
			return fmt.Errorf(errorMsg, err)
		}
		if err := db.cache.AddCategory(name, id); err != nil {
			err = fmt.Errorf(errorMsg, err)
			log.Println(err.Error())
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	return nil
}

func (db *DataBase) deleteOrRestoreUserInCache(activeUserId int64, status int, errorMsg string) {
	go func(activeUserId int64) {
		var err error
		if status == REMOVE {
			if err = db.cache.Users.RemoveWithUID(activeUserId); err == nil {
				db.cache.UpdateUserCount(-1)
			}
		} else if status == RESTORE {
			_, err = db.GetUserById(nil, activeUserId)
			if err == nil {
				db.cache.UpdateUserCount(1)
			} else {
				err = fmt.Errorf(errorMsg, err)
				log.Println(err.Error())
			}
		}
		if err != nil {
			err = fmt.Errorf(errorMsg, err)
			fmt.Println(err.Error())
		} else {
			err = fmt.Errorf(errorMsg, err)
			log.Println(err.Error())
		}
	}(activeUserId)
}

func (db *DataBase) deleteOrRestorePostInCache(postId int64, status int, errorMsg string) {
	go func(rowId int64) {
		var err error
		if status == REMOVE {
			if err = db.cache.Posts.Remove(rowId); err == nil {
				db.cache.UpdatePostCount(-1)
			} else {
				err = fmt.Errorf(errorMsg, err)
				log.Println(err.Error())
			}
		} else if status == RESTORE {
			p, err := db.GetPostById(nil, -1, rowId, false)
			if err == nil {
				if err = db.cache.Posts.Put(p); err == nil {
					db.cache.UpdatePostCount(1)
				} else {
					err = fmt.Errorf(errorMsg, err)
					log.Println(err.Error())
				}
			}
		}
		if err != nil {
			err = fmt.Errorf(errorMsg, err)
			log.Println(err.Error())
		}
	}(postId)
}

func (db *DataBase) putPostInCache(p config.Post) {
	errorMsg := "put post in cache: %w"
	if p.Removed == TRUE || p.Deleted == TRUE || p.IsSuperReport {
		return
	}
	p.UserReaction = nil
	go func(p config.Post) {
		if err := db.cache.Posts.Put(&p); err != nil {
			err = fmt.Errorf(errorMsg, err)
			log.Println(err.Error())
		}
	}(p)
}

func (db *DataBase) putUserInCache(u config.User) {
	errorMsg := "put user in cache: %w"
	if u.Removed == REMOVE || u.Deleted == REMOVE {
		return
	}

	go func(u config.User) {
		if err := db.cache.Users.Put(&u); err != nil {
			err = fmt.Errorf(errorMsg, err)
			log.Println(err.Error())
		}
	}(u)
}
