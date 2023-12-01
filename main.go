package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	badger "github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"time"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(wd)

	storePath := filepath.Join(wd, "badger")

	// Open Badger DB
	db, err := badger.Open(badger.DefaultOptions(storePath))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
		again:
			if err := db.RunValueLogGC(0.7); err == nil {
				goto again
			}
		}
	}()

	updates := make(map[string]string)
	txn := db.NewTransaction(true)
	for k, v := range updates {
		if err := txn.Set([]byte(k), []byte(v)); errors.Is(err, badger.ErrTxnTooBig) {
			_ = txn.Commit()
			txn = db.NewTransaction(true)
			_ = txn.Set([]byte(k), []byte(v))
		}
	}
	_ = txn.Commit()

	// Create a new user
	newUser := User{
		ID:    uuid.NewString(),
		Name:  "John Doe",
		Email: "johndoe@example.com",
	}

	// Encode user data to JSON
	userData, err := json.Marshal(newUser)
	if err != nil {
		log.Fatal(err)
	}

	// Generate a unique key for the user
	key := db.NewSequence().Next()

	// Store the encoded user data in Badger DB
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), userData)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the stored user data
	var retrievedUser User
	err = db.View(func(txn *badger.Txn) error {
		value, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		defer value.Close()

		err = json.Unmarshal(value.Value(), &retrievedUser)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved user: %+v\n", retrievedUser)

	// Update the retrieved user's email
	retrievedUser.Email = "johndoe@newemail.com"

	// Encode the updated user data to JSON
	updatedUserData, err := json.Marshal(retrievedUser)
	if err != nil {
		log.Fatal(err)
	}

	// Update the stored user data with the encoded updated user data
	err = db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte(key), updatedUserData); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	//err := db.Update(func(txn *badger.Txn) error {
	//	e := badger.NewEntry([]byte("answer"), []byte("42")).WithMeta(byte(1)).WithTTL(time.Hour)
	//	err := txn.SetEntry(e)
	//	return err
	//})
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Retrieve the updated user data
	var updatedRetrievedUser User
	err = db.View(func(txn *badger.Txn) error {
		value, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		defer value.Close()

		err = json.Unmarshal(value.Value(), &updatedRetrievedUser)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated retrieved user: %+v\n", updatedRetrievedUser)

	// Delete the stored user data
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func uint64ToBytes(i uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[:]
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}
