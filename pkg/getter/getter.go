/*
# Copyright © 2024 Arush Salil

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
package getter

import (
	"context"
	"log"

	"github.com/google/go-github/v39/github"
)

// getDefaultBranch retrieves the default branch of a repository.
func getDefaultBranch(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return *repository.DefaultBranch, nil
}

// getBranchProtectionRules retrieves the branch protection rules for a specific repository.
func getBranchProtection(ctx context.Context, client *github.Client, owner, repo string) (*github.Protection, error) {

	branch, err := getDefaultBranch(ctx, client, owner, repo)
	if err != nil {
		return nil, err
	}

	protection, _, err := client.Repositories.GetBranchProtection(ctx, owner, repo, branch)
	if err != nil {
		return nil, err
	}
	return protection, nil
}

// GetRuleset retrieves the branch protection rules for a specific repository.
func GetRuleset(ctx context.Context, client *github.Client, owner, repo string) *github.Protection {
	gp, err := getBranchProtection(ctx, client, owner, repo)
	// client.Repositories.GetPullRequestReviewEnforcement (ctx context.Context, owner, repo, branch string) (*PullRequestReviewsEnforcement, *Response, error)
	// GetRequiredStatusChecks(ctx context.Context, owner, repo, branch string) (*RequiredStatusChecks, *Response, error)
	// GetSignaturesProtectedBranch(ctx context.Context, owner, repo, branch string) (*SignaturesProtectedBranch, *Response, error)
	// RequireSignaturesOnProtectedBranch(ctx context.Context, owner, repo, branch string, requireSignatures bool) (*Response, error)
	//
	if err != nil {
		log.Fatalf("Error fetching branch protection rules: %v\n", err)
	}
	return gp
}

// GetAllReposFromOrg fetches all repositories for the specified GitHub organization.
func GetAllReposFromOrg(ctx context.Context, client *github.Client, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opts := &github.RepositoryListByOrgOptions{
		// PerPage can be adjusted to your needs
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// Pagination handling
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		// Update opts.Page to fetch the next page
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}
