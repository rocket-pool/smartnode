package assets

import "testing"

func TestLogo(t *testing.T) {
	logo := Logo()
	if logo == "" {
		t.Fatal("Logo wasn't embedded")
	}
}

func TestVersionString(t *testing.T) {
	// Reset v singleton
	v = nil
	vers := RocketPoolVersion()
	if vers == "" {
		t.Fatal("Version string wasn't embedded")
	}

	if v == nil {
		t.Fatalf("v should be initialized")
	}

	oldV := v

	// Make sure subsequent calls to RocketPoolVersion are amortized
	RocketPoolVersion()
	if oldV != v {
		t.Fatalf("v was reallocated")
	}
}

func shouldPanic(t *testing.T) {
	if r := recover(); r == nil {
		t.Fatal("should have panicked!")
	} else {
		t.Log(r)
	}
}

func TestInvalidJsonVersionPanics(t *testing.T) {
	// Reset v singleton
	v = nil

	// Save the version JSON so we can restore it
	oldVersionJSON := versionJSON
	defer func() {
		versionJSON = oldVersionJSON
	}()
	// We want to panic when version can't be parsed.
	defer shouldPanic(t)

	versionJSON = []byte("this is not valid json")
	_ = RocketPoolVersion()

}

func TestEmptyVersionPanics(t *testing.T) {
	// Reset v singleton
	v = nil

	// Save the version JSON so we can restore it
	oldVersionJSON := versionJSON
	defer func() {
		versionJSON = oldVersionJSON
	}()
	// We want to panic when version is empty but json is valid
	defer shouldPanic(t)

	versionJSON = []byte("{}")
	_ = RocketPoolVersion()

}
