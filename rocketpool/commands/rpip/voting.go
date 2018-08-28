package rpip

import (
    "strconv"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/storage"
)


// Commit a vote on a proposal
func CommitVote(proposalId uint64, vote string) error {

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
func CheckVote(proposalId uint64) (string, error) {

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
func RevealVote(proposalId uint64, vote string) error {

    // Reveal vote
    // TODO: implement
    return nil

}

