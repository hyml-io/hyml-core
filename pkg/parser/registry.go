package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	derivationChain = make(map[string]string)
	htmlTags        = make(map[string]bool)
	hymlCoreTags    = make(map[string]bool)
)

func RegisterDerivation(child string, parent string) {
	derivationChain[child] = parent
}

func GetParent(tagName string) string {
	return derivationChain[tagName]
}

// GetOriginalTemplate finds the first non-template parent in the chain
func GetOriginalTemplate(tagName string) string {
	current := tagName
	for {
		p := derivationChain[current]
		if p == "" || p == "template" {
			return current
		}
		current = p
	}
}

func GetTagType(tagName string) string {
	if hymlCoreTags[tagName] {
		return "HYML-Core"
	}
	if derivationChain[tagName] != "" {
		return "HYML-Derived"
	}
	if htmlTags[tagName] {
		return "HTML-Standard"
	}
	return "UNKNOWN/ERROR"
}

// GetLineage returns the full chain of parents for a tag (e.g., "apartament:template")
func GetLineage(tagName string) string {
	p := derivationChain[tagName]
	if p == "" {
		return ""
	}

	// Recursive step: see if the parent has its own parent
	ancestorLineage := GetLineage(p)
	if ancestorLineage != "" {
		return fmt.Sprintf("%s:%s", p, ancestorLineage)
	}

	return p
}

// GetDerivationString returns "tag:parent" if parent exists, otherwise just "tag"
func GetDerivationString(tagName string) string {
	p := derivationChain[tagName]
	if p == "" {
		return tagName
	}
	// Note: We don't recurse infinitely here to keep the trace readable,
	// just showing the immediate relationship defined in the registry.
	return fmt.Sprintf("%s:%s", tagName, p)
}

func LoadConfigs(htmlJson, coreJson string) error {
	load := func(path string, target map[string]bool) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var wrapper struct {
			Tags []string `json:"tags"`
		}
		if err := json.Unmarshal(data, &wrapper); err == nil {
			for _, t := range wrapper.Tags {
				target[t] = true
			}
		}
	}
	load(htmlJson, htmlTags)
	load(coreJson, hymlCoreTags)
	return nil
}
