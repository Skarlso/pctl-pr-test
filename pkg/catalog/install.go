package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	profilesv1 "github.com/weaveworks/profiles/api/v1alpha1"

	"github.com/weaveworks/pctl/pkg/git"
	"github.com/weaveworks/pctl/pkg/writer"
)

// InstallConfig defines parameters for the installation call.
type InstallConfig struct {
	Branch      string
	CatalogName string
	CatalogURL  string
	ConfigMap   string
	Namespace   string
	ProfileName string
	SubName     string
	Writer      writer.Writer
}

// Install using the catalog at catalogURL and a profile matching the provided profileName generates a profile subscription
// writing it out with the provided profile subscription writer.
func Install(cfg InstallConfig) error {
	u, err := url.Parse(cfg.CatalogURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %q: %w", cfg.CatalogURL, err)
	}

	u.Path = fmt.Sprintf("profiles/%s/%s", cfg.CatalogName, cfg.ProfileName)
	resp, err := doRequest(u, nil)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close the response body from profile show with error: %v/n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("unable to find profile `%s` in catalog `%s`", cfg.ProfileName, cfg.CatalogName)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch profile: status code %d", resp.StatusCode)
	}

	profile := profilesv1.ProfileDescription{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	subscription := profilesv1.ProfileSubscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProfileSubscription",
			APIVersion: "weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.SubName,
			Namespace: cfg.Namespace,
		},
		Spec: profilesv1.ProfileSubscriptionSpec{
			ProfileURL: profile.URL,
			Branch:     cfg.Branch,
		},
	}
	if cfg.ConfigMap != "" {
		subscription.Spec.ValuesFrom = []helmv2.ValuesReference{
			{
				Kind:      "ConfigMap",
				Name:      cfg.SubName + "-values",
				ValuesKey: cfg.ConfigMap,
			},
		}
	}
	if err := cfg.Writer.Output(&subscription); err != nil {
		return fmt.Errorf("failed to output subscription information: %w", err)
	}
	return nil
}

// CreatePullRequest creates a pull request from the current changes.
func CreatePullRequest(repo string, base string, branch string, g git.Git) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("failed to find git on path: %w", err)
	}

	if err := g.IsRepository(); err != nil {
		return fmt.Errorf("directory is not a git repository: %w", err)
	}

	if err := g.CreateBranch(); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branch, err)
	}

	if err := g.Add(); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	if err := g.Commit(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	if err := g.Push(); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	// Create the PR
	client, err := factory.NewClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to create scm client: %w", err)
	}

	fmt.Println("Creating pull request with : ", repo, base, branch)
	ctx := context.Background()
	request, _, err := client.PullRequests.Create(ctx, repo, &scm.PullRequestInput{
		Title: "PCTL Generated Profile Resource Update",
		Head:  branch,
		Base:  base,
	})
	if err != nil {
		return fmt.Errorf("error while creating pr: %w", err)
	}
	fmt.Printf("PR created with number: %d and URL: %s\n", request.Number, request.Link)
	return nil

}
