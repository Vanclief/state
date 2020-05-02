# PGDB
PostgresSQL as a database while using go-pg as an ORM to manage it.

## Usage

Create a new database:
```
db, err := pgdb.New("localhost:5432", "vanclief", "", "postgres")
if err != nil {
    panic("Could not create the database")
}
```

Query:
```
// First argument is an array of the Model you are attempting to obtain
// Second argument is an empty instance of the Model you are attempting to obtain
// Third argument is the SQL Query, which is inserted after a "WHERE" statement
// Optional: Fourth argument is the Limit of Rows to return
// Optional: Fifth argument is the Offset
pgdb.Query(&res, &user.User{}, `name = 'Franco' ORDER BY email DESC`, []string{"10", "5"})
``` 
