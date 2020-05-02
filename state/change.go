package state

import (
	"github.com/vanclief/ez"
	"github.com/vanclief/state/object"
)

// Operation codes
const (
	INSERT = "insert"
	UPDATE = "update"
	DELETE = "delete"
)

// Status codes
const (
	PENDING  = "pending"
	SUCCESS  = "success"
	FAILURE  = "failure"
	REVERTED = "reverted"
)

// Change defines the current state of a model that has been staged for changed
type Change struct {
	model  object.Model
	op     string
	status string
	err    error
}

// NewChange creates a new Change struct which is reponsible for tracking changes applied
// to the application state
func NewChange(m object.Model, operation string) (*Change, error) {
	const op = "Changes.New"

	ops := make(map[string]struct{})

	ops[INSERT] = struct{}{}
	ops[UPDATE] = struct{}{}
	ops[DELETE] = struct{}{}

	_, ok := ops[operation]
	if !ok {
		return nil, ez.New(op, ez.EINVALID, "Operation type is not suported", nil)
	}

	return &Change{model: m, op: operation, status: "pending"}, nil
}

// Apply executes a pending change
func (ch *Change) Apply(db Database, cache Cache) error {
	const op = "Changes.Apply"

	// Ignore changes that have been successfuly applied or reverted
	if ch.status == "success" || ch.status == "reverted" {
		return nil
	}

	switch ch.op {
	case INSERT:
		if db != nil {
			err := db.Insert(ch.model)
			if err != nil {
				ch.status = FAILURE
				ch.err = err
				return ez.New(op+".INSERT", ez.EINTERNAL, "Database: Could apply insert operation", err)
			}
			ch.status = SUCCESS
		}

		if cache != nil {
			err := cache.Set(ch.model, cache.GetTTL())
			if err != nil {
				ch.status = FAILURE
				ch.err = err
				return ez.New(op+".INSERT", ez.EINTERNAL, "Cache: Could apply set operation", err)
			}
			ch.status = SUCCESS
		}
	case UPDATE:
		if db != nil {
			err := db.Update(ch.model)
			if err != nil {
				ch.status = FAILURE
				ch.err = err
				return ez.New(op+".UPDATE", ez.EINTERNAL, "Database: Could apply update operation", err)
			}
			ch.status = SUCCESS
		}

		if cache != nil {
			err := cache.Set(ch.model, cache.GetTTL())
			if err != nil {
				ch.status = FAILURE
				ch.err = err
				return ez.New(op+".UPDATE", ez.EINTERNAL, "Cache: Could apply set operation", err)
			}
			ch.status = SUCCESS
		}
	case DELETE:
		if db != nil {
			err := db.Delete(ch.model)
			if err != nil {
				ch.status = FAILURE
				ch.err = err
				return ez.New(op+".DELETE", ez.EINTERNAL, "Database: Could apply delete operation", err)
			}
			ch.status = SUCCESS
		}

		if cache != nil {
			err := cache.Delete(ch.model)
			if err != nil {
				ch.status = FAILURE
				ch.err = err
				return ez.New(op+".DELETE", ez.EINTERNAL, "Cache: Could apply delete operation", err)
			}
			ch.status = SUCCESS
		}
	}

	return nil
}

// Revert executes the reverse action of a change, currently only supports insert
func (ch *Change) Revert(db Database, cache Cache) error {
	const op = "Changes.Revert"

	// Ignore changes that have not been successfuly applied
	if ch.status != "success" {
		return nil
	}

	switch ch.op {
	case INSERT:
		if db != nil {
			err := db.Delete(ch.model)
			if err != nil {
				ch.status = SUCCESS
				ch.err = err
				return ez.New(op+".INSERT", ez.EINTERNAL, "Database: Could revert insert operation", err)
			}
			ch.status = REVERTED
		}

		if cache != nil {
			err := cache.Delete(ch.model)
			if err != nil {
				ch.status = SUCCESS
				ch.err = err
				return ez.New(op+".INSERT", ez.EINTERNAL, "Cache: Could revert set operation", err)
			}
			ch.status = REVERTED
		}
	}

	return nil
}
