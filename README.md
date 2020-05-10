# state
Minimalistic package for handling a Go application State with a Database and Cache

## Features

* Use a single API to manage your application Database and Cache 

* Git like interface where you stage changes, commit them and revert them.

* Easily extended with other caches or databases, just need to implement a single interface

## Usage

### State API

**Import the library:**

`import "github.com/vanclief/state/manager"`

**Creating a new State:**
```
state, err := manager.New(db, cache) // Using both a Database and Cache
state, err := manager.New(db, nil) // Just using Database 
state, err := manager.New(nil, cache) // Just using Cache 
```

**Staging changes:**
```
i := user.New("1", "Franco", "email@francovalencia.com") // Your model here

state.Stage(i, "insert") // This stages User "i" to be inserted
state.Stage(u, "update") // This stages User "u" to be updated
state.Stage(d, "delete") // This stages User "d" to be deleted

```

**Commit changes:**
```
err := state.Commit() // Applies all staged changes
if err != nil {
    // Handle that one or more changes where not applied
}
```

**Rollback applied insertions:**
```
err := state.Rollback() // Reverts applied "insert" changes
if err != nil {
    // Handle that one or more inserts could not be reverted 
}
```

**Clear staged changes:**
```
state.Clear() // Clears all staged changes
```

**Display staged changes:**
```
changes := state.Status() 
for _, change := range changes {
    fmt.Println("Model:", change.model, "OP:", change.op, "Status:", change.status, "Error:", change.err)
	}
}
```

**Display applied changes:**
```
applied := state.Applied() 
for _, change := range applied {
    fmt.Println("Model:", change.model, "OP:", change.op, "Status:", change.status, "Error:", change.err)
	}
}
```

**Get a model using its ID:**
```
u := &user.User{}
state.Get(u, "1")
fmt.Println(u) // {"1", "Franco", "email@francovalencia.com"}
```

**Query the database for a single model:**
```
u := &user.User{}
state.QueryOne(u, `email = 'email@francovalencia.com'`)
fmt.Println(u) // {"1", "Franco", "email@francovalencia.com"}
```
*Query format will depend of your database*

**Query the database for multiple models:**
```
users := []user.User{}
state.QueryOne(users, , `name = 'John'`)
fmt.Println(users[0]) // {"2", "John", "john@wick.com"}
fmt.Println(users[1]) // {"3", "John", "john@cena.com"}
```
*Query format will depend of your database*

### Models 
Your models should implement the interfaces.Model interface, you can check 
`examplemodels` to see how this is done.

### Database Interface
Your database should implement the `interfaces.Database` interface, check the folder `databases` for examples.

### Cache Interface
Your cache should implement the `interfaces.Cache` interface, check the folder `caches` for examples.

## Contributions
Feel free to open a PR or an Issue.