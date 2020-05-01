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
func (c *Cache) Get(model object.Model, key string) error {
	const op = "Simplecache.Cache.Get"

	val, ok := c.memory[key]
	if !ok {
		msg := fmt.Sprintf("Object with key: %s was not found in the cache", key)
		return ez.New(op, ez.ENOTFOUND, msg, nil)
	}

	model.Update(val)

	return nil
}

// Set adds a model to the cache
func (c *Cache) Set(m object.Model) error {
	c.memory[m.GetID()] = m
	return nil
}

// Delete removes a model from the cache
func (c *Cache) Delete(m object.Model) error {
	delete(c.memory, m.GetID())
	return nil
}
