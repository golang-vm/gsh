
package main

import (
	"bufio"
	"errors"
	// "flag"
	"os"
	"go/scanner"

	parser "github.com/kmcsr/gsh/parser"
	cmd "github.com/kmcsr/gsh/command"
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
	csl.Printf("%s\r\n\r\n", "")
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
				for _, e := range er {
					csl.Println(e)
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
		csl.Println("nodes:", len(nodes))
		for i, n := range nodes {
			csl.Printf("%d: %#v\r\n", i, n)
		}
	}
	csl.Println("\r\n~~ Good Bye ~~")
}

type consoleWrapper struct{
	*ucl.Console
}

var _ parser.LineReader = consoleWrapper{}

func (c consoleWrapper)Reset(){
	c.Console.SetPrompt(PROMPT)
}

func (c consoleWrapper)ReadLine()(line []byte, err error){
	line, err = c.Console.ReadLine()
	c.Console.SetPrompt(CONTINUE_PROMPT)
	return
}
