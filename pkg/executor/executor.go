/*
Copyright © 2024 Arush Salil

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package executor

import (
	"context"
	"log"

	"github.com/arush-sal/branch-protection-sync/pkg/getter"
	"github.com/arush-sal/branch-protection-sync/pkg/setter"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func Run(owner, sourceRepo, token string) {
	ctx := context.Background()
	client := getGitHubClient(ctx, token)

	ruleset := getter.GetRuleset(ctx, client, owner, sourceRepo)
	repos, err := getter.GetAllReposFromOrg(ctx, client, owner)
	if err != nil {
		log.Fatalf("Error fetching repositories: %v\n", err)
		return
	}

	setter.SetRuleset(ctx, client, owner, repos, ruleset)
}

func getGitHubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}
