package entities

type PartialDocument struct {
	Template string  `yaml:"template"`
	Content  Content `yaml:"content"`
}

type Content map[string]interface{}
