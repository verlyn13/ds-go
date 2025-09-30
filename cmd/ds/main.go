package main

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/verlyn13/ds-go/internal/config"
    "github.com/verlyn13/ds-go/internal/server"
    "github.com/verlyn13/ds-go/internal/scan"
    "github.com/verlyn13/ds-go/internal/ui"
    "github.com/verlyn13/ds-go/internal/policy"
    "github.com/verlyn13/ds-go/internal/runner"
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
    exitOnDirty bool
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
            if err := ui.PrintJSON(repos); err != nil { return err }
            if exitOnDirty && len(filterDirty(repos)) > 0 {
                os.Exit(10)
            }
            return nil
        }
        if err := ui.PrintTable(repos, cfg); err != nil { return err }
        if exitOnDirty && len(filterDirty(repos)) > 0 {
            os.Exit(10)
        }
        return nil
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
        if jsonOutput {
            return ui.PrintJSONFetchResults(results)
        }
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

        if jsonOutput {
            type scanSummary struct{ Count int `json:"count"` }
            enc := json.NewEncoder(os.Stdout)
            enc.SetIndent("", "  ")
            return enc.Encode(scanSummary{Count: len(repos)})
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

var organizeCmd = &cobra.Command{
	Use:   "organize",
	Short: "Reorganize repositories into account/org folders",
	Long:  `Moves repositories to the proper folder structure based on their remote URLs.
Repositories will be organized as: ~/Projects/account/repo-name`,
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
		
        dryRun, _ := cmd.Flags().GetBool("dry-run")
        force, _ := cmd.Flags().GetBool("force")
        plan, _ := cmd.Flags().GetBool("plan")
        requireClean, _ := cmd.Flags().GetBool("require-clean")

        if plan {
            plans := scan.OrganizePlanJSON(repos, cfg)
            if jsonOutput {
                return ui.PrintJSONResponse(true, plans, nil)
            }
            for _, p := range plans {
                mark := "USR"
                if p.IsOrg { mark = "ORG" }
                fmt.Printf("[%s] %s -> %s\n", mark, p.OldPath, p.NewPath)
            }
            fmt.Printf("%d moves planned\n", len(plans))
            return nil
        }

        if requireClean {
            for _, r := range repos {
                if !r.IsClean {
                    return fmt.Errorf("require-clean: '%s' has uncommitted changes", r.Name)
                }
            }
        }

        return scan.OrganizeRepos(repos, cfg, dryRun, force)
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
    statusCmd.Flags().BoolVar(&exitOnDirty, "exit-on-dirty", false, "exit with code 10 when dirty repos are found")
    statusCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: ~/Projects)")

	scanCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: ~/Projects)")
	scanCmd.Flags().BoolVar(&fetchFirst, "fetch", false, "fetch all repos before scanning")
	scanCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

    fetchCmd.Flags().StringVar(&scanPath, "path", "", "path to scan (default: ~/Projects)")
    fetchCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	cloneCmd.Flags().StringVarP(&clonePath, "path", "p", "", "directory to clone into")

	configViewCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configEditCmd)

    organizeCmd.Flags().Bool("dry-run", false, "preview changes without moving files")
    organizeCmd.Flags().Bool("force", false, "move repos even if destination exists")
    organizeCmd.Flags().Bool("plan", false, "show planned moves and exit")
    organizeCmd.Flags().Bool("require-clean", false, "abort if any repository has uncommitted changes")
    organizeCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(cdCmd)
    rootCmd.AddCommand(configCmd)
    rootCmd.AddCommand(organizeCmd)
    rootCmd.AddCommand(serveCmd)
    rootCmd.AddCommand(policyCmd)
    rootCmd.AddCommand(execCmd)
    rootCmd.AddCommand(hooksCmd)
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

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start a local HTTP API for agents",
    RunE: func(cmd *cobra.Command, args []string) error {
        addr, _ := cmd.Flags().GetString("addr")
        token, _ := cmd.Flags().GetString("token")
        cfg, err := config.Load(cfgFile)
        if err != nil { return fmt.Errorf("loading config: %w", err) }
        s := server.New(cfg, workerCount).WithToken(token)
        return s.Start(addr)
    },
}

func init() {
    serveCmd.Flags().String("addr", "127.0.0.1:7777", "address to bind the local API server")
    serveCmd.Flags().String("token", os.Getenv("DS_TOKEN"), "optional bearer token for API auth (overrides DS_TOKEN)")
}

var policyCmd = &cobra.Command{
    Use:   "policy",
    Short: "Policy and compliance checks",
}

