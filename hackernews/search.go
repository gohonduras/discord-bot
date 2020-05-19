// Package hackernews provides an API client for searching
// its website through its public, JSON REST endpoint.
package hackernews

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	baseAPIPath    = "https://hn.algolia.com/api/v1/search_by_date"
	queryFormat    = `?query=%s&tags=story`
	dateTimeFormat = "Mon, Jan _2 at 15:04"
	maxMessageSize = 2000
)

var log = logrus.WithField("prefix", "hackernews")

// APIClient for the hackernews website by Y-Combinator. Allows
// for searching its API through JSON HTTP requests.
// TODO: Make configurable, being able to query by story, comments,
// or more based on the hacker news API.
type APIClient struct {
	httpClient *http.Client
	apiURL     string
}

// SearchResults of hackernews queries.
type SearchResults struct {
	Hits []*Story `json:"hits"`
}

// Story defines an individual hacker news search
// result item from the website's JSON REST API.
type Story struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// NewAPIClient instantiates a new hackernews search client.
func NewAPIClient() *APIClient {
	client := &http.Client{
		Timeout: time.Second * 1, // Define a 1 second timeout for queries.
	}
	return &APIClient{
		httpClient: client,
		apiURL:     baseAPIPath,
	}
}

// Search the hacker news website with a given query.
func (a *APIClient) Search(ctx context.Context, query string) (*SearchResults, error) {
	queryParams := fmt.Sprintf(queryFormat, query)
	url := a.apiURL + queryParams
	req, err := http.NewRequest(http.MethodGet, url, nil /* Nil request body */)
	if err != nil {
		return nil, errors.Wrapf(err, "could initialize search api request: %s", query)
	}

	res, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "could not search query: %s", query)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Errorf("Could not close response body: %v", err)
		}
	}()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse response body")
	}

	searchRes := &SearchResults{}
	if err := json.Unmarshal(body, searchRes); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal search results")
	}
	if searchRes == nil {
		return nil, errors.New("empty response")
	}
	return searchRes, nil
}

// String allows for pretty printing of the search results for
// representation within a discord message.
func (res *SearchResults) String() string {
	if res == nil {
		return ""
	}
	var b strings.Builder
	for _, item := range res.Hits {
		b.WriteString(fmt.Sprintf("**%s**\n", item.Title))
		if item.URL != "" {
			b.WriteString(fmt.Sprintf("Link: %s\n", item.URL))
		}
		createdAt := item.CreatedAt.Format(dateTimeFormat)
		b.WriteString(fmt.Sprintf("Posted: %v\n", createdAt))
		b.WriteRune('\n')
	}
	str := b.String()
	if len(str) >= maxMessageSize {
		return str[:maxMessageSize]
	}
	return b.String()
}
