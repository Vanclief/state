package redis

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/vanclief/ez"
	"github.com/vanclief/state/interfaces"
)

type RedisStorage struct {
	Client *redis.Client
	ttl    int
}

// New instances a new redis client
func New(host, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	c, err := newRedis(client)
	if err != nil {
		return nil, err
	}
	if err := c.Client.Ping().Err(); err != nil {
		return nil, err
	}
	return c, nil
}

// Redis initializes a redis cache implementation wrapper
func newRedis(client *redis.Client) (*RedisStorage, error) {
	// Return wrapper
	return &RedisStorage{
		Client: client,
		ttl:    1,
	}, nil
}

func (s *RedisStorage) Get(m interfaces.Model, id string) error {
	key := m.GetSchema().PKey + "-" + id
	value, err := s.Client.Get(key).Bytes()
	if err == redis.Nil {
		return ez.New("redis.Get", ez.ENOTFOUND, "not found", err)
	}
	if err != nil {
		return ez.New("redis.Get", ez.EINTERNAL, "", err)
	}
	err = json.Unmarshal(value, &m)
	if err != nil {
		return err
	}
	return nil
}

func (s *RedisStorage) Set(m interfaces.Model, ttl int) error {
	key := m.GetSchema().PKey + "-" + m.GetID()
	encoded, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = s.Client.Set(key, encoded, time.Duration(ttl)*time.Millisecond).Err()
	if err != nil {
		return ez.New("redis.Set", ez.EINTERNAL, "", err)
	}
	return nil
}

func (s *RedisStorage) Delete(m interfaces.Model) error {
	key := m.GetSchema().PKey + "-" + m.GetID()
	err := s.Client.Del(key).Err()
	if err != nil {
		return ez.New("redis.Remove", ez.EINTERNAL, "", err)
	}
	return nil
}

func (s *RedisStorage) Purge() error {
	err := s.Client.Del("*").Err()
	if err != nil {
		return ez.New("redis.Purge", ez.EINTERNAL, "", err)
	}
	return nil
}

func (s *RedisStorage) GetTTL() int {
	return s.ttl
}

func (s *RedisStorage) SetTTL(ttl int) error {
	s.ttl = ttl
	return nil
}
