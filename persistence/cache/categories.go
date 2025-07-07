package cache

import (
	"forum/common/custom_errs"
	"sync"
)

type Categories struct {
	keys             []string
	cachedCategories sync.Map
	mu               sync.Mutex
	capacity         int
}

// Use this to get a category by name from cache
func (c *Cache) GetCategoryIdByName(name string) (int64, error) {
	if c == nil {
		return 0, custom_errs.ErrNilCache
	}
	cats := c.categories
	if val, ok := cats.cachedCategories.Load(name); ok {
		return val.(int64), nil
	}
	return 0, custom_errs.ErrFetchingFromCache
}

func (c *Cache) GetAllCategories() ([]string, error) {
	if c == nil {
		return nil, custom_errs.ErrNilCache
	}
	cats := c.categories
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(cats.keys) == 0 {
		return nil, custom_errs.ErrFetchingFromCache
	}
	//making a copy so that the returned value is safe from cache in the cache, even though it's most likely going to be unecessary due to the cache only growing in size, so no changes to a subslice can occur.
	categories := make([]string, len(cats.keys))
	copy(categories, cats.keys)
	return categories, nil
}

// Adds new category with name key and id value
func (c *Cache) AddCategory(name string, id int64) error {
	if c == nil {
		return custom_errs.ErrNilCache
	}
	cats := c.categories
	cats.cachedCategories.Store(name, id)

	cats.mu.Lock()
	cats.keys = append(cats.keys, name)
	cats.mu.Unlock()

	return nil
}

func NewCategories(capacity int) *Categories {
	return &Categories{capacity: capacity}
}
