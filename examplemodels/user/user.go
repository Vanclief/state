package user

import (
	"github.com/vanclief/ez"
	"github.com/vanclief/state/object"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func New(id, name, email string) *User {
	return &User{ID: id, Name: name, Email: email}
}

func Fixture() *User {
	return &User{ID: "1", Name: "Mock", Email: "mock@gmail.com"}
}

func (u *User) Schema() *object.Schema {
	return &object.Schema{Name: "users", PKey: "id"}
}

func (u *User) GetID() string {
	return u.ID
}

func (u *User) Update(i interface{}) error {
	const op = "User.Update"

	user, ok := i.(*User)
	if !ok {
		return ez.New(op, ez.EINVALID, "Provided interface is not of type User", nil)
	}

	*u = *user

	return nil
}
