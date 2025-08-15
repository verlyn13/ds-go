package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/verlyn13/ds-go/internal/config"
	"github.com/verlyn13/ds-go/internal/scan"
)

// Native ANSI color codes for maximum performance
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorBold   = "\033[1m"
)

// Styles using lipgloss for consistent theming
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1)
	
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("7"))
	
	cleanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
	
	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))
	
	aheadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12"))
	
	behindStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14"))
	
	stashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("13"))
	
	fetchWarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

// PrintTable renders repositories in a formatted table - optimized for speed
func PrintTable(repos []scan.Repository, cfg *config.Config) error {
	if len(repos) == 0 {
		fmt.Println("No repositories found")
		return nil
	}

	// Group by account for organized output
	grouped := groupByAccount(repos)
	
	// Statistics - single pass
	var total, clean, dirty, ahead, behind int
	for _, repo := range repos {
		total++
		if repo.IsClean && repo.Ahead == 0 && repo.Behind == 0 {
			clean++
		}
		if !repo.IsClean {
			dirty++
		}
		if repo.Ahead > 0 {
			ahead++
		}
		if repo.Behind > 0 {
			behind++
		}
	}
	
	// Print header with stats
	header := fmt.Sprintf("📊 Repository Status: %d total | %s%d clean%s | %s%d changes%s | %s%d ahead%s | %s%d behind%s",
		total,
		ColorGreen, clean, ColorReset,
		ColorYellow, dirty, ColorReset,
		ColorBlue, ahead, ColorReset,
		ColorCyan, behind, ColorReset)
	
	fmt.Println(titleStyle.Render(header))
	
	// Create table with minimal styling for performance
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = false
	t.Style().Options.DrawBorder = false
	
	// Print each account group
	for account, accountRepos := range grouped {
		if len(grouped) > 1 {
			fmt.Printf("\n%s%s%s (%d)\n", ColorBold, account, ColorReset, len(accountRepos))
		}
		
		t.ResetHeaders()
		t.ResetRows()
		t.AppendHeader(table.Row{"", "Repository", "Status", "Changes", "Sync", "Last Commit"})
		
		for _, repo := range accountRepos {
			t.AppendRow(formatRepoRow(repo))
		}
		
		t.Render()
	}
	
	// Show fetch hint if needed
	showFetchHint(repos)
	
	return nil
}

// formatRepoRow formats a single repository row - optimized with direct string building
func formatRepoRow(repo scan.Repository) table.Row {
	// Status icon
	var icon string
	switch {
	case !repo.IsClean:
		icon = ColorYellow + "●" + ColorReset
	case repo.Ahead > 0:
		icon = ColorBlue + "↑" + ColorReset
	case repo.Behind > 0:
		icon = ColorCyan + "↓" + ColorReset
	default:
		icon = ColorGreen + "✓" + ColorReset
	}
	
	// Repository name
	name := repo.Name
	if len(name) > 30 {
		name = name[:27] + "..."
	}
	
	// Status column
	status := "clean"
	if !repo.IsClean {
		status = fmt.Sprintf("%s%d files%s", ColorYellow, repo.Uncommitted, ColorReset)
	}
	
	// Changes column
	var changes []string
	if repo.HasStash {
		changes = append(changes, ColorPurple+"stash"+ColorReset)
	}
	changeStr := strings.Join(changes, " ")
	
	// Sync column - compact display
	var sync string
	if repo.Ahead > 0 || repo.Behind > 0 {
		if repo.Ahead > 0 {
			sync += fmt.Sprintf("%s↑%d%s", ColorBlue, repo.Ahead, ColorReset)
		}
		if repo.Behind > 0 {
			if sync != "" {
				sync += " "
			}
			sync += fmt.Sprintf("%s↓%d%s", ColorCyan, repo.Behind, ColorReset)
		}
	} else if !repo.HasUpstream {
		sync = ColorGray + "no upstream" + ColorReset
	} else {
		sync = "synced"
	}
	
	// Last commit - truncate for display
	lastCommit := repo.LastCommit
	if len(lastCommit) > 40 {
		lastCommit = lastCommit[:37] + "..."
	}
	
	// Add fetch warning if stale
	if repo.LastFetch != nil {
		age := time.Since(*repo.LastFetch)
		if age > time.Hour {
			hours := int(age.Hours())
			lastCommit += fmt.Sprintf(" %s(%dh old)%s", ColorGray, hours, ColorReset)
		}
	}
	
	return table.Row{icon, name, status, changeStr, sync, lastCommit}
}

// groupByAccount groups repositories by account
func groupByAccount(repos []scan.Repository) map[string][]scan.Repository {
	grouped := make(map[string][]scan.Repository)
	for _, repo := range repos {
		grouped[repo.Account] = append(grouped[repo.Account], repo)
	}
	return grouped
}

// showFetchHint shows a hint if repositories need fetching
func showFetchHint(repos []scan.Repository) {
	var needsFetch int
	for _, repo := range repos {
		if repo.LastFetch == nil || time.Since(*repo.LastFetch) > 24*time.Hour {
			needsFetch++
		}
	}
	
	if needsFetch > 0 {
		fmt.Printf("\n%sTip:%s %d repositories need updating. Run '%sds fetch%s' to update remote info.\n",
			ColorYellow, ColorReset, needsFetch, ColorBold, ColorReset)
	}
}

// PrintJSON outputs repositories as JSON
func PrintJSON(repos []scan.Repository) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(repos)
}

// PrintFetchResults prints fetch operation results
func PrintFetchResults(results []scan.FetchResult) {
	var succeeded, failed int
	var totalDuration time.Duration
	
	for _, r := range results {
		if r.Success {
			succeeded++
		} else {
			failed++
		}
		totalDuration += r.Duration
	}
	
	fmt.Printf("\n%sFetch complete:%s %d succeeded, %d failed in %.1fs\n",
		ColorBold, ColorReset, succeeded, failed, totalDuration.Seconds())
}