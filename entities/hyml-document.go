package entities

import (
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
