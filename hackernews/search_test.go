package hackernews

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSearch(t *testing.T) {
	wanted := &SearchResults{
		Hits: []*Story{
			{
				Title:     "Sample",
				URL:       "a.com",
				CreatedAt: time.Unix(0, 0),
			},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.URL.String(), "postgres") {
			t.Errorf("Expected 'postgres' in url query params, received: %s", req.URL.String())
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(wanted)
	}))

	// Close the server when test finishes.
	defer server.Close()
	api := &APIClient{
		httpClient: server.Client(),
		apiURL:     server.URL,
	}
	res, err := api.Search(context.Background(), "postgres")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wanted, res) {
		t.Errorf("Wanted: %v, received %v", wanted, res)
	}
}

func TestSearchResults_String(t *testing.T) {
	res := &SearchResults{}
	if str := res.String(); str != "" {
		t.Errorf("Expected empty string for empty results, received %s", str)
	}
	res.Hits = []*Story{
		{
			Title:     "Sample",
			URL:       "a.com",
			CreatedAt: time.Unix(0, 0),
		},
		{
			// We write a case with a missing URL to ensure it is
			// excluded from the formatted string.
			Title:     "Other",
			CreatedAt: time.Unix(0, 0),
		},
	}
	var wanted strings.Builder
	wanted.WriteString("**Sample**\n")
	wanted.WriteString("Link: a.com\n")
	wanted.WriteString("Posted: Wed, Dec 31 at 18:00\n")
	wanted.WriteRune('\n')
	wanted.WriteString("**Other**\n")
	wanted.WriteString("Posted: Wed, Dec 31 at 18:00\n")
	wanted.WriteRune('\n')

	if str := res.String(); str != wanted.String() {
		t.Errorf("Wanted \n%s", wanted.String())
		t.Errorf("Received \n%s", str)
	}
}
