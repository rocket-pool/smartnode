package node

import (
    "time"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/database"
)


// Config
const CHECKIN_INTERVAL string = "15s"


// Start node checkin process
func StartCheckinProcess(c *cli.Context, errors chan error, fatalErrors chan error) {

    // Initialise database
    db := database.NewDatabase(c.GlobalString("database"))
    if err := db.Open(); err != nil {
        fatalErrors <- err
        return
    }

    // Get last checkin time
    lastCheckinTime := new(int64)
    if err := db.Get("node.checkin.latest", lastCheckinTime); err != nil {
        *lastCheckinTime = 0
    }
    if err := db.Close(); err != nil {
        errors <- err
    }

    // Initialise checkin interval
    checkinInterval, _ := time.ParseDuration(CHECKIN_INTERVAL)

    // Get next checkin time
    var nextCheckinTime time.Time
    if *lastCheckinTime == 0 {
        nextCheckinTime = time.Now()
    } else {
        nextCheckinTime = time.Unix(*lastCheckinTime, 0).Add(checkinInterval)
    }

    // Get duration until next checkin
    nextCheckinDuration := time.Until(nextCheckinTime)

    // Initialise checkin timer
    checkinTimer := time.NewTimer(nextCheckinDuration)
    for _ = range checkinTimer.C {
        checkin(db, checkinTimer, &checkinInterval, errors)
    }

}


// Perform node checkin
func checkin(db *database.Database, checkinTimer *time.Timer, checkinInterval *time.Duration, errors chan error) {



    // Set last checkin time
    if err := db.Open(); err != nil {
        errors <- err
    } else {
        if err = db.Put("node.checkin.latest", time.Now().Unix()); err != nil {
            errors <- err
        }
        if err = db.Close(); err != nil {
            errors <- err
        }
    }

    // Reset timer for next checkin
    checkinTimer.Reset(*checkinInterval)

}

