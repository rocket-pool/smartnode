package database

import (
    "errors"
    "sync"

    "github.com/rapidloop/skv"
)


// Database
type Database struct {
    store *skv.KVStore
    path string
    lock sync.Mutex
    atomicLock sync.Mutex
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

    // Lock for modification
    db.lock.Lock()
    defer db.lock.Unlock()

    // Check not open
    if db.store != nil { return errors.New("Database is already open") }

    // Initialise storage
    if store, err := skv.Open(db.path); err != nil {
        return errors.New("Error opening database: " + err.Error())
    } else {
        db.store = store
    }

    // Return
    return nil

}


/**
 * Close database
 */
func (db *Database) Close() error {

    // Lock for modification
    db.lock.Lock()
    defer db.lock.Unlock()

    // Check not closed
    if db.store == nil { return errors.New("Database is already closed") }

    // Close storage
    if err := db.store.Close(); err != nil {
        return errors.New("Error closing database: " + err.Error())
    } else {
        db.store = nil
    }

    // Return
    return nil

}


/**
 * Read a value from the database
 */
func (db *Database) Get(key string, value interface{}) error {

    // Lock during read
    db.lock.Lock()
    defer db.lock.Unlock()

    // Check not closed
    if db.store == nil { return errors.New("Cannot read from closed database") }

    // Read and return; ignore errors as key may not exist
    _ = db.store.Get(key, value)
    return nil

}


/**
 * Write a value to the database
 */
func (db *Database) Put(key string, value interface{}) error {

    // Lock for modification
    db.lock.Lock()
    defer db.lock.Unlock()

    // Check not closed
    if db.store == nil { return errors.New("Cannot write to closed database") }

    // Write and return
    if err := db.store.Put(key, value); err != nil {
        return errors.New("Error writing value to database: " + err.Error())
    }
    return nil

}


/**
 * Remove a value from the database
 */
func (db *Database) Delete(key string) error {

    // Lock for modification
    db.lock.Lock()
    defer db.lock.Unlock()

    // Check not closed
    if db.store == nil { return errors.New("Cannot remove from closed database") }

    // Remove and return
    if err := db.store.Delete(key); err != nil {
        return errors.New("Error removing value from database: " + err.Error())
    }
    return nil

}


/**
 * Read a value from the database (atomic operation)
 */
func (db *Database) GetAtomic(key string, value interface{}) error {

    // Lock during read
    db.atomicLock.Lock()
    defer db.atomicLock.Unlock()

    // Open database
    if err := db.Open(); err != nil { return err }
    defer db.Close()

    // Read value and return
    return db.Get(key, value)

}


/**
 * Write a value to the database (atomic operation)
 */
func (db *Database) PutAtomic(key string, value interface{}) error {

    // Lock for modification
    db.atomicLock.Lock()
    defer db.atomicLock.Unlock()

    // Open database
    if err := db.Open(); err != nil { return err }
    defer db.Close()

    // Write value and return
    return db.Put(key, value)

}


/**
 * Remove a value from the database (atomic operation)
 */
func (db *Database) DeleteAtomic(key string) error {

    // Lock for modification
    db.atomicLock.Lock()
    defer db.atomicLock.Unlock()

    // Open database
    if err := db.Open(); err != nil { return err }
    defer db.Close()

    // Remove value and return
    return db.Delete(key)

}

