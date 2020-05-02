package state

import "github.com/vanclief/state/object"

// Database defines a persistent storage method
type Database interface {
	// Get returns a Model from the database using its ID as PK
	Get(object.Model, string) error
	// QueryOne returns a Model from the database that satisfies a Query. Should
	// return error if it finds more than one model that satisfies the Query
	QueryOne(object.Model, string) error
	// Query returns all Model from the database that satisfy a Query
	Query(interface{}, object.Model, []string) error
	// Insert a model into the database using its ID as PK
	Insert(object.Model) error
	// Update an existing model into the database
	Update(object.Model) error
	// Delete an existing model from the database
	Delete(object.Model) error
	// CreateSchema if applicable, prepares the database Schema to store the different
	// application Models
	CreateSchema([]interface{}, bool) error
}
