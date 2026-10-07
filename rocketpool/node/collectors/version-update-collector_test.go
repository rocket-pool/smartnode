package collectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{
			name:    "latest version is newer",
			current: "1.20.2",
			latest:  "v1.20.3",
			want:    true,
		},
		{
			name:    "same version is not newer",
			current: "1.20.2",
			latest:  "v1.20.2",
			want:    false,
		},
		{
			name:    "older latest version is not newer",
			current: "1.20.2",
			latest:  "v1.20.1",
			want:    false,
		},
		{
			name:    "dev version is newer than latest",
			current: "v1.20.3-dev",
			latest:  "v1.20.2",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isNewerVersion(tt.current, tt.latest)
			if err != nil {
				t.Fatalf("isNewerVersion returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("isNewerVersion(%q, %q) = %t, want %t", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestCheckIfDueCachesLatestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"tag_name":"v1.20.3"}`))
		if err != nil {
			t.Fatalf("error writing response: %v", err)
		}
	}))
	defer server.Close()

	collector := NewVersionUpdateCollector(nil)
	collector.current = "1.20.2"
	collector.latestURL = server.URL
	collector.client = server.Client()

	collector.checkIfDue(context.Background())

	if collector.updateAvailable != 1 {
		t.Fatalf("updateAvailable = %f, want 1", collector.updateAvailable)
	}
	if collector.latestVersion != "v1.20.3" {
		t.Fatalf("latestVersion = %q, want %q", collector.latestVersion, "v1.20.3")
	}
}

func TestCheckIfDueHitsGitHubAtMostHourly(t *testing.T) {
	var hits int
	tag := "v1.20.3"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, err := w.Write([]byte(`{"tag_name":"` + tag + `"}`))
		if err != nil {
			t.Errorf("error writing response: %v", err)
		}
	}))
	defer server.Close()

	collector := NewVersionUpdateCollector(nil)
	collector.current = "1.20.2"
	collector.latestURL = server.URL
	collector.client = server.Client()

	collector.checkIfDue(context.Background())
	collector.checkIfDue(context.Background())
	if hits != 1 {
		t.Fatalf("hits = %d, want 1", hits)
	}
	if collector.latestVersion != "v1.20.3" {
		t.Fatalf("latestVersion = %q, want %q", collector.latestVersion, "v1.20.3")
	}

	tag = "v1.20.4"
	collector.lastChecked = time.Now().Add(-versionCheckInterval - time.Second)
	collector.checkIfDue(context.Background())
	if hits != 2 {
		t.Fatalf("hits = %d, want 2", hits)
	}
	if collector.latestVersion != "v1.20.4" {
		t.Fatalf("latestVersion = %q, want %q", collector.latestVersion, "v1.20.4")
	}
	if collector.updateAvailable != 1 {
		t.Fatalf("updateAvailable = %f, want 1", collector.updateAvailable)
	}
}

func TestCheckIfDueBacksOffAfterGitHubError(t *testing.T) {
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	collector := NewVersionUpdateCollector(nil)
	collector.current = "1.20.2"
	collector.latestURL = server.URL
	collector.client = server.Client()

	collector.checkIfDue(context.Background())
	collector.checkIfDue(context.Background())
	if hits != 1 {
		t.Fatalf("hits = %d, want 1", hits)
	}
	if collector.latestVersion != "" {
		t.Fatalf("latestVersion = %q, want empty", collector.latestVersion)
	}
	if collector.updateAvailable != 0 {
		t.Fatalf("updateAvailable = %f, want 0", collector.updateAvailable)
	}
}
