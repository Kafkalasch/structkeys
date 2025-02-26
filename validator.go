package structkeys

import (
	"fmt"
	"go/ast"
	"go/types"
)

type Validator struct {
	Info      *types.Info
	OnFailure func(Failure)
}

func NewValidator(info *types.Info, onFailure func(failure Failure)) (*Validator, error) {
	if info == nil {
		return nil, fmt.Errorf("info is nil")
	}
	if info.Defs == nil {
		return nil, fmt.Errorf("info.Defs is nil")
	}
	if info.Uses == nil {
		return nil, fmt.Errorf("info.Uses is nil")
	}
	if onFailure == nil {
		return nil, fmt.Errorf("onFailure is nil")
	}
	val := &Validator{
		Info:      info,
		OnFailure: onFailure,
	}
	return val, nil
}

func (v *Validator) Visit(node ast.Node) ast.Visitor {
	structInit, ok := v.asStructCompositeLit(node)
	if !ok {
		return v
	}

	elements := structInit.Elts
	if len(elements) == 0 {
		// empty struct initialization
		return v
	}

	// it is enough to check the first one
	// either all of them are named or none of them
	element := elements[0]
	_, isNamedFieldElement := element.(*ast.KeyValueExpr)
	if isNamedFieldElement {
		return v
	}

	failure := Failure{
		Message: "struct literals must use keys during initialization",
		Node:    structInit,
	}
	v.OnFailure(failure)
	return v
}

func (v *Validator) asStructCompositeLit(n ast.Node) (*ast.CompositeLit, bool) {
	compLit, ok := n.(*ast.CompositeLit)
	if !ok {
		return nil, false
	}

	// anonymous struct initialization
	_, ok = compLit.Type.(*ast.StructType)
	if ok {
		return compLit, true
	}

	// named struct initialization
	ident, ok := compLit.Type.(*ast.Ident)
	if !ok {
		return nil, false
	}

	// get referenced object
	object := v.Info.ObjectOf(ident)
	_, isStructType := object.Type().Underlying().(*types.Struct)

	if !isStructType {
		return nil, false
	}
	return compLit, true
}
