package fullhunt

import (
	"context"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"github.com/projectdiscovery/subfinder/v2/pkg/subscraping"
)

//fullhunt response
type fullHuntResponse struct {
	Hosts   []string `json:"hosts"`
	Message string   `json:"message"`
	Status  int      `json:"status"`
}

// Source is the passive scraping agent
type Source struct {
	apiKeys []string
}

func (s *Source) Run(ctx context.Context, domain string, session *subscraping.Session) <-chan subscraping.Result {
	results := make(chan subscraping.Result)

	go func() {
		defer close(results)

		randomApiKey := subscraping.PickRandom(s.apiKeys, s.Name())
		if randomApiKey == "" {
			return
		}

		resp, err := session.Get(ctx, fmt.Sprintf("https://fullhunt.io/api/v1/domain/%s/subdomains", domain), "", map[string]string{"X-API-KEY": randomApiKey})
		if err != nil {
			results <- subscraping.Result{Source: s.Name(), Type: subscraping.Error, Error: err}
			session.DiscardHTTPResponse(resp)
			return
		}

		var response fullHuntResponse
		err = jsoniter.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			results <- subscraping.Result{Source: s.Name(), Type: subscraping.Error, Error: err}
			resp.Body.Close()
			return
		}
		resp.Body.Close()
		for _, record := range response.Hosts {
			results <- subscraping.Result{Source: s.Name(), Type: subscraping.Subdomain, Value: record}
		}
	}()
	return results
}

// Name returns the name of the source
func (s *Source) Name() string {
	return "fullhunt"
}

func (s *Source) IsDefault() bool {
	return true
}

func (s *Source) HasRecursiveSupport() bool {
	return false
}

func (s *Source) NeedsKey() bool {
	return true
}

func (s *Source) AddApiKeys(keys []string) {
	s.apiKeys = keys
}
