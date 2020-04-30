package state

import "github.com/vanclief/state/object"

// DB defines a database storage method
type DB interface {
	GetFromPKey(object.Model, string) error
	QueryOne(object.Model, string) error
	Query(interface{}, object.Model, string) error
	Insert(object.Model) error // Insert a row into the database
	Update(object.Model) error // Update a row from the database
	Delete(object.Model) error // Delete a row from the database
	CreateSchema([]interface{}, bool) error
}
