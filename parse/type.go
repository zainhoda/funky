package parse

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/faiface/funky/types"
)

func IsConstructor(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r) || !unicode.IsLetter(r)
}

func Type(tokens []Token) (types.Type, error) {
	tree, err := MultiTree(tokens)
	if err != nil {
		return nil, err
	}
	return TreeToType(tree)
}

func TreeToType(tree Tree) (types.Type, error) {
	if tree == nil {
		return nil, nil
	}

	switch tree := tree.(type) {
	case *Literal:
		if IsConstructor(tree.Value) {
			return &types.Appl{SI: tree.SI, Cons: tree.Value}, nil
		}
		return &types.Var{SI: tree.SI, Name: tree.Value}, nil

	case *Paren:
		switch tree.Type {
		case "(":
			return TreeToType(tree.Inside)
		}
		return nil, &Error{tree.SI, fmt.Sprintf("unexpected: %s", tree.Type)}

	case *Special:
		return nil, &Error{tree.SI, fmt.Sprintf("unexpected: %s", tree.Type)}

	case *Lambda:
		return nil, &Error{tree.SI, fmt.Sprintf("unexpected: %s", tree.Type)}

	case *Prefix:
		left, err := TreeToType(tree.Left)
		if err != nil {
			return nil, err
		}
		leftAppl, ok := left.(*types.Appl)
		if !ok {
			return nil, &Error{
				left.SourceInfo(),
				fmt.Sprintf("not a type constructor: %v", left),
			}
		}
		right, err := TreeToType(tree.Right)
		leftAppl.Args = append(leftAppl.Args, right)
		return leftAppl, nil

	case *Infix:
		in, err := TreeToType(tree.In)
		if err != nil {
			return nil, err
		}
		left, err := TreeToType(tree.Left)
		if err != nil {
			return nil, err
		}
		right, err := TreeToType(tree.Right)
		if err != nil {
			return nil, err
		}
		if left == nil || right == nil {
			return nil, &Error{
				in.SourceInfo(),
				"missing operand in infix constructor",
			}
		}
		inAppl, ok := in.(*types.Appl)
		if !ok || inAppl.Cons != "->" || len(inAppl.Args) != 0 {
			return nil, &Error{
				left.SourceInfo(),
				fmt.Sprintf("not a type constructor: %v", in),
			}
		}
		return &types.Func{
			From: left,
			To:   right,
		}, nil
	}

	panic("unreachable")
}
