package state

import "github.com/vanclief/state/object"

// Cache defines a cache storage method
type Cache interface {
	Get(object.Model, string) error
	Set(object.Model) error
	Delete(object.Model) error
	Purge() error
}
