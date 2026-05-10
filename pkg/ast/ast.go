package ast

type NodeType int

const (
	ElementNode NodeType = iota
	TextNode
)

type Node struct {
	Type        NodeType `json:"type"`
	Tag         string   `json:"tag"`
	ParentClass string   `json:"parent_class,omitempty"`
	Value       string   `json:"value,omitempty"`
	Children    []*Node  `json:"children,omitempty"`
}
