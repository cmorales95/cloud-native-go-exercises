package service

import (
	"errors"
	"sync"
)

var (
	ErrorNoSuchKey = errors.New("error no such key")
	store          = struct {
		sync.RWMutex
		m map[string]string
	}{m: make(map[string]string)}
)

func Put(key, value string) error {
	store.Lock()
	store.m[key] = value
	store.Unlock()
	return nil
}

func Get(key string) (string, error) {
	store.RLock()
	value, ok := store.m[key]
	store.RUnlock()
	if !ok {
		return "", ErrorNoSuchKey
	}
	return value, nil
}

func Delete(key string) error {
	store.Lock()
	delete(store.m, key)
	store.Unlock()

	return nil
}
