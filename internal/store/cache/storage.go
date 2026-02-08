package cache

import (
	"context"

	"github.com/Pedro-Foramilio/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
	}
}

func NewRedisStorage(rbd *redis.Client) *Storage {
	return &Storage{
		Users: &UserStore{rbd: rbd},
	}
}
