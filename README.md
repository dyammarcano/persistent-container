# Data Store 

### Simple container struct to wrap any struct and persist it to disk as CBOR encoded data.

## Under the hood

This package uses [bbolt](https://github.com/etcd-io/bbolt) to persist data to disk, and [cbor](github.com/fxamacker/cbor/v2) to encode and decode data.


## Disclaimer

This package is not intended to be used in production, it is just a simple wrapper to persist data to disk, is thread safe, and it is not intended to be used in a concurrent environment.

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