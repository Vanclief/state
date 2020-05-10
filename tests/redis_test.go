package tests

import (
	"testing"

	"github.com/vanclief/ez"

	"github.com/stretchr/testify/assert"
	"github.com/vanclief/state/caches/redis"
	"github.com/vanclief/state/examplemodels/user"
	"github.com/vanclief/state/interfaces"
	"github.com/vanclief/state/manager"
)

func NewTestRedisCache() interfaces.Cache {
	const (
		address  = "localhost:6379"
		password = ""
		database = 1
	)
	cache, err := redis.New(address, password, database)
	if err != nil {
		panic(err)
	}
	return cache
}

func NewMockManagerWithRedis() *manager.Manager {
	// DB Setup
	db := NewTestDatabase()
	cache := NewTestRedisCache()

	state, err := manager.New(db, cache)
	if err != nil {
		panic(err)
	}

	return state
}

func TestGetwWithRedis(t *testing.T) {
	// Test Setup
	state := NewMockManagerWithRedis()
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

func TestQueryOneWithRedis(t *testing.T) {
	// Test Setup
	state := NewMockManagerWithRedis()
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

func TestQueryWithRedis(t *testing.T) {
	// Test Setup
	state := NewMockManagerWithRedis()
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
