package book

import (
	"github.com/vanclief/ez"
	"github.com/vanclief/state/interfaces"
)

type Book struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Autor string `json:"autor"`
}

func New(id, name, autor string) *Book {
	return &Book{id, name, autor}
}

func (b *Book) GetSchema() *interfaces.Schema {
	return &interfaces.Schema{Name: "books", PKey: "id"}
}

func (b *Book) GetID() string {
	return b.ID
}

func (b *Book) Update(i interface{}) error {
	const op = "Book.Update"

	book, ok := i.(Book)
	if !ok {
		return ez.New(op, ez.EINVALID, "Provided interface is not of type Book", nil)
	}

	*b = book
	return nil
}
