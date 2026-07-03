package git

// git_repository_helpers.go — shared helpers used by both framework and (formerly) SDKv2
// implementations of the git resource/data-source types. The SDKv2 constructors and CRUD
// handlers have been removed (AC3); these helpers are kept because the framework
// implementations still reference them.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ---------------------------------------------------------------------------
// Repository init type enum (used by schema + framework expand logic)
// ---------------------------------------------------------------------------

// RepoInitType strategy for initializing the repo.
type RepoInitType string

type repoInitTypeValuesType struct {
	Uninitialized RepoInitType
	Clean         RepoInitType
	Fork          RepoInitType
	Import        RepoInitType
}

// RepoInitTypeValues enum of strategy for initializing the repo.
var RepoInitTypeValues = repoInitTypeValuesType{
	Uninitialized: "Uninitialized",
	Clean:         "Clean",
	Fork:          "Fork",
	Import:        "Import",
}

// repoInitializationMeta is a helper type that carries transient info used only during
// repository creation / import operations.
type repoInitializationMeta struct {
	initType            string
	sourceType          string
	sourceURL           string
	serviceConnectionID string
	userName            string
	password            string
}

// ---------------------------------------------------------------------------
// Branch ref constants and helper (used by framework branch resource)
// ---------------------------------------------------------------------------

const (
	REF_BRANCH_PREFIX = "refs/heads/"
	REF_TAG_PREFIX    = "refs/tags/"
)

func withPrefix(prefix, name string) string {
	if strings.HasPrefix(name, prefix) {
		return name
	}
	return prefix + name
}

// ---------------------------------------------------------------------------
// Repository CRUD helpers (used by framework git_repository resource)
// ---------------------------------------------------------------------------

// waitForBranch polls until the repository's default branch is set.
func waitForBranch(clients *client.AggregatedClient, repoName *string, projectID fmt.Stringer, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"Waiting"},
		Target:  []string{"Synched"},
		Refresh: func() (interface{}, string, error) {
			state := "Waiting"
			gitRepo, err := gitRepositoryRead(clients, "", *repoName, projectID.String())
			if err != nil {
				return nil, "", fmt.Errorf("Retrieving repository: %+v", err)
			}
			if converter.ToString(gitRepo.DefaultBranch, "") != "" {
				state = "Synched"
			}
			return state, state, nil
		},
		Timeout:                   timeout,
		MinTimeout:                2 * time.Second,
		Delay:                     1 * time.Second,
		ContinuousTargetOccurence: 1,
	}
	if _, err := stateConf.WaitForStateContext(clients.Ctx); err != nil {
		return fmt.Errorf("Retrieving expected branch for repository [%s]: %+v", *repoName, err)
	}
	return nil
}

// initializeGitRepository pushes an initial README commit to a newly created repo.
func initializeGitRepository(clients *client.AggregatedClient, repo *git.GitRepository, defaultBranch *string) error {
	branchName := converter.ToString(defaultBranch, "")
	if strings.EqualFold(branchName, "") {
		branchName = "refs/heads/master"
	}
	args := git.CreatePushArgs{
		RepositoryId: repo.Name,
		Project:      repo.Project.Name,
		Push: &git.GitPush{
			RefUpdates: &[]git.GitRefUpdate{
				{
					Name:        converter.String(branchName),
					OldObjectId: converter.String("0000000000000000000000000000000000000000"),
				},
			},
			Commits: &[]git.GitCommitRef{
				{
					Comment: converter.String("Initial commit."),
					Changes: &[]interface{}{
						git.Change{
							ChangeType: &git.VersionControlChangeTypeValues.Add,
							Item: git.GitItem{
								Path: converter.String("/README.md"),
							},
							NewContent: &git.ItemContent{
								ContentType: &git.ItemContentTypeValues.RawText,
								Content:     repo.Project.Name,
							},
						},
					},
				},
			},
		},
	}
	_, err := clients.GitReposClient.CreatePush(clients.Ctx, args)
	return err
}

