package database

import (
    "io/ioutil"
    "testing"
)


// Test database functionality
func TestDatabase(t *testing.T) {

    // Create temporary database path
    dbPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }
    dbPath += "/database"

    // Initialise database
    db := NewDatabase(dbPath)

    // Initialise data pointer
    foobar := new(int64)

    // Check database open status
    if open := db.IsOpen(); open {
        t.Errorf("Incorrect database open status: expected %t, got %t", false, open)
    }

    // Attempt to close database while closed
    if err := db.Close(); err == nil {
        t.Error("Database Close() method should return error when closed")
    }

    // Attempt to write to database while closed
    if err := db.Put("foo.bar", 3); err == nil {
        t.Error("Database Put() method should return error when closed")
    }

    // Attempt to read from database while closed
    if err := db.Get("foo.bar", foobar); err == nil {
        t.Error("Database Get() method should return error when closed")
    }

    // Attempt to remove from database while closed
    if err := db.Delete("foo.bar"); err == nil {
        t.Error("Database Delete() method should return error when closed")
    }

    // Atomic write to database
    if err := db.PutAtomic("foo.bar", 5); err != nil { t.Error(err) }

    // Atomic read from database
    if err := db.GetAtomic("foo.bar", foobar); err != nil {
        t.Error(err)
    } else if *foobar != 5 {
        t.Errorf("Incorrect value read from database: expected %d, got %d", 5, *foobar)
    }

    // Atomic remove from database
    if err := db.DeleteAtomic("foo.bar"); err != nil { t.Error(err) }

    // Check database value
    *foobar = 0
    if err := db.GetAtomic("foo.bar", foobar); err != nil {
        t.Error(err)
    } else if *foobar != 0 {
        t.Errorf("Incorrect value read from database: expected %d, got %d", 0, *foobar)
    }

    // Open database
    if err := db.Open(); err != nil { t.Error(err) }

    // Check database open status
    if open := db.IsOpen(); !open {
        t.Errorf("Incorrect database open status: expected %t, got %t", true, open)
    }

    // Attempt to open database while open
    if err := db.Open(); err == nil {
        t.Error("Database Open() method should return error when open")
    }

    // Write to database
    if err := db.Put("foo.bar", 7); err != nil { t.Error(err) }

    // Read from database
    if err := db.Get("foo.bar", foobar); err != nil {
        t.Error(err)
    } else if *foobar != 7 {
        t.Errorf("Incorrect value read from database: expected %d, got %d", 7, *foobar)
    }

    // Remove from database
    if err := db.Delete("foo.bar"); err != nil { t.Error(err) }

    // Check database value
    *foobar = 0
    if err := db.Get("foo.bar", foobar); err != nil {
        t.Error(err)
    } else if *foobar != 0 {
        t.Errorf("Incorrect value read from database: expected %d, got %d", 0, *foobar)
    }

    // Attempt atomic write to database while open
    if err := db.PutAtomic("foo.bar", 9); err == nil {
        t.Error("Database PutAtomic() method should return error when open")
    }

    // Attempt atomic read from database while open
    if err := db.GetAtomic("foo.bar", foobar); err == nil {
        t.Error("Database GetAtomic() method should return error when open")
    }

    // Attempt atomic remove from database while open
    if err := db.DeleteAtomic("foo.bar"); err == nil {
        t.Error("Database DeleteAtomic() method should return error when open")
    }

    // Close database
    if err := db.Close(); err != nil { t.Error(err) }

}

