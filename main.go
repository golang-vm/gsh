
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
		printIndent(csl, deep)
		csl.Printf(" %s ", n0.Name)
	case *ast.BasicLit:
		printIndent(csl, deep)
		csl.Printf("(%s)", n0.Value)
	case *ast.FuncLit:
		printNode0(csl, n0.Type, deep)
		csl.Print(" ")
		printNode0(csl, n0.Body, deep + 1)
	case *ast.CompositeLit:
		if n0.Type != nil {
			printNode0(csl, n0.Type, deep)
		}else{
			printIndent(csl, deep)
		}
		csl.Print("{\r\n")
		for _, e := range n0.Elts {
			printNode0(csl, e, deep + 1)
			csl.Print(",\r\n")
		}
		printIndent(csl, deep)
		csl.Print("}")
	case *ast.ParenExpr:
		printIndent(csl, deep)
		csl.Print("(\r\n")
		printNode0(csl, n0.X, deep + 1)
		csl.Print("\r\n")
		printIndent(csl, deep)
		csl.Print(")")
	case *ast.SelectorExpr:
		// TODO
		printIndent(csl, deep)
		csl.Printf("%#v", *n0)
	case *ast.IndexExpr:
		printNode0(csl, n0.X, deep)
		csl.Print("[\r\n")
		printNode0(csl, n0.Index, deep + 1)
		csl.Print("\r\n")
		printIndent(csl, deep)
		csl.Print("]")
	case *ast.IndexListExpr:
		printNode0(csl, n0.X, deep)
		csl.Print("[\r\n")
		for _, e := range n0.Indices {
			printNode0(csl, e, deep + 1)
			csl.Print(",\r\n")
		}
		printIndent(csl, deep)
		csl.Print("]")
	case *ast.SliceExpr:
		// TODO
		printIndent(csl, deep)
		csl.Printf("%#v", *n0)
	case *ast.TypeAssertExpr:
		// TODO
		printIndent(csl, deep)
		csl.Printf("%#v", *n0)
	case *ast.CallExpr:
		printNode0(csl, n0.Fun, deep)
		csl.Print("(\r\n")
		for _, e := range n0.Args {
			printNode0(csl, e, deep + 1)
			csl.Print(",\r\n")
		}
		printIndent(csl, deep)
		csl.Print(")")
	case *ast.StarExpr:
		// TODO
		printIndent(csl, deep)
		csl.Printf("%#v", *n0)
	case *ast.UnaryExpr:
		// TODO
		printIndent(csl, deep)
		csl.Printf("%#v", *n0)
	case *ast.BinaryExpr:
		printIndent(csl, deep)
		csl.Printf("[ %s\r\n", n0.Op.String())
		printNode0(csl, n0.X, deep + 1)
		csl.Print(",\r\n")
		printNode0(csl, n0.Y, deep + 1)
		csl.Print("\r\n")
		printIndent(csl, deep)
		csl.Print("]")
	case *ast.ExprStmt:
		printNode0(csl, n0.X, deep)
	default:
		printIndent(csl, deep)
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
