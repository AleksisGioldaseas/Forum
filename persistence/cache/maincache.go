package cache

import (
	"forum/common/custom_errs"
	"sync"
	"sync/atomic"
)

type Cache struct {
	Users      *userCache
	categories *Categories
	Posts      *postCache
	usersCount int64
	postsCount int64
	mu         sync.Mutex
}

type CHConfig struct {
	UsersLimit      int
	PostsLimit      int
	CategoriesLimit int
}

func NewCache(cfg *CHConfig) *Cache {
	return &Cache{
		Users:      NewUserCache(cfg.UsersLimit),
		Posts:      NewPostCache(cfg.PostsLimit),
		categories: NewCategories(cfg.CategoriesLimit),
	}
}

// Use this to get the total users in db from cache
// Concurrent safe using atomic opetations
func (c *Cache) GetUserCount() (int64, error) {
	if c == nil {
		return 0, custom_errs.ErrNilCache
	}
	return atomic.LoadInt64(&c.usersCount), nil
}

// Use this to get total posts in db from cache
// Concurrent safe using atomic opetations
func (c *Cache) GetPostCount() (int64, error) {
	if c == nil {
		return 0, custom_errs.ErrNilCache
	}
	return atomic.LoadInt64(&c.postsCount), nil
}

// Increment this when adding or removing users
func (c *Cache) UpdateUserCount(increment int64) error {
	if c == nil {
		return custom_errs.ErrNilCache
	}
	atomic.AddInt64(&c.usersCount, increment)
	return nil
}

// Incrementthis when adding or removing posts
func (c *Cache) UpdatePostCount(increment int64) error {
	if c == nil {
		return custom_errs.ErrNilCache
	}
	atomic.AddInt64(&c.postsCount, increment)
	return nil
}

func (c *Cache) NewUserCount(users int64) error {
	if c == nil {
		return custom_errs.ErrNilCache
	}
	atomic.StoreInt64(&c.usersCount, users)
	return nil
}

func (c *Cache) NewPostCount(posts int64) error {
	if c == nil {
		return custom_errs.ErrNilCache
	}
	atomic.StoreInt64(&c.postsCount, posts)
	return nil
}
