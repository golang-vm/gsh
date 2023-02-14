
package main

import (
	"bufio"
	"errors"
	// "flag"
	"os"
	"go/ast"
	"go/scanner"

	parser "github.com/golang-vm/gsh/parser"
	cmd "github.com/golang-vm/gsh/command"
	ucl "github.com/kmcsr/unixcsl"
)

func main(){
	{
		reseter, err := ucl.MakeRaw(os.Stdin)
		if err != nil {
			panic(err)
		}
		defer reseter()
	}
	bufreader := bufio.NewReader(os.Stdin)
	src := cmd.NewSource(bufreader, os.Stdout)
	csl := ucl.NewConsole(bufreader, os.Stdout)
	csl.Printf("%s\r\n\r\n", "Hi")
	csl.Printf("GSH (go shell) v%s\r\n", VERSION)
	csl.Println("Copyright (C) 2022 <https://github.com/zyxkad>")
	csl.Println("GPLv3.0 LICENSE")
	csl.Println()

	cslw := consoleWrapper{csl}
	scn := parser.NewScanner(cslw)
	main: for {
		cslw.Reset()
		nodes, err := scn.Scan()
		if err != nil {
			switch er := err.(type) {
			case parser.CmdLine:
				err = cmd.ExecuteCommand(src, (string)(er.Line))
				if err != nil {
					if errors.Is(err, cmd.ExitErr) {
						break main
					}
					csl.Println(err)
					break
				}
			case scanner.ErrorList:
				csl.Println("<<< Error list:")
				for i, e := range er {
					csl.Printf("E%d: %v\r\n", i + 1, e)
				}
			case ucl.CBreakErr:
				switch er {
				case 3: // Ctrl-C
					csl.Println("\r\nsignal: SIGINT")
					csl.SetLine(nil)
				case 4: // Ctrl-D
					break main
				case 18: // Ctrl-R
					csl.Println("\r\nTODO: search command")
				default:
					panic((byte)(er))
				}
			default:
				panic(err)
			}
			continue
		}
		if len(nodes) > 0 {
			csl.Println("nodes:", len(nodes))
			for i, n := range nodes {
				csl.Printf("#%d: ", i)
				printNode(csl, n)
				csl.Print("\r\n")
			}
		}
	}
	csl.Println("\r\n~~ Good Bye ~~")
}

type consoleWrapper struct{
	*ucl.Console
}

var _ parser.LineReader = consoleWrapper{}

func (c consoleWrapper)Reset(){
	c.Console.SetPrompt(Prompt)
}

func (c consoleWrapper)ReadLine()(line []byte, err error){
	line, err = c.Console.ReadLine()
	c.Console.SetPrompt(ContinuePrompt)
	return
}

func printNode(csl *ucl.Console, n ast.Node){
	printNode0(csl, n, 0)
}

