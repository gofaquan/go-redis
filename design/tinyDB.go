package design

import (
	"errors"
	"sync"
)

type TinyDB interface {
	Get(key string) (any, error)
	Set(key string, value any) error
	Del(key string) error
}

type redis struct {
	data map[string]any
	sync.RWMutex
}

func NewTinyDB() TinyDB {
	return &redis{
		map[string]any{},
		sync.RWMutex{},
	}
}

var (
	redisKeyNotFound = errors.New("redis key not found")
)

func (r *redis) Get(key string) (any, error) {
	r.Lock()
	defer r.Unlock()
	if v, ok := r.data[key]; ok {
		return v, nil
	}

	return nil, redisKeyNotFound
}

func (r *redis) Set(key string, value any) error {
	r.Lock()
	defer r.Unlock()
	r.data[key] = value
	return nil
}

func (r *redis) Del(key string) error {
	r.Lock()
	defer r.Unlock()
	delete(r.data, key)
	return nil
}
