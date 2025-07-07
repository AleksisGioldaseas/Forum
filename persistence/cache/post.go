package cache

import (
	"container/list"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"log"
	"sync"
)

type postCache struct {
	mu          sync.Mutex
	capacity    int
	cachedPosts map[int64]*list.Element
	list        *list.List
}

type postCacheEntry struct {
	post      *config.Post
	frequency int
}

func (uc *postCache) GetAll(limit, offset int) ([]*config.Post, error) {
	errMsg := "cache: post: get all: %w"
	if limit > uc.capacity || limit*offset > uc.capacity {
		return nil, fmt.Errorf(errMsg, "limit exceeds cache limit")
	}

	if offset < 0 || limit < 0 {
		return nil, fmt.Errorf(errMsg, "limit and offset must be non-negative")
	}
	if limit == 0 {
		return []*config.Post{}, nil
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	var posts []*config.Post
	count := 0
	for e := uc.list.Front(); e != nil; e = e.Next() {
		if count < offset {
			count++
			continue
		}
		if len(posts) >= limit {
			break
		}
		entry, ok := e.Value.(*postCacheEntry)
		if !ok {
			return nil, fmt.Errorf(errMsg, "unexpected entry type in list")
		}
		posts = append(posts, entry.post)
		count++
	}
	if len(posts) == 0 {
		return nil, fmt.Errorf(errMsg, "no posts in cache within given offset/limit")
	}
	return posts, nil
}

func (uc *postCache) Get(id int64) (*config.Post, error) {
	if uc == nil {
		return nil, custom_errs.ErrNilCache
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if elem, exists := uc.cachedPosts[id]; exists {
		uc.list.MoveToFront(elem)
		entry := elem.Value.(*postCacheEntry)
		entry.frequency++
		log.Println("Fetched post from cache") // debug
		return entry.post, nil
	}

	// log.Println(custom_errs.ErrFetchingFromCache.Error())
	return nil, custom_errs.ErrFetchingFromCache
}

// Put inserts or updates a post in the cache and deletes last
func (uc *postCache) Put(post *config.Post) error {
	errorMsg := "cache: post: put: %w"
	if uc == nil {
		return fmt.Errorf(errorMsg, custom_errs.ErrNilCache)
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if uc.list.Len() >= uc.capacity {
		lastElem := uc.list.Back()
		if lastElem != nil {
			removedPost := lastElem.Value.(*postCacheEntry).post
			delete(uc.cachedPosts, removedPost.ID)
			uc.list.Remove(lastElem)
			log.Println("Removed last element from cache") // debug
		}
	}

	if elem, exists := uc.cachedPosts[post.ID]; exists {
		entry := elem.Value.(*postCacheEntry)
		entry.frequency++
		entry.post = post
		uc.list.MoveToFront(elem)
	} else {
		entry := &postCacheEntry{post: post, frequency: 1}
		elem := uc.list.PushFront(entry)
		uc.cachedPosts[post.ID] = elem
	}

	if len(uc.cachedPosts) == 0 {
		log.Println(custom_errs.ErrPlacingInCache.Error())
		return fmt.Errorf(errorMsg, custom_errs.ErrPlacingInCache)
	}
	return nil
}

func (uc *postCache) Remove(id int64) error {
	errMsg := "cache: post: remove: %w"
	if uc == nil {
		return custom_errs.ErrNilCache
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()

	elem, exists := uc.cachedPosts[id]
	if !exists {
		log.Println("post id not in posts cache")
		return fmt.Errorf(errMsg, custom_errs.ErrNoRows)
	}

	uc.list.Remove(elem)

	delete(uc.cachedPosts, id)
	return nil
}

func NewPostCache(capacity int) *postCache {
	return &postCache{
		capacity:    capacity,
		cachedPosts: make(map[int64]*list.Element),
		list:        list.New(),
	}
}
