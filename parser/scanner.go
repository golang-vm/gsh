
package parser

import (
	"go/ast"
	"go/token"
)

type CmdLine struct{
	Line []byte
}
func (CmdLine)Error()(string){ return "" }

type LineReader interface{
	ReadLine()([]byte, error)
}

type Scanner struct{
	r LineReader
	fileset *token.FileSet
	// unstable
	CommentEOFHack bool
}

func NewScanner(r LineReader)(*Scanner){
	return &Scanner{
		r: r,
		fileset: token.NewFileSet(),
	}
}

func (s *Scanner)Scan()(nodes []ast.Node, err error){
	var line []byte
	line, err = s.r.ReadLine()
	if err != nil { return }
	if len(line) == 0 { return }
	if line[0] == '.' {
		return nil, CmdLine{line[1:]}
	}

	var parser Parser
	parser.Init(append(line, '\n'))
	parser.CommentEOFHack = s.CommentEOFHack
	for {
		nodes, err = parser.Parse()
		if err == nil {
			return nodes, nil
		}
		if !parser.IsUnexpectEOF() {
			return
		}
		var ext []byte
		ext, err = s.r.ReadLine()
		if err != nil { return nil, err }
		parser.AddLine(ext)
		parser.Reset()
	}
}
