package storage

import (
    "github.com/rapidloop/skv"
)


// Config
var storagePath = "/var/tmp/rocketpool-cli.db"


// Open storage
func Open() (*skv.KVStore, error) {
    return skv.Open(storagePath)
}

