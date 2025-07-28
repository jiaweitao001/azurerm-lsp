package processor

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Azure/ms-terraform-lsp/module-schema/registry"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

const (
	githubOrg       = "Azure"
	repoNamePattern = "^terraform-azurerm-avm-.*$"      // Adjust this regex to match your repository naming convention
	OutputDir       = "module-schema/fetched_hcl_files" // Directory to save fetched files
)

func FetchRepositoryData() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set. Please set it to your GitHub Personal Access Token.")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repoRegex, err := regexp.Compile(repoNamePattern)
	if err != nil {
		log.Fatalf("Error compiling regex for repo name pattern: %v", err)
	}

	variablesOutputDir := filepath.Join(OutputDir, "variables")
	examplesOutputDir := filepath.Join(OutputDir, "examples")

	if err := os.MkdirAll(variablesOutputDir, 0755); err != nil {
		log.Fatalf("Error creating variables output directory: %v", err)
	}
	if err := os.MkdirAll(examplesOutputDir, 0755); err != nil {
		log.Fatalf("Error creating examples output directory: %v", err)
	}

	log.Printf("Searching for repositories matching pattern '%s' in organization/user '%s'...", repoNamePattern, githubOrg)
	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{
		Type: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := client.Repositories.List(ctx, githubOrg, opt)
		if err != nil {
			log.Fatalf("Error listing repositories: %v", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.Printf("Found %d repositories. Filtering by pattern...", len(allRepos))

	for _, repo := range allRepos {
		if repoRegex.MatchString(repo.GetName()) {
			log.Printf("Repository '%s' matches pattern. Checking for 'variables.tf'...", repo.GetName())

			moduleName, err := registry.GetModuleByRepoName(repo.GetName())
			if err != nil {
				log.Printf("Error retrieving module name for repo '%s': %v", repo.GetName(), err)
				continue
			}
			nameArr := strings.Split(moduleName, "/")
			fmt.Printf("Module name for repo '%s': %s\n", repo.GetName(), moduleName)
			registryName := nameArr[1]

			// Fetch variables.tf
			variablesFilePath := "variables.tf"
			variablesOutputFilePath := filepath.Join(variablesOutputDir, fmt.Sprintf("%s_variables.tf", registryName))
			err = fetchAndSaveFile(ctx, client, githubOrg, repo, variablesFilePath, variablesOutputFilePath)
			if err != nil {
				log.Printf("Could not fetch 'variables.tf' for repo '%s', skipping: %v", repo.GetName(), err)
				continue // Skip to the next repository if variables.tf is not found
			}

			// Fetch examples/default/main.tf
			exampleFilePath := "examples/default/main.tf"
			exampleOutputFilePath := filepath.Join(examplesOutputDir, fmt.Sprintf("%s_example.tf", registryName))
			err = fetchAndSaveFile(ctx, client, githubOrg, repo, exampleFilePath, exampleOutputFilePath)
			if err != nil {
				log.Printf("Could not fetch 'examples/default/main.tf' for repo '%s': %v", repo.GetName(), err)
				// Continue even if the example file is not found
			}
		}
	}
}

func fetchAndSaveFile(ctx context.Context, client *github.Client, org string, repo *github.Repository, filePath string, outputFilePath string) error {
	fileContent, _, resp, err := client.Repositories.GetContents(
		ctx,
		org,
		repo.GetName(),
		filePath,
		&github.RepositoryContentGetOptions{Ref: repo.GetDefaultBranch()},
	)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return fmt.Errorf("file '%s' not found", filePath)
		}
		return fmt.Errorf("error fetching file '%s': %w", filePath, err)
	}

	if fileContent.Content == nil {
		return fmt.Errorf("file '%s' has no content or is a directory", filePath)
	}

	decodedContent, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(*fileContent.Content, "\n", ""))
	if err != nil {
		return fmt.Errorf("error decoding content for '%s': %w", filePath, err)
	}

	err = os.WriteFile(outputFilePath, decodedContent, 0644)
	if err != nil {
		return fmt.Errorf("error writing file '%s': %w", outputFilePath, err)
	}

	log.Printf("Successfully fetched and saved '%s' from '%s' to '%s'", filePath, repo.GetName(), outputFilePath)
	return nil
}
