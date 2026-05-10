package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hyml-io/hyml-core/internal/lexer"
	"github.com/hyml-io/hyml-core/pkg/parser"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	filePath := os.Args[2]

	_ = parser.LoadConfigs("config/html_tags.json", "config/hyml_core.json")

	switch command {
	case "validate":
		runValidate(filePath)
	case "to-html":
		runToHTML(filePath)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: hyml <command> <file>\nCommands: validate, to-html")
}

func runValidate(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Read error: %v", err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseHYML()

	fmt.Printf("✅ %s parsed successfully.\n", path)
	fmt.Println("\n--- Structural Tracing ---")

	seen := make(map[string]bool)

	for _, node := range program.Children {
		if node.Tag == "" || node.Tag == "root" || node.Tag == "#text" {
			continue
		}

		uniqueKey := fmt.Sprintf("%s|%s", node.Tag, node.ParentClass)
		if seen[uniqueKey] {
			continue
		}
		seen[uniqueKey] = true

		var displayTag string
		var displayType string

		// Get the ancestry string from the registry
		lineage := parser.GetLineage(node.Tag)

		if node.ParentClass != "" {
			// It's a Definition (Set)
			displayTag = fmt.Sprintf("%s:%s", node.Tag, node.ParentClass)

			// For a definition, the "lineage" in the registry already includes ParentClass.
			// We just want to show what the ParentClass itself points to.
			parentLineage := parser.GetLineage(node.ParentClass)
			if parentLineage != "" {
				displayType = fmt.Sprintf("HYML-Derivation Set (%s:%s)", node.ParentClass, parentLineage)
			} else {
				displayType = fmt.Sprintf("HYML-Derivation Set (%s)", node.ParentClass)
			}
		} else if lineage != "" {
			// It's a Usage (Derived)
			displayTag = node.Tag
			displayType = fmt.Sprintf("HYML-Derived (%s)", lineage)
		} else {
			// Standard/Core with no inheritance
			displayTag = node.Tag
			displayType = parser.GetTagType(node.Tag)
		}

		fmt.Printf("Tag: %-40s | Type: %s\n", displayTag, displayType)
	}
}

func runToHTML(path string) {
	fmt.Printf("🔨 Compiling %s to HTML...\n", path)
	// Integration with RenderHTML goes here
}
