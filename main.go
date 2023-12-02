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
	db, err := store.NewStore(context.TODO(), "persiste.db")
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
