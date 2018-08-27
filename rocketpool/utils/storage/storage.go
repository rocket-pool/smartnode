package storage

import (
    "github.com/rapidloop/skv"
)


// Config
var storagePath = "/Users/moles/.rocketpool/cli.db";


// Open storage
func Open() (*skv.KVStore, error) {
    return skv.Open(storagePath)
}

