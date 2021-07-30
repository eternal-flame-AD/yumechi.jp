package comment

import (
	"context"
	"time"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

var githubClient *github.Client

//go:generate /bin/sh generate.sh
func init() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	tc.Timeout = 5 * time.Second
	githubClient = github.NewClient(tc)
}

const (
	owner = "eternal-flame-AD"
	repo  = "yumechi.jp"
)
