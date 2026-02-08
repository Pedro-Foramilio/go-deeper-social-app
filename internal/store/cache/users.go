package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Pedro-Foramilio/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	rbd *redis.Client
}

const cacheKey = "user-%v"

func (s *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	data, err := s.rbd.Get(ctx, fmt.Sprintf(cacheKey, id)).Result()
	if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	finalKey := fmt.Sprintf(cacheKey, user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rbd.SetEX(ctx, finalKey, data, time.Minute).Err()
}
