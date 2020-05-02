package simplecache

import (
	"fmt"

	"github.com/vanclief/ez"
	"github.com/vanclief/state/object"
)

// Cache defines a simple map cache
type Cache struct {
	memory map[string]object.Model
}

// New creates a new SimpleCache
func New() *Cache {
	return &Cache{map[string]object.Model{}}
}

// Get obtains a model from the cache
func (c *Cache) Get(m object.Model, id string) error {
	const op = "Simplecache.Cache.Get"

	key := m.Schema().PKey + "-" + id
	val, ok := c.memory[key]
	if !ok {
		msg := fmt.Sprintf("Object with key: %s was not found in the cache", key)
		return ez.New(op, ez.ENOTFOUND, msg, nil)
	}

	err := m.Update(val)
	if err != nil {
		return ez.New(op, ez.ECONFLICT, "Could not save retrieved object from cache", err)
	}

	return nil
}

// Set adds a model to the cache
func (c *Cache) Set(m object.Model) error {
	key := m.Schema().PKey + "-" + m.GetID()
	c.memory[key] = m
	return nil
}

// Delete removes a model from the cache
func (c *Cache) Delete(m object.Model) error {
	key := m.Schema().PKey + "-" + m.GetID()
	delete(c.memory, key)
	return nil
}
