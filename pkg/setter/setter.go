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
package setter

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/arush-sal/branch-protection-sync/pkg/helpers"
	"github.com/google/go-github/v39/github"
)

// SetRuleset sets the branch protection rules for the list of repositories provided
// under a particular GitHub user or organization.
func SetRuleset(ctx context.Context, client *github.Client, owner string, repos []*github.Repository, ruleset *github.Protection) {

	// Calculate the number of semaphores as one tenth of the total number of repos
	// with a minimum of 1
	semaphoreCount := len(repos) / 10
	if semaphoreCount < 1 {
		semaphoreCount = 1
	}
	semaphore := make(chan struct{}, semaphoreCount)

	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a semaphore

		go func(repo *github.Repository) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore

			if repo == nil || repo.Name == nil || repo.DefaultBranch == nil {
				log.Printf("Skipping repository due to missing information: %+v\n", repo)
				return
			}

			// Check and handle rate limit before attempting to set branch protection
			if !checkAndHandleRateLimit(ctx, client) {
				log.Printf("Failed to handle rate limit, skipping repo: %s\n", *repo.Name)
				return
			}

			log.Printf("Starting branch protection sync for repo %s...", *repo.Name)
			// Assume setBranchProtectionRules is implemented to call the GitHub API
			err := setBranchProtectionRules(ctx, client, owner, *repo.Name, *repo.DefaultBranch, convertProtectionToRequest(ruleset))
			// GetSignaturesOnProtectedBranch
			// RequireSignaturesOnProtectedBranch
			// GetRequiredDeploymentsEnforcementLevel
			//
			if err != nil {
				log.Fatalf("Error applying branch protection to repo %s: %v\n", *repo.Name, err)
			} else {
				log.Printf("Branch protection applied to repo %s successfully\n",
					*repo.Name)
			}
		}(repo)
	}

	wg.Wait() // Wait for all goroutines to complete
}

// checkAndHandleRateLimit checks the rate limit for the GitHub API and
// in case if the rate limiting exceeds it handles the particular scenario by
// adding a wait time before the next request is made.
func checkAndHandleRateLimit(ctx context.Context, client *github.Client) bool {
	rateLimits, _, err := client.RateLimits(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch rate limit: %v\n", err)
		return false
	}

	if rateLimits.Core.Remaining < 1 {
		resetTime := rateLimits.Core.Reset.Time
		waitDuration := time.Until(resetTime)
		log.Printf("Rate limit exceeded. Waiting until %v (%v)\n", resetTime, waitDuration)
		time.Sleep(waitDuration + time.Second) // Add a buffer to ensure limit has reset
	}

	return true
}

// setBranchProtectionRules applies branch protection rules to a specified branch in a GitHub repository.
func setBranchProtectionRules(ctx context.Context, client *github.Client, owner, repo, branch string, protection *github.ProtectionRequest) error {
	_, response, err := client.Repositories.UpdateBranchProtection(ctx, owner, repo, branch, protection)
	if err != nil {
		log.Fatalf("Error applying branch protection: %v\n", err)
		return err
	}

	// log.Printf("Branch protection details: %v\n", protectionDetails)

	// Optionally, inspect response.StatusCode to ensure it's 200 OK
	// or handle redirections (HTTP 301, 302) if necessary.
	// log.Println("Branch protection applied successfully.")

	return helpers.HTTPStatusCodeCheck(response.StatusCode)

}