// updateIsDisabledGitRepository sets the isDisabled flag on a repository.
func updateIsDisabledGitRepository(clients *client.AggregatedClient, repoID string, projectID string, isDisabled bool) (*git.GitRepository, error) {
	id, err := uuid.Parse(repoID)
	if err != nil {
		return nil, fmt.Errorf("invalid repositoryId UUID: %s", repoID)
	}
	repo, err := clients.GitReposClient.UpdateRepository(
		clients.Ctx,
		git.UpdateRepositoryArgs{
			NewRepositoryInfo: &git.GitRepository{IsDisabled: converter.Bool(isDisabled)},
			RepositoryId:      &id,
			Project:           converter.String(projectID),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("updating isDisabled on repository: %+v", err)
	}
	return repo, nil
}

// gitRepositoryRead looks up a repository by ID or name.
// If the repository is disabled (returns 404 from the GET endpoint) it falls
// back to listing all repos (including hidden) and searching by ID/name.
func gitRepositoryRead(clients *client.AggregatedClient, repoID string, repoName string, projectID string) (*git.GitRepository, error) {
	identifier := repoID
	if strings.EqualFold(identifier, "") {
		identifier = repoName
	}

	repo, err := clients.GitReposClient.GetRepository(clients.Ctx, git.GetRepositoryArgs{
		RepositoryId: converter.String(identifier),
		Project:      converter.String(projectID),
	})

	// Disabled repositories cannot be obtained through the normal GET endpoint.
	if utils.ResponseWasNotFound(err) {
		var allRepo *[]git.GitRepository
		allRepo, err = clients.GitReposClient.GetRepositories(clients.Ctx, git.GetRepositoriesArgs{
			Project:       converter.String(projectID),
			IncludeHidden: converter.Bool(true),
		})
		if err != nil {
			return nil, err
		}
		for _, gitRepo := range *allRepo {
			if strings.EqualFold(gitRepo.Id.String(), identifier) ||
				strings.EqualFold(*gitRepo.Name, identifier) {
				repo = &gitRepo
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// ---------------------------------------------------------------------------
// Data-source helpers (used by framework git_repositories + git_repository data sources)
// ---------------------------------------------------------------------------

// getGitRepositoriesByNameAndProject returns repositories filtered by optional name inside a project.
// When both name and projectID are non-empty it uses gitRepositoryRead for an exact lookup;
// otherwise it lists all repos (respecting includeHidden) and optionally filters by name.
func getGitRepositoriesByNameAndProject(clients *client.AggregatedClient, name string, projectID string, includeHidden bool) (*[]git.GitRepository, error) {
	var repos *[]git.GitRepository
	var err error

	if name != "" && projectID != "" {
		repo, err := gitRepositoryRead(clients, "", name, projectID)
		if err != nil {
			return nil, err
		}
		if repo != nil {
			repos = &[]git.GitRepository{*repo}
		}
	} else {
		repos, err = clients.GitReposClient.GetRepositories(clients.Ctx, git.GetRepositoriesArgs{
			Project:       converter.String(projectID),
			IncludeHidden: converter.Bool(includeHidden),
		})
		if err != nil {
			return nil, err
		}
		if name != "" {
			for _, repo := range *repos {
				if strings.EqualFold(*repo.Name, name) {
					repos = &[]git.GitRepository{repo}
					break
				}
			}
		}
	}
	return repos, nil
}

// ---------------------------------------------------------------------------
// File resource helpers (used by framework git_repository_file resource+datasource)
// ---------------------------------------------------------------------------

// shortBranchName removes the refs/heads/ prefix that some API endpoints require.
func shortBranchName(branch string) string {
	return strings.TrimPrefix(branch, "refs/heads/")
}

// splitRepoFilePath splits a resource ID into repository ID and file path components.
func splitRepoFilePath(path string) (string, string) {
	parts := strings.Split(path, "/")
	return parts[0], strings.Join(parts[1:], "/")
}

// checkRepositoryBranchExists returns the GitRef for the given branch, or nil if it does not exist.
func checkRepositoryBranchExists(c *client.AggregatedClient, repoId, branch string) (*git.GitRef, error) {
	ctx := context.Background()
	branchName := shortBranchName(branch)
	resp, err := c.GitReposClient.GetRefs(ctx, git.GetRefsArgs{
		RepositoryId: &repoId,
		Filter:       converter.String("heads/" + branchName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get repository branch %q: %+v", branch, err)
	}
	if resp != nil {
		for _, ref := range resp.Value {
			if strings.EqualFold(branchName, shortBranchName(*ref.Name)) {
				return &ref, nil
			}
		}
	}
	return nil, nil
}

// checkRepositoryFileExists returns an error if the file does not exist in the repository.
func checkRepositoryFileExists(c *client.AggregatedClient, repoId, file, branch string) error {
	ctx := context.Background()
	_, err := c.GitReposClient.GetItem(ctx, git.GetItemArgs{
		RepositoryId: &repoId,
		Path:         &file,
		VersionDescriptor: &git.GitVersionDescriptor{
			Version: converter.String(shortBranchName(branch)),
		},
	})
	return err
}

// getLastCommitId returns the most recent commit ID on the given branch.
func getLastCommitId(c *client.AggregatedClient, repoId, branch string) (string, error) {
	ctx := context.Background()
	commits, err := c.GitReposClient.GetCommits(ctx, git.GetCommitsArgs{
		RepositoryId: &repoId,
		Top:          converter.Int(1),
		SearchCriteria: &git.GitQueryCommitsCriteria{
			ItemVersion: &git.GitVersionDescriptor{
				Version: converter.String(shortBranchName(branch)),
			},
		},
	})
	if err != nil {
		return "", err
	}
	return *(*commits)[0].CommitId, nil
}

// shortTagName removes the refs/tags/ prefix.
func shortTagName(tag string) string {
	return strings.TrimPrefix(tag, "refs/tags/")
}

// updateGitRepository updates repository metadata (name, default branch, etc.).
func updateGitRepository(clients *client.AggregatedClient, repository *git.GitRepository, project fmt.Stringer) (*git.GitRepository, error) {
	return clients.GitReposClient.UpdateRepository(
		clients.Ctx,
		git.UpdateRepositoryArgs{
			NewRepositoryInfo: repository,
			RepositoryId:      repository.Id,
			Project:           converter.String(project.String()),
		},
	)
}
