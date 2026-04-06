package versioning

import (
	"fmt"
	"strings"
	"time"
)

// ChangelogEntry represents a changelog entry
type ChangelogEntry struct {
	Version  string
	Date     time.Time
	Added    []string
	Changed  []string
	Removed  []string
	Fixed    []string
	Breaking []string
}

// ChangelogGenerator generates changelogs from diffs
type ChangelogGenerator struct{}

// NewChangelogGenerator creates a new changelog generator
func NewChangelogGenerator() *ChangelogGenerator {
	return &ChangelogGenerator{}
}

// Generate creates a changelog entry from a diff
func (g *ChangelogGenerator) Generate(diff *Diff) *ChangelogEntry {
	entry := &ChangelogEntry{
		Version:  diff.NewVersion,
		Date:     time.Now(),
		Added:    []string{},
		Changed:  []string{},
		Removed:  []string{},
		Breaking: []string{},
	}

	for _, change := range diff.Changes {
		switch change.Type {
		case ChangeAdded:
			entry.Added = append(entry.Added, change.Description)
		case ChangeRemoved:
			entry.Removed = append(entry.Removed, change.Description)
			if change.IsBreaking {
				entry.Breaking = append(entry.Breaking, change.Description)
			}
		case ChangeModified:
			entry.Changed = append(entry.Changed, change.Description)
			if change.IsBreaking {
				entry.Breaking = append(entry.Breaking, change.Description)
			}
		}
	}

	return entry
}

// ToMarkdown converts changelog entry to markdown
func (e *ChangelogEntry) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## [%s] - %s\n\n", e.Version, e.Date.Format("2006-01-02")))

	if len(e.Breaking) > 0 {
		sb.WriteString("### ⚠️ Breaking Changes\n\n")
		for _, item := range e.Breaking {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(e.Added) > 0 {
		sb.WriteString("### Added\n\n")
		for _, item := range e.Added {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(e.Changed) > 0 {
		sb.WriteString("### Changed\n\n")
		for _, item := range e.Changed {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(e.Removed) > 0 {
		sb.WriteString("### Removed\n\n")
		for _, item := range e.Removed {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateChangelog creates a markdown changelog from a diff
func GenerateChangelog(diff *Diff) string {
	gen := NewChangelogGenerator()
	entry := gen.Generate(diff)
	return entry.ToMarkdown()
}
