package entities

import (
	"gopkg.in/yaml.v3"
)

type HymlDocument struct {
	HymlVersion string                   `yaml:"hyml-version"`
	Def         []map[string]interface{} `yaml:"def"`
	Html        Html                     `yaml:"html"`
	FileName    string
}

type Html struct {
	Title map[string]interface{} `yaml:"title"`
	Body  []yaml.Node            `yaml:"body"`
}
