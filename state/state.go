package state

import (
	"fmt"

	"github.com/vanclief/ez"
	"github.com/vanclief/state/object"
)

// State defines an application state
type State struct {
	db             DB
	cache          Cache
	stagedChanges  []*Change
	appliedChanges []*Change
}

// New receives a DB and/or a Cache interface and returns a new state
func New(db DB, cache Cache) (*State, error) {
	const op = "State.New"

	if db == nil && cache == nil {
		return nil, ez.New(op, ez.EINVALID, "Creating a State requires at least a database or a cache", nil)
	}

	return &State{db: db, cache: cache}, nil
}

// Get obtains a model from the database using its ID
func (s *State) Get(model object.Model, query string) error {
	const op = "State.Select"

	var err error

	if s.cache != nil {
		err = s.cache.Get(model, query)
	} else if s.db != nil {
		err = s.db.GetFromPKey(model, query)
	}

	if err != nil {
		return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
	}

	return nil
}

// QueryOne searches for a single model that satisfies a Query statement
func (s *State) QueryOne(model object.Model, query string) error {
	const op = "State.QueryTest"

	if s.db != nil {
		err := s.db.QueryOne(model, query)
		if err != nil {
			return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
		}
	}

	return nil
}

// Query searches for multiple models using a statement
func (s *State) Query(mList interface{}, model object.Model, query string) error {
	const op = "State.Query"

	if s.db != nil {
		err := s.db.Query(mList, model, query)
		if err != nil {
			return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
		}
	}

	return nil
}

// Stage setups a model for applying changes
func (s *State) Stage(model object.Model, operation string) error {
	const op = "State.Stage"

	ch, err := NewChange(model, operation)
	if err != nil {
		return ez.New(op, ez.EINVALID, "Failed to stage model for changes", err)
	}

	s.stagedChanges = append(s.stagedChanges, ch)
	return nil
}

// Commit takes all staged changes and applies them
func (s *State) Commit() error {
	const op = "State.Commit"

	var err error
	s.appliedChanges = []*Change{}

	for _, change := range s.stagedChanges {
		err = change.Apply(s.db, s.cache)
		if change.status == "success" {
			s.appliedChanges = append(s.appliedChanges, change)
		}
	}

	if err != nil {
		return ez.New(op, ez.ECONFLICT, "One or more changes could not be commited", nil)
	}

	// If everything was ok, clear the current state
	s.Clear()
	return nil
}

// Rollback reverts the changes done in an Apply
func (s *State) Rollback() error {
	const op = "State.Rollback"

	var err error
	rollbackChanges := s.appliedChanges
	s.appliedChanges = []*Change{}

	for _, change := range rollbackChanges {
		err = change.Revert(s.db, s.cache)
		if change.status != "reverted" {
			s.appliedChanges = append(s.appliedChanges, change)
		}
	}

	if err != nil {
		return ez.New(op, ez.ECONFLICT, "Could not rollback one or more changes", nil)
	}

	return nil
}

// Clear removes all staged changes
func (s *State) Clear() {
	s.stagedChanges = []*Change{}
}

// Status returns the current list of staged changes
func (s *State) Status() []*Change {
	return s.stagedChanges
}

// Applied returns the previous list of applied changes
func (s *State) Applied() []*Change {
	return s.appliedChanges
}

// PrintStatus display the current status of staged changes
func (s *State) PrintStatus() {
	for _, change := range s.stagedChanges {
		fmt.Println("Model:", change.model, "OP:", change.op, "Status:", change.status, "Error:", change.err)
	}
}