// convertProtectionToRequest converts a github.Protection object to a github.ProtectionRequest object.
func convertProtectionToRequest(protection *github.Protection) *github.ProtectionRequest {
	if protection == nil {
		log.Fatal("Protection object is nil")
	}

	// Initialize the ProtectionRequest with zero values.
	request := &github.ProtectionRequest{}

	// Required status checks
	if protection.RequiredStatusChecks != nil {
		request.RequiredStatusChecks = protection.GetRequiredStatusChecks()
	}

	// Required pull request reviews
	if protection.RequiredPullRequestReviews != nil {
		request.RequiredPullRequestReviews = &github.PullRequestReviewsEnforcementRequest{
			DismissStaleReviews:          protection.RequiredPullRequestReviews.DismissStaleReviews,
			RequireCodeOwnerReviews:      protection.RequiredPullRequestReviews.RequireCodeOwnerReviews,
			RequiredApprovingReviewCount: protection.RequiredPullRequestReviews.RequiredApprovingReviewCount,
		}

		// Add Dismissal restrictions
		if protection.RequiredPullRequestReviews.DismissalRestrictions != nil {
			request.RequiredPullRequestReviews.DismissalRestrictionsRequest =
				convertPRDismissalRestrictionsToRequest(
					protection.RequiredPullRequestReviews.DismissalRestrictions,
				)
		}
	} else {
		request.RequiredPullRequestReviews = &github.PullRequestReviewsEnforcementRequest{}
	}

	// Enforce admin restrictions
	if protection.EnforceAdmins != nil {
		request.EnforceAdmins = protection.EnforceAdmins.Enabled
	} else {
		request.EnforceAdmins = false
	}

	// Add User, Team and Apps restrictions
	if protection.Restrictions != nil {
		request.Restrictions = convertProtectionRestrictionToRequest(protection.Restrictions)
	} else {
		request.Restrictions = &github.BranchRestrictionsRequest{}
	}

	request.RequireLinearHistory = &protection.RequireLinearHistory.Enabled
	request.AllowForcePushes = &protection.AllowForcePushes.Enabled
	request.AllowDeletions = &protection.AllowDeletions.Enabled
	request.RequiredConversationResolution = &protection.RequiredConversationResolution.Enabled

	return request
}

// convertPRDismissalRestrictionsToRequest converts a DismissalRestrictions object to a DismissalRestrictionsRequest object.
func convertPRDismissalRestrictionsToRequest(dr *github.DismissalRestrictions) *github.DismissalRestrictionsRequest {
	drr := &github.DismissalRestrictionsRequest{}

	// Add User dismissal restrictions
	if len(dr.Users) > 0 {
		dru := make([]string, 0, len(dr.Users))
		for _, user := range dr.Users {
			dru = append(dru, user.GetLogin())
		}
		drr.Users = &dru
	} else {
		drr.Users = &[]string{}
	}

	// Add Team dismissal restrictions
	if len(dr.Teams) > 0 {
		tru := make([]string, 0, len(dr.Teams))
		for _, team := range dr.Teams {
			tru = append(tru, team.GetSlug())
		}
		drr.Teams = &tru
	} else {
		drr.Teams = &[]string{}
	}

	return drr

}

// convertProtectionRestrictionToRequest converts a BranchRestrictions object to a BranchRestrictionsRequest object.
func convertProtectionRestrictionToRequest(br *github.BranchRestrictions) *github.BranchRestrictionsRequest {
	brr := &github.BranchRestrictionsRequest{}

	// Add User restrictions
	if len(br.Users) > 0 {
		ur := make([]string, 0, len(br.Users))
		for _, user := range br.Users {
			ur = append(ur, user.GetLogin())
		}

		brr.Users = ur
	} else {
		brr.Users = []string{}
	}

	// Add Team restrictions
	if len(br.Teams) > 0 {
		tr := make([]string, 0, len(br.Teams))
		for _, team := range br.Teams {
			tr = append(tr, team.GetSlug())
		}

		brr.Teams = tr
	} else {
		brr.Teams = []string{}
	}

	// Add App restrictions
	if len(br.Apps) > 0 {
		ar := make([]string, 0, len(br.Apps))
		for _, app := range br.Apps {
			ar = append(ar, app.GetSlug())
		}

		brr.Apps = ar
	} else {
		brr.Apps = []string{}
	}

	return brr
}
