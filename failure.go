package structkeys

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Failure struct {
	Message string
	Node    *ast.CompositeLit
}

func (f Failure) Information(fs *token.FileSet) string {
	pos := fs.Position(f.Node.Pos())
	return fmt.Sprintf("%s: %s", pos, f.Message)
}
