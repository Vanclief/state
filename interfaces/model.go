package interfaces

// Model defines a struct with properties that should be part of the application state
type Model interface {
	GetSchema() *Schema
	GetID() string
	Update(interface{}) error
}

// Schema defines the structure of the model for storage
type Schema struct {
	Name string
	PKey string
}
