package main

import (
	"context"
	"dataStore/internal/monitoring"
	"dataStore/internal/store"
	"github.com/caarlos0/log"
	"path/filepath"
)

type MyStruct struct {
	Name     string
	LastName string
	Age      int
}

func main() {
	ctx := context.TODO()
	databasePath, err := filepath.Abs("./dataStore.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := store.NewStore(ctx, databasePath)
	if err != nil {
		log.Fatal(err.Error())
	}

	newMonitoring := monitoring.NewMonitoring(ctx, db, 8080)

	callback := func(err error) {
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	newMonitoring.StartServer(callback)

	//db, err := store.NewStore(context.TODO(), "persiste.db")
	//if err != nil {
	//	panic(err)
	//}
	//defer db.Close()
	//
	//// Create a new instance of MyStruct
	//myStruct := MyStruct{
	//	Name:     "John",
	//	LastName: "Wick",
	//	Age:      42,
	//}
	//
	//newContainer := container.NewContainer[MyStruct](myStruct, "testBucket", "testKey", db)
	//
	//myStruct.Age = 43
	//myStruct.Name = "John2"
	//myStruct.LastName = "Wick2"
	//
	//// any changes made to the struct will be persisted to disk whitout the need to worry about it.
	//
	//fmt.Printf("age: %d\n", newContainer.GetObject().Age)
	//fmt.Printf("name: %s\n", newContainer.GetObject().Name)
	//fmt.Printf("lastName: %s\n", newContainer.GetObject().LastName)
}
