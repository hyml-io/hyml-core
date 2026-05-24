package entities

type PartialDocument struct {
	Template string  `yaml:"template"`
	Content  Content `yaml:"content"`
	Args     RawArgs `yaml:"args"`
}

type Content map[string]interface{}

type RawArgs []map[string]interface{}
