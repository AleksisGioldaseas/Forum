package cache_test

import (
	"forum/persistence/cache"
	"forum/server/core/config"
	"testing"
)

var nc = &cache.Cache{
	Users: cache.NewUserCache(10),
	Posts: cache.NewPostCache(10),
}

func TestUserPutGet(t *testing.T) {
	user := &config.User{
		ID:       1,
		UserName: "VaGeliS",
	}
	nc.Users.Put(user)
	_, err := nc.Users.Get("vagelis")
	if err != nil {
		t.Error()
	}
	nc.UpdateUserCount(1)
	uc, _ := nc.GetUserCount()
	if uc != 1 {
		t.Error(uc)
	}
}

func TestPostPutGet(t *testing.T) {
	user := &config.Post{
		ID:       1,
		UserName: "vagelis",
	}

	nc.Posts.Put(user)
	_, err := nc.Posts.Get(1)
	if err != nil {
		t.Error()
	}
}

func TestUserCount(t *testing.T) {

}
