package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/verlyn13/ds-go/internal/config"
	"github.com/verlyn13/ds-go/internal/scan"
	"github.com/verlyn13/ds-go/internal/ui"
)

var (
	cfgFile     string
	jsonOutput  bool
	fetchFirst  bool
	dirtyOnly   bool
	accountFilter string
	scanPath    string
	clonePath   string
	quietMode   bool
	workerCount int
)

var rootCmd = &cobra.Command{
	Use:   "ds",
	Short: "Dead Simple repository manager",
	Long:  `A blazing fast Git repository scanner and status reporter.`,
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"s"},
	Short:   "Show repository status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		scanner := scan.New(cfg, workerCount)
		repos, err := scanner.Scan(scanPath)
		if err != nil {
			return fmt.Errorf("scanning repos: %w", err)
		}

		// Apply filters
		if dirtyOnly {
			repos = filterDirty(repos)
		}
		if accountFilter != "" {
			repos = filterByAccount(repos, accountFilter)
		}

		if jsonOutput {
			return ui.PrintJSON(repos)
		}
		return ui.PrintTable(repos, cfg)
	},
}

var fetchCmd = &cobra.Command{
	Use:     "fetch",
	Aliases: []string{"f"},
	Short:   "Fetch all repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		scanner := scan.New(cfg, workerCount)
		repos, err := scanner.Scan(scanPath)
		if err != nil {
			return fmt.Errorf("scanning repos: %w", err)
		}

		fetcher := scan.NewFetcher(workerCount)
		results := fetcher.FetchAll(repos, !quietMode)
		
		if !quietMode {
			ui.PrintFetchResults(results)
		}
		return nil
	},
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for repositories and rebuild index",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		scanner := scan.New(cfg, workerCount)
		
		if fetchFirst {
			repos, _ := scanner.Scan(scanPath)
			fetcher := scan.NewFetcher(workerCount)
			fetcher.FetchAll(repos, !quietMode)
		}
		
		repos, err := scanner.Scan(scanPath)
		if err != nil {
			return fmt.Errorf("scanning: %w", err)
		}

		// Save index
		if err := scanner.SaveIndex(repos); err != nil {
			return fmt.Errorf("saving index: %w", err)
		}

		fmt.Printf("Scanned %d repositories\n", len(repos))
		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize or update configuration",
	Long:  `Creates a new configuration file or updates an existing one with your GitHub accounts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.InitInteractive(cfgFile)
	},
}

var cloneCmd = &cobra.Command{
	Use:   "clone <repo-url>",
	Short: "Clone a repository with proper SSH config",
	Long:  `Clone a GitHub repository using the appropriate SSH host configuration based on the owner.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		
		return scan.CloneRepo(args[0], cfg, clonePath)
	},
}

var cdCmd = &cobra.Command{
	Use:   "cd <repo-name>",
	Short: "Print path to change directory to a repository",
	Long:  `Prints the cd command to navigate to a repository. Use with: $(ds cd repo-name)`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		
		scanner := scan.New(cfg, 1)
		repos, err := scanner.Scan("")
		if err != nil {
			return fmt.Errorf("scanning repos: %w", err)
		}
		
		repoName := args[0]
		for _, repo := range repos {
			if repo.Name == repoName || 
			   strings.Contains(repo.Path, "/"+repoName) ||
			   strings.HasSuffix(repo.Path, repoName) {
				fmt.Print(repo.Path)
				return nil
			}
		}
		
		return fmt.Errorf("repository '%s' not found", repoName)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and manage ds configuration settings.`,
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		
		if jsonOutput {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(cfg)
		}
		
		// Print in readable format
		fmt.Printf("Configuration file: %s\n\n", config.DefaultPath())
		fmt.Printf("Base directory: %s\n\n", cfg.BaseDir)
		
		fmt.Println("Accounts:")
		for name, acc := range cfg.Accounts {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    Type: %s\n", acc.Type)
			fmt.Printf("    SSH Host: %s\n", acc.SSHHost)
			if acc.Email != "" {
				fmt.Printf("    Email: %s\n", acc.Email)
			}
		}
		
		if len(cfg.Orgs) > 0 {
			fmt.Println("\nOrganizations:")
			for org, host := range cfg.Orgs {
				fmt.Printf("  %s: %s\n", org, host)
			}
		}
		
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration in editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := config.DefaultPath()
		if cfgFile != "" {
			configPath = cfgFile
		}
		
		// Try common editors in order of preference
		editors := []string{"$EDITOR", "vim", "nano", "vi"}
		for _, editor := range editors {
			if editor == "$EDITOR" {
				editor = os.Getenv("EDITOR")
				if editor == "" {
					continue
				}
			}
			
			cmd := exec.Command(editor, configPath)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		
		return fmt.Errorf("no editor found. Set $EDITOR or install vim/nano")
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $XDG_CONFIG_HOME/ds/config.yaml)")
	rootCmd.PersistentFlags().IntVarP(&workerCount, "workers", "w", 10, "number of concurrent workers")
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "suppress progress output")

	statusCmd.Flags().BoolVarP(&dirtyOnly, "dirty", "d", false, "show only repositories with uncommitted changes")
	statusCmd.Flags().StringVarP(&accountFilter, "account", "a", "", "filter by account")
	statusCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	statusCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: ~/Projects)")

	scanCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: ~/Projects)")
	scanCmd.Flags().BoolVar(&fetchFirst, "fetch", false, "fetch all repos before scanning")
	scanCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	fetchCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: ~/Projects)")

	cloneCmd.Flags().StringVarP(&clonePath, "path", "p", "", "directory to clone into")

	configViewCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configEditCmd)

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(cdCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = config.DefaultPath()
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func filterDirty(repos []scan.Repository) []scan.Repository {
	var filtered []scan.Repository
	for _, repo := range repos {
		if !repo.IsClean {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}

func filterByAccount(repos []scan.Repository, account string) []scan.Repository {
	var filtered []scan.Repository
	for _, repo := range repos {
		if repo.Account == account {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}