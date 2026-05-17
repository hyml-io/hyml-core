package entities

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type HymlDocument struct {
	HymlVersion string `yaml:"hyml-version"` // Ignored during normal key reflection
	Def         Def    `yaml:"def"`
	Html        Html   `yaml:"html"`
	FileName    string
}

type Html struct {
	Title string      `yaml:"title"`
	Body  []yaml.Node `yaml:"body"`
}

type Def []map[string]interface{}

// UnmarshalYAML intercepts the raw YAML tree before Go maps it to fields
func (document *HymlDocument) UnmarshalYAML(value *yaml.Node) error {
	// If wrapped in a DocumentNode, dig one level deeper to find the tag
	if value.Kind == yaml.DocumentNode && len(value.Content) > 0 {
		value = value.Content[0]
	}

	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("hyml: expected document mapping block, got kind %d", value.Kind)
	}

	// Create a local alias type to prevent infinite recursion during decoding
	type Alias HymlDocument
	return value.Decode((*Alias)(document))
}
