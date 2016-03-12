package session

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/redis.v3"
)

const (
	// MemoryDialect indicates use in memory store
	MemoryDialect = 1 + iota
	// RedisDialect indicates use redis store
	RedisDialect
)

type registry interface {
	get(string) ([]byte, error)
	set(string, []byte, time.Duration) error
	dispose() error
}

type memoryRegistry struct {
	memory    map[string][]byte
	namespace string
	lock      *sync.RWMutex
}

type redisRegistry struct {
	client    *redis.Client
	namespace string
}

func newRegistry(namespace string, dialect int, options *redis.Options) registry {
	if dialect == RedisDialect {
		newClient := redis.NewClient(options)
		_, err := newClient.Ping().Result()
		if err != nil {
			panic(err)
		}
		return &redisRegistry{namespace: namespace, client: newClient}
	}

	return &memoryRegistry{
		namespace: namespace,
		memory:    make(map[string][]byte, 0),
		lock:      new(sync.RWMutex),
	}
}

func (r *memoryRegistry) dispose() error {
	return nil
}

func (r *redisRegistry) dispose() error {
	return r.client.Close()
}

func (r *memoryRegistry) get(key string) ([]byte, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.memory[r.key(key)], nil
}

func (r *redisRegistry) get(key string) ([]byte, error) {
	result, err := r.client.Get(r.key(key)).Result()
	if err == redis.Nil {
		return make([]byte, 0), nil
	} else if err != nil {
		return make([]byte, 0), err
	} else {
		return []byte(result), nil
	}
}

func (r *memoryRegistry) set(key string, value []byte, expiration time.Duration) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.memory[r.key(key)] = value
	return nil
}

func (r *redisRegistry) set(key string, value []byte, expiration time.Duration) error {
	return r.client.Set(r.key(key), value, expiration).Err()
}

func (r *memoryRegistry) key(k string) string {
	return fmt.Sprintf("%s:%s", r.namespace, k)
}

func (r *redisRegistry) key(k string) string {
	return fmt.Sprintf("%s:%s", r.namespace, k)
}
