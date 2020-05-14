package manager

import (
	"fmt"

	log "github.com/inconshreveable/log15"
	"github.com/vanclief/ez"
	"github.com/vanclief/state/interfaces"
)

// Manager defines an application state controller
type Manager struct {
	DB             interfaces.Database
	Cache          interfaces.Cache
	stagedChanges  []*Change
	appliedChanges []*Change
	logging        bool
}

// New creates a new Application State Manager from storage. It supports using a Database
// and a Cache or one of both.
func New(db interfaces.Database, cache interfaces.Cache) (*Manager, error) {
	const op = "Manager.New"

	if db == nil && cache == nil {
		return nil, ez.New(op, ez.EINVALID, "Creating a State Manager requires at least a database or a cache", nil)
	}

	return &Manager{DB: db, Cache: cache}, nil
}

// Get obtains a model from the database using its ID, will attempt to fetch it
// first from Cache and then from Database
func (m *Manager) Get(model interfaces.Model, id interface{}) error {
	const op = "Manager.Select"

	var err error

	if m.Cache != nil {
		m.log(op, "Source", "Cache", "ID", id)
		err = m.Cache.Get(model, id)
		m.logError(op, err, "Source", "Cache", "ID", id)
	}

	if m.DB != nil && err != nil {
		m.log(op, "Source", "DB", "ID", id)
		err = m.DB.Get(model, id)
		m.logError(op, err, "Source", "DB", "ID", id)
	}

	// TODO: If found in database and no error add to cache

	if err != nil {
		return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
	}

	return nil
}

// QueryOne receives a model and a query. Will return a single model that
// satifies the query.
func (m *Manager) QueryOne(model interfaces.Model, query string) error {
	const op = "Manager.QueryTest"

	if m.DB != nil {
		m.log(op, "Query", query)

		err := m.DB.QueryOne(model, query)
		if err != nil {
			m.logError(op, err, "Source", "DB", "ID", query)
			return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
		}
	}

	return nil
}

// Query receives a model and a query. Will return all models that satisfies the
// query.
func (m *Manager) Query(mList interface{}, model interfaces.Model, query ...string) error {
	const op = "Manager.Query"

	if m.DB != nil {
		m.log(op, "Query", query)

		err := m.DB.Query(mList, model, query)
		if err != nil {
			m.logError(op, err, "Source", "DB", "ID", query)
			return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
		}
	}

	return nil
}

// Stage setups a model for changes, no change will be applied until State.Commit() is run
func (m *Manager) Stage(model interfaces.Model, operation string) error {
	const op = "Manager.Stage"

	m.log(op, "Model", model.GetSchema(), "ID", model.GetID())
	ch, err := NewChange(model, operation)

	if err != nil {
		m.logError(op, err, "Model", model.GetSchema(), "ID", model.GetID())
		return ez.New(op, ez.EINVALID, "Failed to stage model for changes", err)
	}

	m.stagedChanges = append(m.stagedChanges, ch)
	return nil
}

// Commit applies all of the staged changes
func (m *Manager) Commit() error {
	const op = "Manager.Commit"

	var err error
	m.appliedChanges = []*Change{}

	for _, change := range m.stagedChanges {
		err = change.Apply(m.DB, m.Cache)
		if change.status == "success" {
			m.appliedChanges = append(m.appliedChanges, change)
		}
	}

	if err != nil {
		return ez.New(op, ez.ECONFLICT, "One or more changes could not be commited", nil)
	}

	m.Clear()
	return nil
}

// Rollback reverts the latest applied changes for the insert operation
func (m *Manager) Rollback() error {
	const op = "Manager.Rollback"

	var err error
	rollbackChanges := m.appliedChanges
	m.appliedChanges = []*Change{}

	for _, change := range rollbackChanges {
		err = change.Revert(m.DB, m.Cache)
		if change.status != "reverted" {
			m.appliedChanges = append(m.appliedChanges, change)
		}
	}

	if err != nil {
		return ez.New(op, ez.ECONFLICT, "Could not rollback one or more changes", nil)
	}

	return nil
}

// Clear deletes the list of staged chanbes
func (m *Manager) Clear() {
	m.stagedChanges = []*Change{}
}

// Status returns the current list of staged changes
func (m *Manager) Status() []*Change {
	return m.stagedChanges
}

// Applied returns the previous list of applied changes
func (m *Manager) Applied() []*Change {
	return m.appliedChanges
}

// PrintStatus display the current status of staged changes
func (m *Manager) PrintStatus() {
	for _, change := range m.stagedChanges {
		fmt.Println("Model:", change.model, "OP:", change.op, "Status:", change.status, "Error:", change.err)
	}
}

// ToggleLogs enables or disables detailed logs
func (m *Manager) ToggleLogs() {
	m.logging = !m.logging
}

func (m *Manager) log(op string, ctx ...interface{}) {
	if m.logging {
		log.Info(op, ctx...)
	}
}

func (m *Manager) logError(op string, err error, ctx ...interface{}) {
	if m.logging && err != nil {
		ctx = append(ctx, "Error", ez.ErrorMessage(err))
		log.Info(op, ctx...)
	}
}
