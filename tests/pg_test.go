package tests

import (
	"testing"

	"github.com/vanclief/ez"

	"github.com/stretchr/testify/assert"
	"github.com/vanclief/state/caches/simplecache"
	pg "github.com/vanclief/state/databases/pgdb"
	"github.com/vanclief/state/examplemodels/book"
	"github.com/vanclief/state/examplemodels/user"
	"github.com/vanclief/state/interfaces"
	"github.com/vanclief/state/manager"
)

func NewTestDatabase() interfaces.Database {
	const (
		address  = "localhost:5432"
		username = "vanclief"
		password = ""
		database = "postgres"
	)

	// Create a new database connection
	db, err := pg.New(address, username, password, database)
	if err != nil {
		panic(err)
	}

	dbModels := []interface{}{(&user.User{})}

	// Create the database schema
	err = db.CreateSchema(dbModels, true)
	if err != nil {
		panic(err)
	}

	return db
}

func NewTestCache() interfaces.Cache {
	return simplecache.New()
}

func NewMockManager() *manager.Manager {
	// DB Setup
	db := NewTestDatabase()
	cache := NewTestCache()

	state, err := manager.New(db, cache)
	if err != nil {
		panic(err)
	}

	return state
}

func TestNew(t *testing.T) {
	// Test Setup
	db := NewTestDatabase()
	cache := NewTestCache()

	// Should be able to create a new State with a Database and No Cache
	state, err := manager.New(db, nil)
	assert.Nil(t, err)
	assert.NotNil(t, state)

	// Should be able to create a new state with Cache and no Database
	state, err = manager.New(nil, cache)
	assert.Nil(t, err)
	assert.NotNil(t, state)

	// Should be able to create a new State with a Database and Cache
	state, err = manager.New(db, cache)
	assert.Nil(t, err)
	assert.NotNil(t, state)

	// Should NOT be able to create a new state without a Database or a Cache
	state, err = manager.New(nil, nil)
	assert.NotNil(t, err)
	assert.Nil(t, state)
}

func TestStage(t *testing.T) {
	// Test Setup
	state := NewMockManager()
	user := user.New("1", "Franco", "franco@gmail.com")

	// Should be able to stage insert
	err := state.Stage(user, "insert")
	assert.Nil(t, err)

	// Should be able to stage update
	user.Email = "franco.new@gmail.com"
	err = state.Stage(user, "update")
	assert.Nil(t, err)

	// Should be able to stage delete
	err = state.Stage(user, "delete")
	assert.Nil(t, err)
}

func TestCommit(t *testing.T) {
	// Test Setup
	state := NewMockManager()
	user := user.New("1", "Franco", "franco@gmail.com")

	// Should be able to apply insert
	state.Stage(user, "insert")
	err := state.Commit()
	assert.Nil(t, err)

	// Should have an empty state after successful apply
	assert.Len(t, state.Status(), 0)

	// Should be able to apply update
	user.Name = "Not Franco"
	state.Stage(user, "update")
	err = state.Commit()
	assert.Nil(t, err)

	// Should be able to apply delete
	state.Stage(user, "delete")
	err = state.Commit()
	assert.Nil(t, err)
}

func TestRollback(t *testing.T) {
	// Test Setup
	state := NewMockManager()
	user1 := user.New("1", "Franco", "franco@gmail.com")
	book := book.New("1", "El master fuster", "Franco") // Book is not in the database schema

	// Should not have an empty state after failure to apply a change
	state.Stage(user1, "insert")
	state.Stage(book, "insert")

	err := state.Commit()
	assert.NotNil(t, err)

	assert.Len(t, state.Applied(), 1)
	assert.Len(t, state.Status(), 2)

	// Should be able to rollback changes that where applied
	err = state.Rollback()
	assert.Len(t, state.Applied(), 0)
	assert.Len(t, state.Status(), 2)
	assert.Nil(t, err)

	// Should be able to rollback changes that where successful
	state.Clear()
	assert.Len(t, state.Status(), 0)

	user2 := user.New("2", "Juan", "juan@gmail.com")
	user3 := user.New("3", "Xin", "xin@gmail.com")

	state.Stage(user2, "insert")
	state.Stage(user3, "insert")
	err = state.Commit()
	assert.Nil(t, err)
	assert.Len(t, state.Applied(), 2)

	err = state.Rollback()
	assert.Nil(t, err)
	assert.Len(t, state.Status(), 0)
	assert.Len(t, state.Applied(), 0)
}

