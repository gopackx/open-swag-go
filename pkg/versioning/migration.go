package versioning

import (
	"fmt"
	"strings"
)

// MigrationGuide represents a migration guide
type MigrationGuide struct {
	FromVersion string
	ToVersion   string
	Steps       []MigrationStep
}

// MigrationStep represents a single migration step
type MigrationStep struct {
	Title       string
	Description string
	Before      string
	After       string
	Endpoint    string
	Method      string
}

// MigrationGenerator generates migration guides
type MigrationGenerator struct{}

// NewMigrationGenerator creates a new migration generator
func NewMigrationGenerator() *MigrationGenerator {
	return &MigrationGenerator{}
}

// Generate creates a migration guide from a diff
func (g *MigrationGenerator) Generate(diff *Diff) *MigrationGuide {
	guide := &MigrationGuide{
		FromVersion: diff.OldVersion,
		ToVersion:   diff.NewVersion,
		Steps:       []MigrationStep{},
	}

	for _, breaking := range diff.Breaking {
		step := g.createMigrationStep(breaking)
		guide.Steps = append(guide.Steps, step)
	}

	return guide
}

func (g *MigrationGenerator) createMigrationStep(breaking BreakingChange) MigrationStep {
	step := MigrationStep{
		Endpoint: breaking.Path,
		Method:   breaking.Method,
	}

	switch {
	case strings.Contains(breaking.Reason, "removed"):
		step.Title = fmt.Sprintf("Handle removed endpoint: %s %s", breaking.Method, breaking.Path)
		step.Description = breaking.Migration
		step.Before = fmt.Sprintf("// Old code using %s %s", breaking.Method, breaking.Path)
		step.After = "// Remove or replace with alternative endpoint"

	case strings.Contains(breaking.Reason, "required"):
		step.Title = fmt.Sprintf("Add required field for: %s %s", breaking.Method, breaking.Path)
		step.Description = breaking.Migration
		step.Before = "// Request without the new required field"
		step.After = "// Add the new required field to your request"

	case strings.Contains(breaking.Reason, "parameter"):
		step.Title = fmt.Sprintf("Update parameters for: %s %s", breaking.Method, breaking.Path)
		step.Description = breaking.Migration
		step.Before = "// Old parameter usage"
		step.After = "// Updated parameter usage"

	default:
		step.Title = fmt.Sprintf("Update: %s %s", breaking.Method, breaking.Path)
		step.Description = breaking.Migration
	}

	return step
}

// ToMarkdown converts migration guide to markdown
func (g *MigrationGuide) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Migration Guide: %s → %s\n\n", g.FromVersion, g.ToVersion))

	if len(g.Steps) == 0 {
		sb.WriteString("No breaking changes. No migration required.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("This guide covers %d breaking change(s) that require updates.\n\n", len(g.Steps)))

	for i, step := range g.Steps {
		sb.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, step.Title))
		sb.WriteString(fmt.Sprintf("**Endpoint:** `%s %s`\n\n", step.Method, step.Endpoint))
		sb.WriteString(fmt.Sprintf("%s\n\n", step.Description))

		if step.Before != "" {
			sb.WriteString("**Before:**\n```\n")
			sb.WriteString(step.Before)
			sb.WriteString("\n```\n\n")
		}

		if step.After != "" {
			sb.WriteString("**After:**\n```\n")
			sb.WriteString(step.After)
			sb.WriteString("\n```\n\n")
		}
	}

	return sb.String()
}
