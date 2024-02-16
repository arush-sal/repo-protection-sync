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
package cmd

import (
	"os"

	"github.com/arush-sal/branch-protection-sync/pkg/executor"
	"github.com/spf13/cobra"
)

var owner, repo, githubToken string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "branch-protection-sync",
	Short: "Applies a GitHub branch protection ruleset from a source repository to all repositories in an organization",
	Run: func(cmd *cobra.Command, args []string) {
		if owner == "" || repo == "" || githubToken == "" {
			cmd.Help()
			os.Exit(1)
		}
		executor.Run(owner, repo, githubToken)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&owner, "owner", "o", "", "GitHub repo owner")
	rootCmd.MarkPersistentFlagRequired("owner")
	rootCmd.PersistentFlags().StringVarP(&repo, "repo", "r", "", "GitHub template repo for using the ruleset from")
	rootCmd.MarkPersistentFlagRequired("repo")
	rootCmd.PersistentFlags().StringVarP(&githubToken, "token", "t", "", "GitHub token for authentication")
	rootCmd.MarkPersistentFlagRequired("token")
}
