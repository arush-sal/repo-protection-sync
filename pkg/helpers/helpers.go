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
package helpers

import (
	"context"
	"errors"
	"log"

	"github.com/google/go-github/v59/github"
)

func HTTPStatusCodeCheck(statuscode int) error {
	// Check the HTTP status code
	switch statuscode {
	case 200, 201:
		// Created
		log.Printf("request successful[%d]", statuscode)
		return nil
	case 303:
		log.Printf("same branch name pattern already exists[%d]", statuscode)
		return errors.New("same branch name pattern already exists")
	case 400:
		// Bad Request
		break
	case 401:
		// Unauthorized
		break
	case 403:
		// Forbidden
		log.Printf("Forbidden[%d]", statuscode)
		return errors.New("Forbidden")
	case 404:
		// Not Found
		log.Printf("resource not found[%d]", statuscode)
		return errors.New("resource not found")
	case 405:
		// Method Not Allowed
		break
	case 406:
		// Not Acceptable
		break
	case 409:
		// Conflict
		break
	case 410:
		// Gone
		break
	case 422:
		// Unprocessable Entity
		log.Printf("Validation failed, or the endpoint has been spammed.[%d]", statuscode)
		return errors.New("Validation failed, or the endpoint has been spammed")
	// case 429:
	// 	// Too Many Requests
	// 	break
	// case 500:
	// 	// Internal Server Error
	// 	break
	// case 501:
	// 	// Not Implemented
	// 	break
	// case 502:
	// 	// Bad Gateway
	// 	break
	// case 503:
	// 	// Service Unavailable
	// 	break
	// case 504:
	// 	// Gateway Timeout
	// 	break
	default:
		// Handle unexpected response status code
		log.Fatalf("ERROR: HTTP request failed with status code: %d\n", statuscode)
		return errors.New("HTTP request failed with status code")
	}

	return nil
}

func DoesRulesetExist(ctx context.Context, client *github.Client, owner, repo, branch, sourceRuleset string) bool {
	targetRulesets, response, err := client.Repositories.GetAllRulesets(ctx, owner, repo, false)
	switch {
	case err != nil:
		log.Fatalf("Error fetching branch ruleset: %v\n", err)
	case HTTPStatusCodeCheck(response.StatusCode) != nil:
		log.Fatalf("Error fetching branch ruleset: %v\n", HTTPStatusCodeCheck(response.StatusCode))
	}
	for _, targetRuleset := range targetRulesets {
		if sourceRuleset == targetRuleset.Name {
			return true
		}
	}
	return false
}
