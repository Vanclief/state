package state

import "github.com/vanclief/state/object"

// Cache defines a cache storage method
type Cache interface {
	// Get attempts to retrieve a model using its ID as Key, if found it
	// updates with it the receiving Model.
	Get(object.Model, string) error
	// Set adds a model  to the Cache using its ID as Key
	Set(object.Model, int) error
	// Delete destroys a model stored in the Cache
	Delete(object.Model) error
	// GetTTL returns the currently set TTL (Time To Live) of the Cache
	GetTTL() int
	// SetTTL updates the TTL (Time To Live) of the Cache
	SetTTL(int) error
	// Purge clears the cache
	Purge() error
}
