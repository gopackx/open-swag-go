package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gopackx/open-swag-go/pkg/versioning"
)

func main() {
	// Example: Compare two OpenAPI spec files
	if len(os.Args) < 3 {
		fmt.Println("Usage: version-diff <old-spec.json> <new-spec.json>")
		fmt.Println("\nThis tool compares two OpenAPI specifications and detects:")
		fmt.Println("  - Added endpoints")
		fmt.Println("  - Removed endpoints (breaking)")
		fmt.Println("  - Modified endpoints")
		fmt.Println("  - Breaking changes with migration guides")

		// Demo with inline specs
		fmt.Println("\n--- Demo Mode ---")
		runDemo()
		return
	}

	oldPath := os.Args[1]
	newPath := os.Args[2]

	differ := versioning.NewDiffer()
	diff, err := differ.CompareFiles(oldPath, newPath)
	if err != nil {
		log.Fatalf("Error comparing specs: %v", err)
	}

	printDiff(diff)
}

func runDemo() {
	oldSpec := map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":   "Demo API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List users",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "OK"},
					},
				},
				"post": map[string]interface{}{
					"summary": "Create user",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"name"},
								},
							},
						},
					},
				},
			},
			"/users/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get user",
					"parameters": []interface{}{
						map[string]interface{}{"name": "id", "in": "path", "required": true},
					},
				},
				"delete": map[string]interface{}{
					"summary": "Delete user",
				},
			},
		},
	}

	newSpec := map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":   "Demo API",
			"version": "2.0.0",
		},
		"paths": map[string]interface{}{
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List users",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "OK"},
					},
				},
				"post": map[string]interface{}{
					"summary": "Create user",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"name", "email"}, // Added required field
								},
							},
						},
					},
				},
			},
			"/users/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get user",
					"parameters": []interface{}{
						map[string]interface{}{"name": "id", "in": "path", "required": true},
					},
				},
				// DELETE removed - breaking change!
			},
			"/products": map[string]interface{}{ // New endpoint
				"get": map[string]interface{}{
					"summary": "List products",
				},
			},
		},
	}

	differ := versioning.NewDiffer()
	diff, err := differ.Compare(oldSpec, newSpec)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	printDiff(diff)
}

func printDiff(diff *versioning.Diff) {
	fmt.Printf("\n=== API Version Diff ===\n")
	fmt.Printf("Old Version: %s\n", diff.OldVersion)
	fmt.Printf("New Version: %s\n", diff.NewVersion)

	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Added endpoints:    %d\n", diff.Summary.AddedEndpoints)
	fmt.Printf("Removed endpoints:  %d\n", diff.Summary.RemovedEndpoints)
	fmt.Printf("Modified endpoints: %d\n", diff.Summary.ModifiedEndpoints)
	fmt.Printf("Breaking changes:   %d\n", diff.Summary.BreakingChanges)

	if len(diff.Changes) > 0 {
		fmt.Printf("\n--- Changes ---\n")
		for _, change := range diff.Changes {
			icon := "+"
			if change.Type == versioning.ChangeRemoved {
				icon = "-"
			} else if change.Type == versioning.ChangeModified {
				icon = "~"
			}

			breaking := ""
			if change.IsBreaking {
				breaking = " [BREAKING]"
			}

			fmt.Printf("%s %s%s\n", icon, change.Description, breaking)
		}
	}

	if len(diff.Breaking) > 0 {
		fmt.Printf("\n--- Breaking Changes & Migration Guide ---\n")
		for _, bc := range diff.Breaking {
			fmt.Printf("\n[%s %s]\n", bc.Method, bc.Path)
			fmt.Printf("  Reason: %s\n", bc.Reason)
			fmt.Printf("  Migration: %s\n", bc.Migration)
		}
	}

	// Generate changelog
	changelog := versioning.GenerateChangelog(diff)
	fmt.Printf("\n--- Generated Changelog ---\n%s\n", changelog)

	// Output as JSON
	fmt.Printf("\n--- JSON Output ---\n")
	jsonData, _ := json.MarshalIndent(diff, "", "  ")
	fmt.Println(string(jsonData))
}
