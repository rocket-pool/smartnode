package database

import (
    "errors"

    "github.com/rapidloop/skv"
)


// Database
type Database struct {
    store *skv.KVStore
    path string
}


/**
 * Create database
 */
func NewDatabase(path string) *Database {
    return &Database{
        path: path,
    }
}


/**
 * Get database open status
 */
func (db *Database) IsOpen() bool {
    return (db.store != nil)
}


/**
 * Open database (if closed)
 */
func (db *Database) Open() error {

    // Check not open
    if db.store != nil {
        return errors.New("Database is already open")
    }

    // Initialise storage
    store, err := skv.Open(db.path)
    if err != nil {
        return errors.New("Error opening database: " + err.Error())
    }

    // Set store and return
    db.store = store
    return nil

}


/**
 * Close database
 */
func (db *Database) Close() error {

    // Check not closed
    if db.store == nil {
        return errors.New("Database is already closed")
    }

    // Close storage
    err := db.store.Close()
    if err != nil {
        return errors.New("Error closing database: " + err.Error())
    }

    // Unset store and return
    db.store = nil
    return nil

}


/**
 * Read a value from the database
 */
func (db *Database) Get(key string, value interface{}) error {

    // Check not closed
    if db.store == nil {
        return errors.New("Cannot read from closed database")
    }

    // Read and return
    err := db.store.Get(key, value)
    if err != nil {
        return errors.New("Error reading value from database: " + err.Error())
    }
    return nil

}


/**
 * Write a value to the database
 */
func (db *Database) Put(key string, value interface{}) error {

    // Check not closed
    if db.store == nil {
        return errors.New("Cannot write to closed database")
    }

    // Write and return
    err := db.store.Put(key, value)
    if err != nil {
        return errors.New("Error writing value to database: " + err.Error())
    }
    return nil

}


/**
 * Remove a value from the database
 */
func (db *Database) Delete(key string) error {

    // Check not closed
    if db.store == nil {
        return errors.New("Cannot remove from closed database")
    }

    // Remove and return
    err := db.store.Delete(key)
    if err != nil {
        return errors.New("Error removing value from database: " + err.Error())
    }
    return nil

}