func printNode0(csl *ucl.Console, n ast.Node, deep int){
	switch n0 := n.(type) {
	case *ast.Ident:
		csl.Printf("%s", n0.Name)
	case *ast.Ellipsis:
		csl.Print("...")
		if n0.Elt != nil {
			printNode0(csl, n0.Elt, deep)
		}
	case *ast.BasicLit:
		csl.Printf("(%s)", n0.Value)
	case *ast.FuncLit:
		printNode0(csl, n0.Type, deep)
		csl.Print(" ")
		printNode0(csl, n0.Body, deep + 1)
	case *ast.CompositeLit:
		if n0.Type != nil {
			printNode0(csl, n0.Type, deep)
		}
		csl.Print("{\r\n")
		for _, e := range n0.Elts {
			printIndent(csl, deep)
			printNode0(csl, e, deep + 1)
			csl.Print(",\r\n")
		}
		printIndent(csl, deep)
		csl.Print("}")
	case *ast.ParenExpr:
		csl.Print("(")
		printNode0(csl, n0.X, deep + 1)
		csl.Print(")")
	case *ast.SelectorExpr:
		// TODO
		csl.Printf("%#v", *n0)
	case *ast.IndexExpr:
		printNode0(csl, n0.X, deep)
		csl.Print("[\r\n")
		printIndent(csl, deep)
		printNode0(csl, n0.Index, deep + 1)
		csl.Print("\r\n")
		printIndent(csl, deep)
		csl.Print("]")
	case *ast.IndexListExpr:
		printNode0(csl, n0.X, deep)
		csl.Print("[\r\n")
		for _, e := range n0.Indices {
		printIndent(csl, deep)
			printNode0(csl, e, deep + 1)
			csl.Print(",\r\n")
		}
		printIndent(csl, deep)
		csl.Print("]")
	case *ast.SliceExpr:
		// TODO
		csl.Printf("%#v", *n0)
	case *ast.TypeAssertExpr:
		// TODO
		csl.Printf("%#v", *n0)
	case *ast.CallExpr:
		printNode0(csl, n0.Fun, deep)
		csl.Print("(")
		if len(n0.Args) != 0 {
			csl.Print("\r\n")
			for _, e := range n0.Args {
				printIndent(csl, deep + 1)
				printNode0(csl, e, deep + 1)
				csl.Print(",\r\n")
			}
		}
		printIndent(csl, deep)
		csl.Print(")")
	case *ast.StarExpr:
		csl.Print("*")
		printIndent(csl, deep)
		printNode0(csl, n0.X, deep)
	case *ast.UnaryExpr:
		csl.Print(n0.Op)
		printIndent(csl, deep)
		printNode0(csl, n0.X, deep)
	case *ast.BinaryExpr:
		csl.Printf("[ %s\r\n", n0.Op.String())
		printIndent(csl, deep + 1)
		printNode0(csl, n0.X, deep + 1)
		csl.Print(",\r\n")
		printIndent(csl, deep + 1)
		printNode0(csl, n0.Y, deep + 1)
		csl.Print("\r\n")
		printIndent(csl, deep)
		csl.Print("]")
	case *ast.KeyValueExpr:
		printNode0(csl, n0.Key, deep)
		csl.Print(": ")
		printNode0(csl, n0.Value, deep)
	case *ast.ArrayType:
		csl.Print("[")
		if n0.Len != nil {
			printNode0(csl, n0.Len, deep + 1)
		}
		csl.Print("]")
		printNode0(csl, n0.Elt, deep + 1)
	case *ast.StructType:
		csl.Print("struct{")
		if n0.Fields != nil && len(n0.Fields.List) != 0 {
			csl.Print("\r\n")
			for _, e := range n0.Fields.List {
				printIndent(csl, deep + 1)
				printNode0(csl, e, deep + 1)
				csl.Print("\r\n")
			}
		}
		csl.Print("}")
	case *ast.FuncType:
		csl.Print("func(")
		if n0.Params != nil && len(n0.Params.List) != 0 {
			csl.Print("\r\n")
			for _, e := range n0.Params.List {
				printIndent(csl, deep + 1)
				printNode0(csl, e, deep + 1)
				csl.Print(",\r\n")
			}
			printIndent(csl, deep)
		}
		csl.Print(")")
		if n0.Params != nil && len(n0.Params.List) != 0 {
			csl.Print(" (")
			for _, e := range n0.Params.List {
				printIndent(csl, deep + 1)
				printNode0(csl, e, deep + 1)
				csl.Print(",\r\n")
			}
			printIndent(csl, deep)
			csl.Print(")")
		}
	case *ast.InterfaceType:
		csl.Print("interface{")
		if n0.Methods != nil && len(n0.Methods.List) != 0 {
			csl.Print("\r\n")
			for _, e := range n0.Methods.List {
				printIndent(csl, deep + 1)
				printNode0(csl, e, deep + 1)
				csl.Print("\r\n")
			}
		}
		csl.Print("}")
	case *ast.MapType:
		csl.Print("map[")
		printNode0(csl, n0.Key, deep + 1)
		csl.Print("]")
		printNode0(csl, n0.Value, deep + 1)
	case *ast.ChanType:
		if n0.Dir == ast.RECV {
			csl.Print("<-")
		}
		csl.Print("chan")
		if n0.Dir == ast.SEND {
			csl.Print("<-")
		}
		csl.Print(" ")
		printNode0(csl, n0.Value, deep + 1)
	case *ast.DeclStmt:
		printNode0(csl, n0.Decl, deep)
	case *ast.LabeledStmt:
		printNode0(csl, n0.Label, deep)
		csl.Print(": ")
		printNode0(csl, n0.Stmt, deep)
	case *ast.ExprStmt:
		printNode0(csl, n0.X, deep)
	default:
		csl.Printf("%#v", n)
	}
}

var indentCache string = "        "

func printIndent(csl *ucl.Console, deep int){
	if deep == 0 {
		return
	}
	for deep > len(indentCache) {
		indentCache += " "
	}
	csl.Print(indentCache[:deep])
}
