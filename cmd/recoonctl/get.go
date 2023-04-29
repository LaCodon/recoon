package main

import (
	"fmt"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/client"
	"github.com/lacodon/recoon/pkg/controller/configrepo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"text/tabwriter"
	"time"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an object from the API",
	RunE:  getCmdRun,
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func getCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must pass object type")
	}

	switch args[0] {
	case "repository":
		fallthrough
	case "repositories":
		fallthrough
	case "repo":
		return getRepository(args)

	case "project":
		fallthrough
	case "proj":
		return getProject(args)

	case "container":
		return getContainer(args)

	default:
		return fmt.Errorf("unknown type")
	}
}

func getRepository(args []string) error {
	c := client.New("http://localhost:3680/api/v1")

	if len(args) == 1 {
		repos, err := c.GetRepositories()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "PROJECT\tREPO\tBRANCH\tPATH\tCOMMIT\t")

		for _, repo := range repos {
			projectName := repo.Spec.ProjectName
			if repo.GetName() == configrepo.ConfigRepoName {
				projectName = "RECOON-CONFIG"
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n",
				projectName, repo.Spec.Url, repo.Spec.Branch, repo.Spec.Path, repo.Status.CurrentCommitId)
		}

		return w.Flush()
	}

	repo, err := c.GetRepository(args[1])
	if err != nil {
		return err
	}

	out, _ := yaml.Marshal(repo)
	fmt.Print(string(out))

	return nil
}

func getProject(args []string) error {
	c := client.New("http://localhost:3680/api/v1")

	if len(args) == 1 {
		projects, err := c.GetProjects()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "PROJECT\tLAST_APPLIED_COMMIT_ID\tSTATUS\tTRANSITION_TIME\t")

		for _, project := range projects {
			lastAppliedCommit := ""
			status := "PENDING"
			transitionTime := ""

			if project.Status != nil {
				lastAppliedCommit = project.Status.LastAppliedCommitId

				if cond, ok := project.Status.Conditions[projectv1.ConditionSuccess]; ok {
					status = cond.Message
					transitionTime = cond.LastTransitionTime.Format(time.RFC822)
				} else if cond, ok := project.Status.Conditions[projectv1.ConditionFailure]; ok {
					status = "ERROR; see details"
					transitionTime = cond.LastTransitionTime.Format(time.RFC822)
				}

				if project.Spec.CommitId != lastAppliedCommit {
					status = "PENDING"
				}
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n",
				project.GetName(), lastAppliedCommit, status, transitionTime)
		}

		return w.Flush()
	}

	project, err := c.GetProject(args[1])
	if err != nil {
		return err
	}

	out, _ := yaml.Marshal(project)
	fmt.Print(string(out))

	return nil
}

func getContainer(args []string) error {
	c := client.New("http://localhost:3680/api/v1")

	projectName := ""
	if len(args) == 2 {
		projectName = args[1]
	}

	containers, err := c.GetContainers(projectName)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tPROJECT\tSERVICE\tIMAGE\tDIGEST\tSTATE\tSTATUS\t")

	for _, container := range containers {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
			container.ID[:20], container.Labels["com.docker.compose.project"], container.Labels["com.docker.compose.service"], container.Image, container.ImageID, container.State, container.Status)
	}

	return w.Flush()
}
