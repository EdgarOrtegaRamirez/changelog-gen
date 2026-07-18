package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// Section represents a changelog section (e.g., "Added", "Fixed")
type Section struct {
	Name    string
	Entries []string
}

// Format represents the output format
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatJSON     Format = "json"
	FormatText     Format = "text"
)

// CommitType maps git conventional commit types to changelog sections
var commitTypeMap = map[string]struct {
	section string
	icon    string
}{
	"feat":        {"Added", "✨"},
	"feature":     {"Added", "✨"},
	"fix":         {"Fixed", "🐛"},
	"bugfix":      {"Fixed", "🐛"},
	"perf":        {"Performance", "⚡"},
	"performance": {"Performance", "⚡"},
	"refactor":    {"Refactored", "♻️"},
	"chore":       {"Chores", "🔧"},
	"docs":        {"Documentation", "📝"},
	"style":       {"Style", "🎨"},
	"test":        {"Tests", "🧪"},
	"ci":          {"CI/CD", "🤖"},
	"build":       {"Build", "📦"},
	"revert":      {"Reverted", "⏪"},
	"breaking":    {"Breaking Changes", "💥"},
	"deprecate":   {"Deprecated", "⚠️"},
}

// parseCommitType extracts the conventional commit type from a message
func parseCommitType(msg string) (string, string) {
	// Match conventional commit format: type(scope): description
	re := regexp.MustCompile(`^(\w+)(\([^)]*\))?:\s*(.+)$`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) >= 3 {
		return matches[1], matches[3]
	}
	return "", msg
}

// getSectionName returns the section name for a commit type
func getSectionName(commitType string) (string, bool) {
	entry, ok := commitTypeMap[commitType]
	if ok {
		return entry.section, true
	}
	return "", false
}

// getGitLog runs git log and returns commit messages
func getGitLog(args ...string) ([]string, error) {
	cmd := exec.Command("git", append([]string{"log", "--pretty=format:%s"}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	var commits []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commits = append(commits, line)
		}
	}
	return commits, scanner.Err()
}

