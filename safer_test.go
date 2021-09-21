package safer

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Error("expected client to not be nil but it was")
	}
}

func TestClient_GetCompanyByDOTNumber(t *testing.T) {
	s := newTestServer()
	defer s.Close()

	c := &Client{
		scraper: scraper{
			companySnapshotURL: s.URL + "/snapshot",
			searchURL:          s.URL + "/search",
		},
	}
	snapshot, err := c.GetCompanyByDOTNumber("")
	if snapshot == nil {
		t.Error("snapshot returned nil")
	}
	if err != nil {
		t.Errorf("error expected nil but got %v", err)
	}

	snapshot, err = c.GetCompanyByMCMX("")
	if snapshot == nil {
		t.Error("snapshot returned nil")
	}
	if err != nil {
		t.Errorf("error expected nil but got %v", err)
	}

	results, err := c.SearchCompaniesByName("")
	if results == nil {
		t.Error("results returned nil")
	}
	if len(results) == 0 {
		t.Error("results length = 0")
	}
	if err != nil {
		t.Errorf("error expected nil but got %v", err)
	}
}

// benchmarks the time it takes for mapping a snapshot response.
// doesn't include any time spent waiting for response from server.
func BenchmarkClient_GetCompanyByDOTNumber(b *testing.B) {
	s := newTestServer()
	defer s.Close()

	c := &Client{
		scraper: scraper{
			companySnapshotURL: s.URL + "/snapshot",
			searchURL:          s.URL + "/search",
		},
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = c.GetCompanyByDOTNumber("")
	}
}
