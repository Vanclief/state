package pgdb

import (
	"fmt"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/vanclief/ez"
	"github.com/vanclief/state/interfaces"
)

const (
	// ETABLEEXISTS happens when a table already exists
	ETABLEEXISTS = "ERROR #42P07"
	// ENOROWS happens when no result was found
	ENOROWS = "pg: no rows in result set"
	// EMULTIPLEROWS happens when multiple results where found with QueryOne
	EMULTIPLEROWS = "pg: multiple rows in result set"
)

// DB defines a PostgreSQL database that will use go-pg as an ORM
type DB struct {
	pg *pg.DB
}

// New returns a new PG Database instance
func New(address string, user string, password string, database string) (*DB, error) {
	db := pg.Connect(&pg.Options{
		Addr:     address,
		User:     user,
		Password: password,
		Database: database,
	})

	_, err := db.Exec("SELECT 1")

	if err != nil {
		db.Close()
		return nil, err
	}

	return &DB{pg: db}, nil
}

// Get returns a single model from the database using its primary key
func (db *DB) Get(m interfaces.Model, ID interface{}) error {
	const op = "PG.DB.Get"

	switch ID.(type) {
	case string:
	case []byte:
	default:
		return ez.New(op, ez.EINVALID, "Can not use provided interface type", nil)
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s = ?`, m.GetSchema().Name, m.GetSchema().PKey)

	res, err := db.pg.QueryOne(m, query, ID)

	if res != nil && res.RowsReturned() < 1 {
		msg := fmt.Sprintf("Could not find a %s model with id %s", m.GetSchema().Name, ID)
		return ez.New(op, ez.ENOTFOUND, msg, nil)
	}

	if err != nil {
		switch err.Error() {
		case ENOROWS:
			msg := fmt.Sprintf("Could not find a %s model with id %s", m.GetSchema().Name, ID)
			return ez.New(op, ez.ENOTFOUND, msg, nil)
		default:
			return ez.New(op, ez.EINTERNAL, "Error making query to the database", err)
		}
	}

	return nil
}

// QueryOne returns a single model from the database that satisfies a Query.
// The method will return an error if there is more than one result from the query
func (db *DB) QueryOne(m interfaces.Model, query string) error {
	const op = "PG.DB.QueryOne"

	q := fmt.Sprintf(`SELECT * FROM %s WHERE %s`, m.GetSchema().Name, query)

	_, err := db.pg.QueryOne(m, q, nil)
	if err != nil {
		switch err.Error() {
		case ENOROWS:
			msg := fmt.Sprintf("Could not find a %s model with query %s", m.GetSchema().Name, query)
			return ez.New(op, ez.ENOTFOUND, msg, nil)
		case EMULTIPLEROWS:
			msg := fmt.Sprintf("Could find multiple %s models that satisfy QueryOne %s", m.GetSchema().Name, query)
			return ez.New(op, ez.ECONFLICT, msg, nil)

		default:
			return ez.New(op, ez.EINTERNAL, "Error making query to the database", err)
		}
	}

	return nil
}

// Query returns a list of models from the database that satisfy a Query, extra parameters
// in the Query allow for Limit and Offset
func (db *DB) Query(mList interface{}, model interfaces.Model, query []string) error {
	const op = "PG.DB.Query"

	var q string

	switch len(query) {
	case 1:
		q = fmt.Sprintf(`SELECT * FROM %s WHERE %s`, model.GetSchema().Name, query[0])
	case 2:
		q = fmt.Sprintf(`SELECT * FROM %s WHERE %s LIMIT %s`, model.GetSchema().Name, query[0], query[1])
	default:
		q = fmt.Sprintf(`SELECT * FROM %s WHERE %s LIMIT %s OFFSET %s`, model.GetSchema().Name, query[0], query[1], query[2])
	}

	result, err := db.pg.Query(mList, q, nil)
	if result.RowsReturned() == 0 {
		msg := fmt.Sprintf("Could not find any %s with query %s", model.GetSchema().Name, q)
		return ez.New(op, ez.ENOTFOUND, msg, nil)
	}

	if err != nil {
		fmt.Println("err", err)
		switch err.Error() {
		case ENOROWS:
			msg := fmt.Sprintf("Could not find a %s model with query %s", model.GetSchema().Name, query)
			return ez.New(op, ez.ENOTFOUND, msg, nil)
		default:
			return ez.New(op, ez.EINTERNAL, "Error making query to the database", err)
		}
	}

	return nil
}

// Insert adds a model into the database
func (db *DB) Insert(m interfaces.Model) error {
	const op = "PG.DB.Insert"

	err := db.pg.Insert(m)
	if err != nil {
		switch err.Error() {
		default:
			return ez.New(op, ez.EINTERNAL, "Error inserting to the database", err)
		}
	}

	return nil
}

// Update changes an existing model from the database
func (db *DB) Update(m interfaces.Model) error {
	const op = "PG.DB.Update"

	err := db.pg.Update(m)
	if err != nil {
		switch err.Error() {
		default:
			return ez.New(op, ez.EINTERNAL, "Error updating from the database", err)
		}
	}

	return nil
}

// Delete removes an existing model from the database
func (db *DB) Delete(m interfaces.Model) error {
	const op = "PG.DB.Delete"

	err := db.pg.Delete(m)
	if err != nil {
		switch err.Error() {
		default:
			return ez.New(op, ez.EINTERNAL, "Error deleting from the database", err)
		}
	}

	return nil
}

// CreateSchema creates the database tables if dropExisting is set to true it will drop the current schema
func (db *DB) CreateSchema(modelsList []interface{}, dropExisting bool) error {
	const op = "PG.DB.CreateSchema"
	for _, model := range modelsList {
		if dropExisting {
			err := db.DropTable(model)
			if err != nil {
				return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
			}
		}
		err := db.CreateTable(model)
		if err != nil {
			return ez.New(op, ez.ErrorCode(err), ez.ErrorMessage(err), err)
		}
	}
	return nil
}

// CreateTable creates a new table in the database
func (db *DB) CreateTable(model interface{}) error {
	const op = "PG.DB.CreateTable"

	err := db.pg.CreateTable(model, &orm.CreateTableOptions{
		Temp: false,
	})
	if err != nil {
		errorCode := err.Error()[0:12]
		if errorCode != ETABLEEXISTS {
			return ez.New(op, ez.EINTERNAL, "Could not create table", err)
		}
	}

	return nil
}

// DropTable deletes the existing tables
func (db *DB) DropTable(model interface{}) error {
	const op = "PG.DB.DropTable"

	err := db.pg.DropTable(model, &orm.DropTableOptions{})
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Could not drop table", err)
	}

	return nil
}
