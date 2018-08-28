package rpip

import (
    "strconv"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/storage"
)


// Commit a vote on a proposal
func commitVote(proposalId uint64, vote string) error {

    // Commit vote
    // TODO: implement

    // Open storage
    store, err := storage.Open()
    if err != nil {
        return err
    }
    defer store.Close()

    // Store vote
    return store.Put("rpip.votes.committed." + strconv.FormatUint(proposalId, 10), vote)

}


// Check a vote on a proposal
func checkVote(proposalId uint64) (string, error) {

    // Open storage
    store, err := storage.Open()
    if err != nil {
        return "", err
    }
    defer store.Close()

    // Get stored vote
    var vote string = ""
    store.Get("rpip.votes.committed." + strconv.FormatUint(proposalId, 10), &vote)
    return vote, nil

}


// Reveal a vote on a proposal
func revealVote(proposalId uint64, vote string) error {

    // Reveal vote
    // TODO: implement
    return nil

}

