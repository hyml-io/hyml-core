package parser

import (
	"fmt"
	"strings"

	"github.com/hyml-io/hyml-core/pkg/ast"
)

func RenderHTML(node *ast.Node) string {
	if node.Tag == "root" {
		var sb strings.Builder
		for _, child := range node.Children {
			sb.WriteString(RenderHTML(child))
		}
		return sb.String()
	}

	// Special case: DOCTYPE
	if node.Tag == "DOCTYPE" {
		return fmt.Sprintf("<!DOCTYPE html %s>\n", node.ParentClass)
	}

	// Normal HTML tags
	opening := fmt.Sprintf("<%s>", node.Tag)
	closing := fmt.Sprintf("</%s>", node.Tag)

	var childrenHTML strings.Builder
	for _, child := range node.Children {
		childrenHTML.WriteString(RenderHTML(child))
	}

	return opening + childrenHTML.String() + closing + "\n"
}