// getTagName returns the latest tag
func getTagName() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getGitTag returns a specific tag
func getGitTag(tag string) (string, error) {
	cmd := exec.Command("git", "tag", "-l", tag)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getGitDate returns the commit date
func getGitDate(commitHash string) string {
	cmd := exec.Command("git", "log", "-1", "--format=%ci", commitHash)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// generateChangelog generates a changelog from git history
func generateChangelog(fromTag, toTag, version string) ([]Section, error) {
	// Determine range
	var rangeArgs []string
	if fromTag != "" {
		rangeArgs = []string{fmt.Sprintf("%s..%s", fromTag, toTag)}
	} else {
		rangeArgs = []string{toTag}
	}

	commits, err := getGitLog(rangeArgs...)
	if err != nil {
		return nil, err
	}

	// Group commits by section
	sections := make(map[string]*Section)
	var sectionOrder []string

	// Breaking changes list
	var breakingChanges []string

	for _, msg := range commits {
		// Skip merge commits
		if strings.HasPrefix(msg, "Merge ") {
			continue
		}

		// Check for breaking changes
		if strings.Contains(msg, "!") || strings.Contains(strings.ToLower(msg), "breaking") {
			breakingChanges = append(breakingChanges, msg)
			continue
		}

		commitType, description := parseCommitType(msg)
		sectionName, ok := getSectionName(commitType)
		if !ok {
			// Unknown type — put in "Other"
			sectionName = "Other"
			ok = true
		}

		if _, exists := sections[sectionName]; !exists {
			sections[sectionName] = &Section{Name: sectionName, Entries: []string{}}
			sectionOrder = append(sectionOrder, sectionName)
		}

		// Clean up description
		desc := strings.TrimSpace(description)
		if desc == "" {
			desc = strings.TrimSpace(msg)
		}
		sections[sectionName].Entries = append(sections[sectionName].Entries, desc)
	}

	// Sort sections in a defined order
	sectionOrderMap := map[string]int{
		"Breaking Changes": 0,
		"Added":            1,
		"Fixed":            2,
		"Performance":      3,
		"Refactored":       4,
		"Documentation":    5,
		"Style":            6,
		"Tests":            7,
		"CI/CD":            8,
		"Build":            9,
		"Chores":           10,
		"Reverted":         11,
		"Deprecated":       12,
		"Other":            13,
	}

	sort.Slice(sectionOrder, func(i, j int) bool {
		a, okA := sectionOrderMap[sectionOrder[i]]
		b, okB := sectionOrderMap[sectionOrder[j]]
		if !okA {
			a = 99
		}
		if !okB {
			b = 99
		}
		return a < b
	})

	var result []Section
	for _, name := range sectionOrder {
		result = append(result, *sections[name])
	}

	return result, nil
}

// renderMarkdown renders sections as markdown
func renderMarkdown(sections []Section, version, date string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Changelog\n\n"))
	if version != "" {
		sb.WriteString(fmt.Sprintf("## [%s] — %s\n\n", version, date))
	} else {
		sb.WriteString(fmt.Sprintf("## %s\n\n", date))
	}

	for _, section := range sections {
		sb.WriteString(fmt.Sprintf("### %s\n\n", section.Name))
		for _, entry := range section.Entries {
			sb.WriteString(fmt.Sprintf("- %s\n", entry))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderJSON renders sections as JSON
func renderJSON(sections []Section, version, date string) string {
	type Output struct {
		Version  string    `json:"version"`
		Date     string    `json:"date"`
		Sections []Section `json:"sections"`
	}
	output := Output{
		Version:  version,
		Date:     date,
		Sections: sections,
	}

	// Simple JSON render
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf(`  "version": %q,
`, output.Version))
	sb.WriteString(fmt.Sprintf(`  "date": %q,
`, output.Date))
	sb.WriteString(`  "sections": [
`)
	for i, section := range output.Sections {
		sb.WriteString(fmt.Sprintf(`    {
      "name": %q,
      "entries": [
`, section.Name))
		for j, entry := range section.Entries {
			escaped := strings.ReplaceAll(entry, `"`, `\"`)
			if j < len(section.Entries)-1 {
				sb.WriteString(fmt.Sprintf(`        %q,
`, escaped))
			} else {
				sb.WriteString(fmt.Sprintf(`        %q
`, escaped))
			}
		}
		sb.WriteString(`      ]
    }`)
		if i < len(output.Sections)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(`  ]
}
`)
	return sb.String()
}

// renderText renders sections as plain text
func renderText(sections []Section, version, date string) string {
	var sb strings.Builder

	if version != "" {
		sb.WriteString(fmt.Sprintf("Changelog %s (%s)\n", version, date))
		sb.WriteString(strings.Repeat("=", len(sb.String())) + "\n\n")
	}

	for _, section := range sections {
		sb.WriteString(fmt.Sprintf("%s:\n", section.Name))
		for _, entry := range section.Entries {
			sb.WriteString(fmt.Sprintf("  - %s\n", entry))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func main() {
	// Simple CLI: changelog-gen [--from <tag>] [--to <tag>] [--version <ver>] [--format <format>] [--output <file>]
	var fromTag, toTag, version, format, output string

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--from":
			if i+1 < len(os.Args) {
				fromTag = os.Args[i+1]
				i++
			}
		case "--to":
			if i+1 < len(os.Args) {
				toTag = os.Args[i+1]
				i++
			}
		case "--version":
			if i+1 < len(os.Args) {
				version = os.Args[i+1]
				i++
			}
		case "--format":
			if i+1 < len(os.Args) {
				format = os.Args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(os.Args) {
				output = os.Args[i+1]
				i++
			}
		case "--help", "-h":
			fmt.Println("Usage: changelog-gen [options]")
			fmt.Println("")
			fmt.Println("Options:")
			fmt.Println("  --from <tag>     Generate from this tag (default: latest tag)")
			fmt.Println("  --to <tag>       Generate up to this tag (default: HEAD)")
			fmt.Println("  --version <ver>  Version label for the changelog entry")
			fmt.Println("  --format <fmt>   Output format: markdown, json, text (default: markdown)")
			fmt.Println("  --output <file>  Write to file instead of stdout")
			fmt.Println("  --help, -h       Show this help message")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  changelog-gen")
			fmt.Println("  changelog-gen --from v1.0.0 --to v2.0.0 --version 2.0.0")
			fmt.Println("  changelog-gen --version 1.1.0 --format json")
			return
		}
	}

	// Default format
	if format == "" {
		format = "markdown"
	}

	// Determine toTag
	if toTag == "" {
		toTag = "HEAD"
	}

	// Determine date
	cmd := exec.Command("git", "log", "-1", "--format=%ci", toTag)
	out, err := cmd.Output()
	var date string
	if err == nil {
		date = strings.TrimSpace(string(out))
	} else {
		date = "unknown"
	}

	// Generate changelog
	sections, err := generateChangelog(fromTag, toTag, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(sections) == 0 {
		fmt.Println("(No changes found)")
		return
	}

	// Render output
	var rendered string
	switch Format(format) {
	case FormatJSON:
		rendered = renderJSON(sections, version, date)
	case FormatText:
		rendered = renderText(sections, version, date)
	default:
		rendered = renderMarkdown(sections, version, date)
	}

	// Output
	if output != "" {
		err = os.WriteFile(output, []byte(rendered), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Changelog written to %s\n", output)
	} else {
		fmt.Print(rendered)
	}
}