func TestGet(t *testing.T) {
	// Test Setup
	state := NewMockManager()
	user1 := user.New("1", "Franco", "franco@gmail.com")
	state.Stage(user1, "insert")
	state.Commit()
	state.Cache.Purge()

	// Should be able to get a model that exists
	res := &user.User{}
	err := state.Get(res, "1")
	assert.Nil(t, err)

	assert.Equal(t, user1.ID, res.ID)
	assert.Equal(t, user1.Name, res.Name)
	assert.Equal(t, user1.Email, res.Email)

	// Should not be able to get a model that doesnt exist
	res = &user.User{}
	err = state.Get(res, "31231")
	assert.NotNil(t, err)
	assert.Equal(t, ez.ENOTFOUND, ez.ErrorCode(err))
}

func TestQueryOne(t *testing.T) {
	// Test Setup
	state := NewMockManager()
	user1 := user.New("1", "Franco", "franco@gmail.com")
	state.Stage(user1, "insert")
	state.Commit()

	// Should be able to get a model that exists
	res := &user.User{}
	err := state.QueryOne(res, `email = 'franco@gmail.com'`)
	assert.Nil(t, err)

	assert.Equal(t, user1.ID, res.ID)
	assert.Equal(t, user1.Name, res.Name)
	assert.Equal(t, user1.Email, res.Email)

	// Should fail if there is no model that matches the query
	res = &user.User{}
	err = state.QueryOne(res, `email = 'arcano@gmail.com'`)
	assert.NotNil(t, err)
	assert.Equal(t, ez.ENOTFOUND, ez.ErrorCode(err))

	// Should fail if there is more than one model that matches the query
	user2 := user.New("2", "Franco's Impostor", "franco@gmail.com")
	state.Stage(user2, "insert")
	state.Commit()

	res = &user.User{}
	err = state.QueryOne(res, `email = 'franco@gmail.com'`)
	assert.NotNil(t, err)
	assert.Equal(t, ez.ECONFLICT, ez.ErrorCode(err))
}

func TestQuery(t *testing.T) {
	// Test Setup
	state := NewMockManager()
	user1 := user.New("1", "Franco", "email@francovalencia.com")
	user2 := user.New("2", "Franco", "franco@gmail.com")
	user3 := user.New("3", "Vanclief", "vanclief@vanclief.com")
	state.Stage(user1, "insert")
	state.Stage(user2, "insert")
	state.Stage(user3, "insert")
	state.Commit()

	// Should be able to get a model that exists
	res := []user.User{}
	err := state.Query(&res, &user.User{}, `name = 'Franco'`)
	assert.Nil(t, err)

	assert.Len(t, res, 2)
	assert.Equal(t, user1.ID, res[0].ID)
	assert.Equal(t, user1.Name, res[0].Name)
	assert.Equal(t, user1.Email, res[0].Email)

	assert.Equal(t, user2.ID, res[1].ID)
	assert.Equal(t, user2.Name, res[1].Name)
	assert.Equal(t, user2.Email, res[1].Email)

	// Should fail if there is no model that matches the query
	res = []user.User{}
	err = state.Query(&res, &user.User{}, `name = 'Francisco'`)
	assert.NotNil(t, err)
	assert.Equal(t, ez.ENOTFOUND, ez.ErrorCode(err))

	// Should be able to use limit in the query
	res = []user.User{}
	err = state.Query(&res, &user.User{}, `name = 'Franco'`, "1")
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, user1.ID, res[0].ID)
	assert.Equal(t, user1.Name, res[0].Name)
	assert.Equal(t, user1.Email, res[0].Email)

	// Should be able to use limit and offset in the query
	res = []user.User{}
	err = state.Query(&res, &user.User{}, `name = 'Franco'`, "1", "1")
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, user2.ID, res[0].ID)
	assert.Equal(t, user2.Name, res[0].Name)
	assert.Equal(t, user2.Email, res[0].Email)

	// Should be able to use limit with order by in the query
	res = []user.User{}
	err = state.Query(&res, &user.User{}, `name = 'Franco' ORDER BY email DESC`, "1")
	assert.Nil(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, user2.ID, res[0].ID)
	assert.Equal(t, user2.Name, res[0].Name)
	assert.Equal(t, user2.Email, res[0].Email)

}
