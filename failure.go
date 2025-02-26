package structkeys

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Failure struct {
	Message  string
	Node     ast.Node
	Position token.Position
}

func NewFailure(message string, node ast.Node, fs *token.FileSet) Failure {
	pos := fs.Position(node.Pos())
	return Failure{
		Message:  message,
		Node:     node,
		Position: pos,
	}
}

func (f Failure) String() string {
	return fmt.Sprintf("%s: %s", f.Position, f.Message)
}
