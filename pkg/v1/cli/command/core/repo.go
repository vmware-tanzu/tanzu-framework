package core

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

var (
	gcpBucketName, gcpRootPath, name string
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage plugin repositories",
}

func init() {
	repoCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	repoCmd.AddCommand(
		listRepoCmd,
		addRepoCmd,
		deleteRepoCmd,
	)
	addRepoCmd.Flags().StringVarP(&name, "name", "n", "", "name of repository")
	addRepoCmd.Flags().StringVarP(&gcpBucketName, "gcp-bucket-name", "b", "", "name of gcp bucket")
	addRepoCmd.Flags().StringVarP(&gcpRootPath, "gcp-root-path", "r", "", "root path in gcp bucket")
	cobra.MarkFlagRequired(addRepoCmd.Flags(), "name")
	cobra.MarkFlagRequired(addRepoCmd.Flags(), "gcp-bucket-name")
	cobra.MarkFlagRequired(addRepoCmd.Flags(), "gcp-root-path")
}

var listRepoCmd = &cobra.Command{
	Use:   "list",
	Short: "List available repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.GetConfig()
		if err != nil {
			return err
		}
		repos := cli.LoadRepositories(cfg)
		var rows = len(repos)
		data := make([][]string, rows)
		for index := 0; index < rows; index++ {
			data[index] = []string{repos[index].Name()}
		}

		table := cli.NewTableWriter("Name")
		table.AppendBulk(data)

		table.Render()

		return nil
	},
}

var addRepoCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.GetConfig()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil {
			cfg.ClientOptions = &clientv1alpha1.ClientOptions{}
		}
		if cfg.ClientOptions.CLI == nil {
			cfg.ClientOptions.CLI = &clientv1alpha1.CLIOptions{}
		}
		repos := cfg.ClientOptions.CLI.Repositories
		gcpRepo := &clientv1alpha1.GCPPluginRepository{
			Name:       name,
			BucketName: gcpBucketName,
			RootPath:   gcpRootPath,
		}
		pluginRepo := clientv1alpha1.PluginRepository{
			GCPPluginRepository: gcpRepo,
		}
		for _, repo := range repos {
			if repo.GCPPluginRepository != nil {
				if repo.GCPPluginRepository.Name == name {
					return fmt.Errorf("repo name %q already exists", name)
				}
			}
		}
		repos = append(repos, pluginRepo)
		cfg.ClientOptions.CLI.Repositories = repos
		err = client.StoreConfig(cfg)
		if err != nil {
			return err
		}
		return nil
	},
}

var deleteRepoCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		repoName := args[0]
		cfg, err := client.GetConfig()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil || cfg.ClientOptions.CLI == nil {
			return fmt.Errorf("repository %q unknown", repoName)
		}

		r := cfg.ClientOptions.CLI.Repositories
		newRepos := []clientv1alpha1.PluginRepository{}
		for _, repo := range r {
			if repo.GCPPluginRepository == nil {
				continue
			}
			if repo.GCPPluginRepository.Name == repoName {
				continue
			}
			newRepos = append(newRepos, repo)
		}
		cfg.ClientOptions.CLI.Repositories = newRepos
		err = client.StoreConfig(cfg)
		if err != nil {
			return err
		}
		return nil
	},
}
