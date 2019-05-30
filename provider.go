package rovers

import (
	"context"
	"net/http"
	"time"

	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"

	"github.com/google/go-github/v25/github"
	"golang.org/x/oauth2"
)

var (
	// ErrEndpointsNotFound is the returned error when couldn't find
	// endpoints for a certain repository.
	ErrEndpointsNotFound = errors.NewKind("endpoinds not found for %s")
)

// Provider will retrieve the information for all the repositories for the given
// github organizations.
type Provider struct {
	persist PersistMentionFn
	orgs    OrganizationIterator
	client  *github.Client
}

// NewProvider builds a new Provider
func NewProvider(
	persist PersistMentionFn,
	orgs OrganizationIterator,
	token string,
) *Provider {
	return &Provider{
		persist: persist,
		orgs:    orgs,
		client:  newGithubClient(token),
	}
}

const (
	httpTimeout    = 30 * time.Second
	resultsPerPage = 100
)

func newGithubClient(token string) *github.Client {
	var client *http.Client
	if token == "" {
		client = &http.Client{}
	} else {
		client = oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		)
	}

	client.Timeout = httpTimeout
	return github.NewClient(client)
}

// Start starts the provider.
func (p *Provider) Start() error {
	return p.orgs.ForEach(func(org string) error {
		return p.requestRepos(org)
	})
}

func (p *Provider) requestRepos(org string) error {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: resultsPerPage},
	}

	for {
		logger := log.New(log.Fields{"org": org, "page": opt.Page})
		repos, res, err := p.client.Repositories.ListByOrg(
			context.Background(),
			org,
			opt,
		)

		if err != nil {
			if _, ok := err.(*github.RateLimitError); !ok {
				logger.Errorf(err, "failing retrieving repositories")
				return err
			}

			wait := timeToRetry(res)
			logger.Infof("rate limit reached, waiting %s to retry", wait.String())
			time.Sleep(wait)
			continue
		}

		for _, r := range repos {
			logger := logger.With(log.Fields{"repository": r.GetFullName()})
			logger.Infof("processing data")
			endpoints, err := getEndpoints(r)
			if err != nil {
				logger.Errorf(err, "failing processing response")
				continue
			}

			if err := p.persist(NewMention(endpoints, r.GetFork())); err != nil {
				logger.Errorf(err, "failing persiting data")
			}

			logger.Infof("data persisted")
		}

		if res.NextPage == 0 {
			break
		}

		opt.Page = res.NextPage
	}

	return nil
}

func getEndpoints(r *github.Repository) ([]string, error) {
	var endpoints []string
	getURLs := []func() string{
		r.GetGitURL,
		r.GetSSHURL,
		r.GetHTMLURL,
	}

	for _, getURL := range getURLs {
		ep := getURL()
		if ep != "" {
			endpoints = append(endpoints, ep)
		}
	}

	if len(endpoints) < 1 {
		return nil, ErrEndpointsNotFound.New(r.GetFullName())
	}

	return endpoints, nil
}

func timeToRetry(res *github.Response) time.Duration {
	now := time.Now().UTC().Unix()
	resetTime := res.Rate.Reset.UTC().Unix()
	timeToReset := time.Duration(resetTime-now) * time.Second
	remaining := res.Rate.Remaining
	if timeToReset < 0 || timeToReset > 1*time.Hour {
		// If this happens, the system clock is probably wrong, so we assume we
		// are at the beginning of the window and consider only total requests
		// per hour.
		timeToReset = 1 * time.Hour
		remaining = res.Rate.Limit
	}

	return timeToReset / time.Duration(remaining+1)
}
