package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestFindNextSubmission(t *testing.T) {
	latestBlockTimestamp := 1713420045 // Same day
	referenceTimestamp := 1713420000   // 06:00AM
	intervalInSeconds := 86400
	result, err := FindNextSubmissionTimestamp(int64(latestBlockTimestamp), int64(referenceTimestamp), int64(intervalInSeconds))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", time.Unix(result, 0).String())
	if result != 1713420000 {
		t.Fatalf("Wrong result")
	}
	latestBlockTimestamp = 1713510000
	result, err = FindNextSubmissionTimestamp(int64(latestBlockTimestamp), int64(referenceTimestamp), int64(intervalInSeconds))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1713506400 {
		t.Fatalf("Wrong result")
	}
	// Test zero values
	_, err = FindNextSubmissionTimestamp(int64(latestBlockTimestamp), int64(referenceTimestamp), 0)
	if err == nil {
		t.Fatalf("Should have errored after using 0 for the interval")
	}
	_, err = FindNextSubmissionTimestamp(int64(latestBlockTimestamp), 0, int64(intervalInSeconds))
	if err == nil {
		t.Fatalf("Should have errored after using 0 for the reference timestamp")
	}
	_, err = FindNextSubmissionTimestamp(0, int64(referenceTimestamp), int64(intervalInSeconds))
	if err == nil {
		t.Fatalf("Should have errored after using 0 for the latest block timestamp")
	}

	// Test reference timestamp in the future
	_, err = FindNextSubmissionTimestamp(int64(latestBlockTimestamp), int64(latestBlockTimestamp+86400), int64(intervalInSeconds))
	if err == nil {
		t.Fatalf("Should have error when using a reference date in the future")
	}
}

func FindNextSubmissionTimestamp(latestBlockTimestamp int64, referenceTimestamp int64, submissionIntervalInSeconds int64) (int64, error) {
	if latestBlockTimestamp == 0 || referenceTimestamp == 0 || submissionIntervalInSeconds == 0 {
		return 0, fmt.Errorf("FindNextSubmissionTimestamp can't use zero values")
	}

	// Calculate the difference between latestBlockTime and the reference timestamp
	timeDifference := latestBlockTimestamp - referenceTimestamp
	if timeDifference < 0 {
		return 0, fmt.Errorf("FindNextSubmissionTimestamp referenceTimestamp in the future")
	}

	// Calculate the remainder to find out how far off from a multiple of the interval the current time is
	remainder := timeDifference % submissionIntervalInSeconds

	// Subtract the remainder from current time to find the first multiple of the interval in the past
	submissionTimeRef := latestBlockTimestamp - remainder
	return submissionTimeRef, nil
}
