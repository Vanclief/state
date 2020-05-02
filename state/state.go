package state

import (
	"fmt"

	"github.com/vanclief/ez"
	"github.com/vanclief/state/object"
)

// State defines an application state
type State struct {
	DB             Database
	Cache          Cache
	stagedChanges  []*Change
	appliedChanges []*Change
}

// New creates a new Application State from storage. It supports using a Database
// and a Cache or one of both.
func New(db Database, cache Cache) (*State, error) {
	const op = "State.New"

	if db == nil && cache == nil {
		return nil, ez.New(op, ez.EINVALID, "Creating a State requires at least a database or a cache", nil)
	}

	return &State{DB: db, Cache: cache}, nil
}

// Get obtains a model from the database using its ID, will attempt to fetch it
// first from Cache and then from Database
func (s *State) Get(model object.Model, query string) error {
	const op = "State.Select"

	var err error

	if s.Cache != nil {
		err = s.Cache.Get(model, query)
	}

	if s.DB != nil && err != nil {
		err = s.DB.Get(model, query)
		// TODO: If found & err == nil add to cache
	}

	if err != nil {
		return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
	}

	return nil
}

// QueryOne receives a model and a query. Will return a single model that
// satifies the query.
func (s *State) QueryOne(model object.Model, query string) error {
	const op = "State.QueryTest"

	if s.DB != nil {
		err := s.DB.QueryOne(model, query)
		if err != nil {
			return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
		}
	}

	return nil
}

// Query receives a model and a query. Will return all models that satisfies the
// query.
func (s *State) Query(mList interface{}, model object.Model, query ...string) error {
	const op = "State.Query"

	if s.DB != nil {
		err := s.DB.Query(mList, model, query)
		if err != nil {
			return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
		}
	}

	return nil
}

// Stage setups a model for changes, no change will be applied until State.Commit() is run
func (s *State) Stage(model object.Model, operation string) error {
	const op = "State.Stage"

	ch, err := NewChange(model, operation)
	if err != nil {
		return ez.New(op, ez.EINVALID, "Failed to stage model for changes", err)
	}

	s.stagedChanges = append(s.stagedChanges, ch)
	return nil
}

// Commit applies all of the staged changes
func (s *State) Commit() error {
	const op = "State.Commit"

	var err error
	s.appliedChanges = []*Change{}

	for _, change := range s.stagedChanges {
		err = change.Apply(s.DB, s.Cache)
		if change.status == "success" {
			s.appliedChanges = append(s.appliedChanges, change)
		}
	}

	if err != nil {
		return ez.New(op, ez.ECONFLICT, "One or more changes could not be commited", nil)
	}

	s.Clear()
	return nil
}

// Rollback reverts the latest applied changes for the insert operation
func (s *State) Rollback() error {
	const op = "State.Rollback"

	var err error
	rollbackChanges := s.appliedChanges
	s.appliedChanges = []*Change{}

	for _, change := range rollbackChanges {
		err = change.Revert(s.DB, s.Cache)
		if change.status != "reverted" {
			s.appliedChanges = append(s.appliedChanges, change)
		}
	}

	if err != nil {
		return ez.New(op, ez.ECONFLICT, "Could not rollback one or more changes", nil)
	}

	return nil
}

// Clear deletes the list of staged chanbes
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
