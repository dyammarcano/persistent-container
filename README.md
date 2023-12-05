[![Lint](https://github.com/dyammarcano/persistent-container/actions/workflows/lint.yaml/badge.svg)](https://github.com/dyammarcano/persistent-container/actions/workflows/lint.yaml)
# Data Store

## Under active development - Not ready for production

## TODO

- [x] Create data
- [x] Read data
- [x] Update data
- [x] Delete data
- [x] Search data
- [ ] Sort data
- [ ] Filter data
- [ ] Pagination
- [ ] Dark mode
- [ ] Responsive design
- [ ] Data validation
- [ ] Data persistence
- [ ] Data encryption
- [ ] Data compression
- [ ] Data backup
- [ ] Data restore
- [ ] Data export
- [ ] Data import
- [ ] Data synchronization
- [ ] Data sharing
- [ ] Data history
- [ ] Data versioning
- [ ] Data migration
- [ ] Data replication
- [ ] Web UI application (Vue 3)
    - [ ] Login page
    - [ ] Metrics page
    - [ ] Create data
    - [x] Read data
    - [ ] Update data
    - [ ] Delete data
    - [ ] List data
- [ ] Standalone application (Electron)

### Simple container struct to wrap any struct and persist it to disk as CBOR encoded data.

## Prerequisites

- Node.js 14.0.0 or higher
- NPM 6.14.0 or higher
- Golang 1.16.0 or higher

## How to use

1. Clone this repository
2. Run `go generate ./...`
3. Run `go build -o store .` or `go run .` to start the server
4. Open `http://localhost:8090` in your browser
5. Enjoy!

## Under the hood

This package uses [bbolt](https://github.com/etcd-io/bbolt) to persist data to disk,
and [cbor](github.com/fxamacker/cbor/v2) to encode and decode data.

## Disclaimer

This package is not intended to be used in production, it is just a simple wrapper to persist data to disk, is thread
safe, and it is not intended to be used in a concurrent environment.

## Usage

```go
package main

import (
	"context"
	"dataStore/internal/container"
	"dataStore/internal/store"
	"fmt"
)

type MyStruct struct {
	Name     string
	LastName string
	Age      int
}

func main() {
	db, err := store.NewPersistence(context.TODO(), "persiste.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create a new instance of MyStruct
	myStruct := MyStruct{
		Name:     "John",
		LastName: "Wick",
		Age:      42,
	}

	newContainer := container.NewContainer[MyStruct](myStruct, "testBucket", "testKey", db)

	myStruct.Age = 43
	myStruct.Name = "John2"
	myStruct.LastName = "Wick2"

	// any changes made to the struct will be persisted to disk whitout the need to worry about it.

	fmt.Printf("age: %d\n", newContainer.GetObject().Age)
	fmt.Printf("name: %s\n", newContainer.GetObject().Name)
	fmt.Printf("lastName: %s\n", newContainer.GetObject().LastName)
}

```
