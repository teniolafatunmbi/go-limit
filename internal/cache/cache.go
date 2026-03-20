package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type WindowKey string
type WindowValue struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}
type WindowMap map[WindowKey]*WindowValue

const (
	CurrentWindowKey  WindowKey = "current"
	PreviousWindowKey WindowKey = "previous"
)

type Cache struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Cache {
	return &Cache{
		rdb: rdb,
	}
}

func (c *Cache) GetCurrentWindow(ctx context.Context) ([]byte, error) {
	return c.rdb.Get(ctx, string(CurrentWindowKey)).Bytes()
}

func (c *Cache) GetPreviousWindow(ctx context.Context) ([]byte, error) {
	return c.rdb.Get(ctx, string(PreviousWindowKey)).Bytes()
}

func (c *Cache) SetCurrentWindow(ctx context.Context, value []byte) error {
	err := c.rdb.Set(ctx, string(CurrentWindowKey), value, 0).Err()

	return err
}

func (c *Cache) SetPreviousWindow(ctx context.Context, value []byte) error {
	err := c.rdb.Set(ctx, string(PreviousWindowKey), value, 0).Err()

	return err
}

func (c *Cache) DoesCurrentWindowExist(ctx context.Context) (*int64, error) {
	currWindowExists, err := c.rdb.Exists(ctx, string(CurrentWindowKey)).Result()

	if err != nil {
		return nil, err
	}

	return &currWindowExists, nil
}

func (c *Cache) DoesPreviousWindowExist(ctx context.Context) (*int64, error) {
	previousWindowExists, err := c.rdb.Exists(ctx, string(PreviousWindowKey)).Result()

	if err != nil {
		return nil, err
	}

	return &previousWindowExists, nil
}
