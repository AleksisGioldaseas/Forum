package cache

import (
	"container/list"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"log"
	"strings"
	"sync"
)

type userCache struct {
	mu          sync.Mutex
	capacity    int
	cachedUsers map[string]*list.Element
	list        *list.List
}

type userCacheEntry struct {
	user      *config.User
	frequency int
}

func NewUserCache(capacity int) *userCache {
	return &userCache{
		capacity:    capacity,
		cachedUsers: make(map[string]*list.Element),
		list:        list.New(),
	}
}

func (uc *userCache) Get(username string) (*config.User, error) {
	if uc == nil {
		return nil, custom_errs.ErrNilCache
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	if elem, exists := uc.cachedUsers[strings.ToLower(username)]; exists {
		uc.list.MoveToFront(elem)
		entry := elem.Value.(*userCacheEntry)
		entry.frequency++
		log.Println("Fetched user from cache")
		return entry.user, nil
	}

	log.Println(custom_errs.ErrFetchingFromCache.Error())
	return nil, custom_errs.ErrFetchingFromCache
}

// Put inserts or updates a user in the cache and deletes last
func (uc *userCache) Put(user *config.User) error {
	if uc == nil {
		return custom_errs.ErrNilCache
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	userName := strings.ToLower(user.UserName)
	if elem, exists := uc.cachedUsers[userName]; exists {
		entry := elem.Value.(*userCacheEntry)
		entry.frequency++
		uc.list.MoveToFront(elem)
		// log.Println("Added user to cache") //commenting this out to remove terminal clutter
	}

	if uc.list.Len() >= uc.capacity {
		lastElem := uc.list.Back()
		if lastElem != nil {
			removedUser := lastElem.Value.(*userCacheEntry).user
			delete(uc.cachedUsers, strings.ToLower(removedUser.UserName))
			uc.list.Remove(lastElem)
			// log.Println("Removed last element from cache")
		}
	}

	entry := &userCacheEntry{user: user, frequency: 1}
	elem := uc.list.PushFront(entry)
	uc.cachedUsers[userName] = elem

	if len(uc.cachedUsers) == 0 {
		log.Println(custom_errs.ErrPlacingInCache.Error())
		return custom_errs.ErrPlacingInCache
	}
	return nil
}

func (uc *userCache) Remove(username string) error {
	if uc == nil {
		return custom_errs.ErrNilCache
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	username = strings.ToLower(username)

	elem, exists := uc.cachedUsers[username]
	if !exists {
		return fmt.Errorf("remove user from cache: %w", custom_errs.ErrNoRows)
	}

	uc.list.Remove(elem)

	delete(uc.cachedUsers, username)
	return nil
}

func (uc *userCache) RemoveWithUID(UId int64) error {
	if uc == nil {
		return custom_errs.ErrNilCache
	}

	for _, u := range uc.cachedUsers {
		entry := u.Value.(*userCacheEntry)
		if entry.user.ID == UId {
			uc.Remove(entry.user.UserName)
			return nil
		}
	}
	return custom_errs.ErrNoRows
}
