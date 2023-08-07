package boil

import (
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
)

// PrintTemplate debug prints t nodes.
func PrintTemplate(t *template.Template) { printNode(t.Tree.Root, 0) }

func i(n int) (s string) {
	for i := 0; i < n; i++ {
		s += "  "
	}
	return
}

func printNode(n parse.Node, e int) {
	if n == nil {
		return
	}
	switch t := n.(type) {
	case *parse.TextNode: // Plain text.
		fmt.Printf("%sText: %s\n", i(e), strings.TrimSpace(string(t.Text)))
	case *parse.ActionNode: // A non-control action such as a field evaluation.
		fmt.Printf("%sAction\n", i(e))
		printNode(t.Pipe, e+1)
	case *parse.BoolNode: // A boolean constant.
		fmt.Printf("%sBool: %t\n", i(e), t.True)
	case *parse.ChainNode: // A sequence of field accesses.
		fmt.Printf("%sChain: %v\n", i(e), t.Field)
	case *parse.CommandNode: // An element of a pipeline.
		fmt.Printf("%sCommand\n", i(e))
		for _, v := range t.Args {
			printNode(v, e+1)
		}
	case *parse.DotNode: // The cursor, dot.
		fmt.Printf("%sDot\n", i(e))
	case *parse.FieldNode: // A field or method name.
		fmt.Printf("%sField: %v\n", i(e), t.Ident)
	case *parse.IdentifierNode: // An identifier; always a function name.
		fmt.Printf("%sIdentifier: %v\n", i(e), t.Ident)
	case *parse.IfNode: // An if action.
		fmt.Printf("%sIf\n", i(e))
		printNode(&t.BranchNode, e+1)
	case *parse.BranchNode:
		fmt.Printf("%sBranch\n", i(e))
		printNode(t.Pipe, e+1)
		printNode(t.List, e+1)
		printNode(t.ElseList, e+1)

	case *parse.ListNode: // A list of Nodes.
		fmt.Printf("%sList\n", i(e))
		for _, v := range t.Nodes {
			printNode(v, e+1)
		}
	case *parse.NilNode: // An untyped nil constant.
		fmt.Printf("%sNil\n", i(e))
	case *parse.NumberNode: // A numerical constant.
		fmt.Printf("%sNumber: %s\n", i(e), t.Text)
	case *parse.PipeNode: // A pipeline of commands.
		if t == nil {
			return
		}
		fmt.Printf("%sPipe\n", i(e))
		for _, v := range t.Decl {
			printNode(v, e+1)
		}
		for _, v := range t.Cmds {
			printNode(v, e+1)
		}
	case *parse.RangeNode: // A range action.
		fmt.Printf("%sRange\n", i(e))
	case *parse.StringNode: // A string constant.
		fmt.Printf("%sString: %s\n", i(e), t.Text)
	case *parse.TemplateNode: // A template invocation action.
		fmt.Printf("%sTemplate: %s\n", i(e), t.Name)
		printNode(t.Pipe, e+1)
	case *parse.VariableNode: // A $ variable.
		fmt.Printf("%sVariable: %v\n", i(e), t.Ident)
	case *parse.WithNode: // A with action.
		fmt.Printf("%sWith\n", i(e))
	case *parse.CommentNode: // A comment.
		fmt.Printf("%sComment\n", i(e))
	case *parse.BreakNode: // A break action.
		fmt.Printf("%sBreak\n", i(e))
	case *parse.ContinueNode: // A continue action.
		fmt.Printf("%sContinue\n", i(e))
	}
}