var policyCheckCmd = &cobra.Command{
    Use:   "check",
    Short: "Run compliance checks from .project-compliance.yaml",
    RunE: func(cmd *cobra.Command, args []string) error {
        path, _ := cmd.Flags().GetString("file")
        if path == "" { path = ".project-compliance.yaml" }
        cfg, err := policy.Load(path)
        if err != nil { return fmt.Errorf("load policy: %w", err) }
        report, err := policy.RunChecks(cfg)
        if err != nil { return fmt.Errorf("run checks: %w", err) }
        if jsonOutput {
            enc := json.NewEncoder(os.Stdout)
            enc.SetIndent("", "  ")
            if err := enc.Encode(report); err != nil { return err }
        } else {
            fmt.Printf("Checks: %d, Passed: %d, Failed: %d\n", report.Summary.Total, report.Summary.Passed, report.Summary.Failed)
            for _, r := range report.Results {
                mark := "✓"
                if !r.Passed { mark = "✗" }
                fmt.Printf(" %s %-10s %s\n", mark, r.Severity, r.Name)
            }
        }
        // Exit non-zero if any critical failure
        failOn, _ := cmd.Flags().GetString("fail-on")
        if failOn != "" {
            th, err := policy.SeverityFromString(failOn)
            if err != nil { return err }
            if policy.FailIfAboveSeverity(report, th, cfg) {
                os.Exit(20)
            }
        }
        return nil
    },
}

func init() {
    policyCmd.AddCommand(policyCheckCmd)
    policyCheckCmd.Flags().String("file", ".project-compliance.yaml", "policy file")
    policyCheckCmd.Flags().String("fail-on", "critical", "fail on failed checks at or above this severity")
    policyCheckCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
}

var hooksCmd = &cobra.Command{
    Use:   "hooks",
    Short: "Manage Git hooks for quality gates",
}

var hooksInstallCmd = &cobra.Command{
    Use:   "install",
    Short: "Install pre-commit and pre-push hooks in the current repo",
    RunE: func(cmd *cobra.Command, args []string) error {
        hooksDir := ".git/hooks"
        if _, err := os.Stat(hooksDir); err != nil {
            return fmt.Errorf("not a git repository (missing %s)", hooksDir)
        }
        preCommit := `#!/usr/bin/env bash
set -e
if command -v mise >/dev/null 2>&1; then
  mise run lint || exit 1
  mise run test || exit 1
else
  golangci-lint run ./... || exit 1
  go test ./... || exit 1
fi
`
        prePush := `#!/usr/bin/env bash
set -e
if command -v mise >/dev/null 2>&1; then
  mise run ci || exit 1
else
  golangci-lint run ./... || exit 1
  go test ./... || exit 1
  go build ./... || exit 1
fi
`
        if err := os.WriteFile(hooksDir+"/pre-commit", []byte(preCommit), 0755); err != nil { return err }
        if err := os.WriteFile(hooksDir+"/pre-push", []byte(prePush), 0755); err != nil { return err }
        fmt.Println("Hooks installed: pre-commit, pre-push")
        return nil
    },
}

func init() {
    hooksCmd.AddCommand(hooksInstallCmd)
}

var execCmd = &cobra.Command{
    Use:   "exec -- <command>",
    Short: "Run a shell command across repositories",
    Long:  "Execute a shell command in each repository. Supports filtering by account and dirty state.",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return fmt.Errorf("provide a command after --")
        }
        cfg, err := config.Load(cfgFile)
        if err != nil { return fmt.Errorf("loading config: %w", err) }
        scanner := scan.New(cfg, workerCount)
        repos, err := scanner.Scan(scanPath)
        if err != nil { return fmt.Errorf("scanning repos: %w", err) }
        if dirtyOnly { repos = filterDirty(repos) }
        if accountFilter != "" { repos = filterByAccount(repos, accountFilter) }
        timeoutSec, _ := cmd.Flags().GetInt("timeout")
        results := runner.ExecInRepos(repos, strings.Join(args, " "), time.Duration(timeoutSec)*time.Second)
        if jsonOutput {
            return ui.PrintJSONResponse(true, results, nil)
        }
        var ok, fail int
        for _, r := range results {
            if r.Success { ok++ } else { fail++ }
        }
        fmt.Printf("Executed in %d repos: %d ok, %d failed\n", len(results), ok, fail)
        if fail > 0 { os.Exit(30) }
        return nil
    },
}

func init() {
    execCmd.Flags().StringVarP(&accountFilter, "account", "a", "", "filter by account")
    execCmd.Flags().BoolVarP(&dirtyOnly, "dirty", "d", false, "only dirty repositories")
    execCmd.Flags().Int("timeout", 0, "timeout in seconds for each command (0=none)")
    execCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
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
