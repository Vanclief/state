package book

import "github.com/vanclief/state/object"

type Book struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Autor string `json:"autor"`
}

func New(id, name, autor string) *Book {
	return &Book{id, name, autor}
}

func (b *Book) Schema() *object.Schema {
	return &object.Schema{Name: "books", PKey: "id"}
}

func (b *Book) GetID() string {
	return b.ID
}
