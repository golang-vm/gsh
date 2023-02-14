
package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/scanner"
)

type tokT struct{
	pos token.Pos
	tok token.Token
	lit string
}

type Parser struct{
	src []byte

	file    *token.File
	errs    scanner.ErrorList
	scanner scanner.Scanner
	unexpectEOF bool

	eof bool
	pos token.Pos
	tok token.Token
	lit string
	tokbuf []tokT

	parsing []token.Token
}

func (p *Parser)Init(src []byte)(*Parser){
	p.src = src
	p.Reset()
	return p
}

func (p *Parser)Reset(){
	p.file = token.NewFileSet().AddFile("", -1, len(p.src))
	p.errs.Reset()
	p.scanner.Init(p.file, p.src, p.errs.Add, 0)
	p.unexpectEOF, p.eof = false, false
	p.pos, p.tok, p.lit = token.NoPos, token.ILLEGAL, ""
	p.tokbuf = p.tokbuf[:0]
	p.parsing = p.parsing[:0]
}

func (p *Parser)unexpectErr(pos token.Pos, tok token.Token, lit string, msg string){
	if tok == token.EOF {
		p.unexpectEOF = true
	}
	err := "Unexpect token '" + tok.String() + "'"
	if tok.IsLiteral() {
		err += "(" + lit + ")"
	}
	if len(msg) > 0 {
		err += ", expect " + msg
	}
	p.errs.Add(p.file.Position(pos), err)
}

func (p *Parser)Errs()(err error){
	if p.errs.Len() > 0 {
		err = p.errs
	}
	return
}

func (p *Parser)IsUnexpectEOF()(bool){
	return p.unexpectEOF
}

func (p *Parser)peekMore()(pos token.Pos, tok token.Token, lit string){
	if p.eof {
		return p.pos, token.EOF, p.lit
	}
	pos, tok, lit = p.scanner.Scan()
	p.tokbuf = append(p.tokbuf, tokT{pos, tok, lit})
	if tok == token.EOF {
		return
	}
	return
}

func (p *Parser)peek()(pos token.Pos, tok token.Token, lit string){
	return p.peekN(1)
}

func (p *Parser)peekN(n int)(pos token.Pos, tok token.Token, lit string){
	if p.eof {
		return token.NoPos, token.EOF, ""
	}
	if len(p.tokbuf) < n {
		for len(p.tokbuf) < n {
			pos, tok, lit = p.peekMore()
			if tok == token.EOF {
				return token.NoPos, token.EOF, ""
			}
		}
		return
	}
	tt := p.tokbuf[n - 1]
	if tt.tok == token.EOF {
		return token.NoPos, token.EOF, ""
	}
	return tt.pos, tt.tok, tt.lit
}

func (p *Parser)peekExpect(tks ...token.Token)(tok token.Token, lit string, ok bool){
	var pos token.Pos
	pos, tok, lit = p.peek()
	for _, tk := range tks {
		if tok == tk {
			ok = true
			return
		}
	}
	p.unexpectErr(pos, tok, lit, fmt.Sprint(tks))
	return token.ILLEGAL, "", false
}

func (p *Parser)hasTokBefore(target, before token.Token)(ok bool){
	for _, tt := range p.tokbuf {
		if tt.tok == target {
			return true
		}
		if tt.tok == before {
			return false
		}
	}
	if p.eof {
		return false
	}
	var (
		pos token.Pos
		tok token.Token
		lit string
	)
	for {
		pos, tok, lit = p.scanner.Scan()
		if tok == token.EOF {
			return false
		}
		p.tokbuf = append(p.tokbuf, tokT{pos, tok, lit})
		if tok == target && (string)(p.src[pos - 1:(int)(pos - 1) + len(lit)]) == lit {
			return true
		}
		if tok == before {
			return false
		}
	}
}

func (p *Parser)next()(bool){
	if p.eof {
		return false
	}
	if len(p.tokbuf) > 0 {
		tt := p.tokbuf[0]
		p.tokbuf = p.tokbuf[:copy(p.tokbuf, p.tokbuf[1:])]
		p.pos, p.tok, p.lit = tt.pos, tt.tok, tt.lit
		if p.tok == token.EOF {
			p.unexpectErr(p.pos, p.tok, p.lit, "")
			p.eof = true
		}
		return !p.eof
	}
	p.pos, p.tok, p.lit = p.scanner.Scan()
	if p.tok == token.EOF {
		p.unexpectErr(p.pos, p.tok, p.lit, "")
		p.eof = true
	}
	return !p.eof
}

func (p *Parser)expect(tks ...token.Token)(bool){
	for _, tk := range tks {
		if p.tok == tk {
			return true
		}
	}
	p.unexpectErr(p.pos, p.tok, p.lit, fmt.Sprint(tks))
	return false
}

func (p *Parser)nextExpect(tks ...token.Token)(bool){
	if !p.next() {
		return false
	}
	return p.expect(tks...)
}

func (p *Parser)skipTo(pos token.Pos)(bool){
	for p.pos < pos {
		if !p.next() {
			return false
		}
	}
	return true
}

func (p *Parser)currentBlock()(token.Token){
	if len(p.parsing) == 0 {
		return token.ILLEGAL
	}
	return p.parsing[len(p.parsing) - 1]
}

func (p *Parser)Source()([]byte){
	return p.src
}

func (p *Parser)SetSource(src []byte){
	p.src = src
}

func (p *Parser)Append(ext []byte){
	p.src = append(p.src, ext...)
}

func (p *Parser)AddLine(line []byte){
	if len(line) == 0 || line[len(line) - 1] != '\n' {
		line = append(line, '\n')
	}
	p.src = append(p.src, line...)
}

func (p *Parser)Parse()(nodes []ast.Node, err error){
	if p.errs.Len() != 0 {
		return nil, p.errs
	}
	main: for {
		pos, tok, _ := p.peek()
		if tok == token.EOF {
			break
		}
		switch tok {
		case token.PACKAGE:
			p.errs.Add(p.file.Position(pos), "Not support package now")
			break main
		case token.IMPORT:
			p.next()
			imp := p.parseImport()
			if imp == nil {
				break main
			}
			nodes = append(nodes, imp)
		case token.SEMICOLON:
			p.next()
		default:
			n := p.ParseStmt()
			if n == nil {
				break main
			}
			nodes = append(nodes, n)
		}
	}
	if p.errs.Len() > 0 {
		err = p.errs
	}
	return
}

func (p *Parser)parseImport()(*ast.GenDecl){
	return p.parseDecl(func()(ast.Spec){
		if !p.nextExpect(token.IDENT, token.PERIOD, token.STRING) { return nil }
		var name *ast.Ident = nil
		if p.tok != token.STRING {
			name = &ast.Ident{
				NamePos: p.pos,
				Name: p.lit,
			}
			if !p.nextExpect(token.STRING) { return nil }
		}
		path := &ast.BasicLit{
			ValuePos: p.pos,
			Kind: token.STRING,
			Value: p.lit,
		}
		if !p.nextExpect(token.SEMICOLON) { return nil }
		return &ast.ImportSpec{
			Name: name,
			Path: path,
			EndPos: p.pos,
		}
	})
}

func (p *Parser)ParseStmt()(stmt ast.Stmt){
	if !p.next() {
		return nil
	}
	if stmt = p.parseStmt(); stmt == nil {
		return nil
	}
	if !p.expect(token.SEMICOLON) {
		return nil
	}
	return
}

func (p *Parser)parseStmt()(stmt ast.Stmt){
	pos0, tok0, lit0 := p.pos, p.tok, p.lit
	switch tok0 {
	case token.SEMICOLON:
		return &ast.EmptyStmt{
			Semicolon: pos0,
		}
	case token.TYPE:
		decl := p.parseDecl(func()(ast.Spec){
			if !p.nextExpect(token.IDENT) {
				return nil
			}
			name := &ast.Ident{
				NamePos: p.pos,
				Name: p.lit,
			}
			var assign token.Pos
			if _, tok, _ := p.peek(); tok == token.ASSIGN {
				p.next()
				assign = p.pos
			}
			typ := p.ParseType()
			if typ == nil {
				return nil
			}
			if !p.nextExpect(token.SEMICOLON) {
				return nil
			}
			return &ast.TypeSpec{
				Name: name,
				Assign: assign,
				Type: typ,
			}
		})
		if decl == nil || !p.nextExpect(token.SEMICOLON) {
			return nil
		}
		return &ast.DeclStmt{Decl: decl}
	case token.CONST, token.VAR:
		decl := p.parseDecl(func()(ast.Spec){
			var (
				lhs []*ast.Ident
				typ ast.Expr
				rhs []ast.Expr
			)
			for {
				if !p.nextExpect(token.IDENT) {
					return nil
				}
			 	lhs = append(lhs, &ast.Ident{
					NamePos: p.pos,
					Name: p.lit,
				})
				if _, tok, _ := p.peek(); tok != token.COMMA {
					if tok != token.ASSIGN {
						if typ = p.ParseType(); typ == nil {
							return nil
						}
					}
					break
				}
				p.next()
			}
			if !p.nextExpect(token.ASSIGN) {
				return nil
			}
			rhs = make([]ast.Expr, 0, len(lhs))
			for {
				if !p.next() {
					return nil
				}
				expr := p.parseExpr(false)
				if expr == nil {
					return nil
				}
				rhs = append(rhs, expr)
				if len(rhs) == len(lhs) {
					break
				}
				if !p.nextExpect(token.COMMA) {
					return nil
				}
			}
			return &ast.ValueSpec{
				Names: lhs,
				Type: typ,
				Values: rhs,
			}
		})
		if decl == nil || !p.nextExpect(token.SEMICOLON) {
			return nil
		}
		return &ast.DeclStmt{Decl: decl}
	case token.IF:
		if !p.next() {
			return nil
		}
		if stmt = p.parseIf(); stmt == nil || !p.nextExpect(token.SEMICOLON) {
			return nil
		}
		return
	case token.SELECT:
		if !p.next() {
			return nil
		}
		if stmt = p.parseSelect(); stmt == nil || !p.nextExpect(token.SEMICOLON) {
			return nil
		}
		return
	case token.SWITCH:
		if !p.next() {
			return nil
		}
		if stmt = p.parseSwitch(); stmt == nil || !p.nextExpect(token.SEMICOLON) {
			return nil
		}
		return
	case token.FOR:
		if !p.next() {
			return nil
		}
		if stmt = p.parseFor(); stmt == nil || !p.nextExpect(token.SEMICOLON) {
			return nil
		}
		return
	case token.GO:
		expr := p.ParseExpr()
		call, ok := expr.(*ast.CallExpr)
		if !ok || call == nil {
			p.unexpectErr(pos0, tok0, lit0, "a func call")
			return nil
		}
		return &ast.GoStmt{
			Go: pos0,
			Call: call,
		}
	case token.DEFER:
		expr := p.ParseExpr()
		call, ok := expr.(*ast.CallExpr)
		if !ok || call == nil {
			p.unexpectErr(pos0, tok0, lit0, "a func call")
			return nil
		}
		return &ast.DeferStmt{
			Defer: pos0,
			Call: call,
		}
	case token.FUNC:
		recv, name, ok := p.parseFuncDecl()
		if !ok {
			return nil
		}
		fc, bd := p.parseFuncExpr()
		if fc == nil {
			return nil
		}
		if !p.next() {
			return nil
		}
		return &ast.DeclStmt{
			Decl: &ast.FuncDecl{
				Recv: recv,
				Name: name,
				Type: fc,
				Body: bd,
			},
		}
	case token.LBRACE:
		stmt = p.parseBlock()
		if !p.nextExpect(token.SEMICOLON) {
			return nil
		}
		return
	case token.IDENT:
		if p.hasTokBefore(token.ASSIGN, token.SEMICOLON) || p.hasTokBefore(token.DEFINE, token.SEMICOLON) {
			if stmt = p.parseAssign(); stmt == nil || !p.next() {
				return nil
			}
			return
		}
		_, tok, _ := p.peek()
		switch tok {
		case token.COLON:
			p.next()
			if stmt = p.ParseStmt(); stmt == nil {
				return nil
			}
			return &ast.LabeledStmt{
				Label: &ast.Ident{
					NamePos: pos0,
					Name: lit0,
				},
				Colon: pos0,
				Stmt: stmt,
			}
		case token.INC, token.DEC:
			p.next()
			if !p.next() {
				return nil
			}
			return &ast.IncDecStmt{
				X: &ast.Ident{
					NamePos: pos0,
					Name: lit0,
				},
				TokPos: p.pos,
				Tok: p.tok,
			}
		}
		fallthrough
	default:
		expr := p.parseExpr(false)
		if expr == nil {
			return nil
		}
		if !p.next() {
			return nil
		}
		return &ast.ExprStmt{X: expr}
	}
}

func (p *Parser)parseDecl(cb func()(ast.Spec))(*ast.GenDecl){
	var (
		pos0, tok0 = p.pos, p.tok
		lpos, rpos token.Pos
		specs []ast.Spec
		paren bool
	)
	if pos, tok, _ := p.peek(); tok == token.LPAREN {
		lpos, paren = pos, true
	}
	if paren {
		p.next()
		for {
			if _, tok, _ := p.peek(); tok == token.RPAREN {
				p.next()
				break
			}
			sp := cb()
			if sp == nil {
				return nil
			}
			specs = append(specs, sp)
			if !p.nextExpect(token.SEMICOLON) {
				return nil
			}
		}
	}else{
		specs = []ast.Spec{cb()}
		if specs[0] == nil { return nil }
	}
	return &ast.GenDecl{
		TokPos: pos0,
		Tok: tok0,
		Lparen: lpos,
		Specs: specs,
		Rparen: rpos,
	}
}

func (p *Parser)parseFuncDecl()(recv *ast.FieldList, name *ast.Ident, ok bool){
	if !p.nextExpect(token.LPAREN, token.IDENT) {
		return nil, nil, false
	}
	if p.tok == token.LPAREN {
		recv = &ast.FieldList{ Opening: p.pos }
		if !p.nextExpect(token.IDENT, token.MUL) {
			return nil, nil, false
		}
		var ident *ast.Ident
		if p.tok == token.IDENT {
			ident = &ast.Ident{
				NamePos: p.pos,
				Name: p.lit,
			}
			if !p.nextExpect(token.IDENT, token.MUL, token.RPAREN) {
				return nil, nil, false
			}
		}
		if p.tok == token.RPAREN {
			recv.List = []*ast.Field{ &ast.Field{
				Type: ident,
			} }
		}else{
			var typ ast.Expr
			if p.tok == token.MUL {
				ps := p.pos
				if !p.nextExpect(token.IDENT) {
					return nil, nil, false
				}
				typ = &ast.StarExpr{
					Star: ps,
					X: &ast.Ident{
						NamePos: p.pos,
						Name: p.lit,
					},
				}
			}else{
				typ = &ast.Ident{
					NamePos: p.pos,
					Name: p.lit,
				}
			}
			recv.List = []*ast.Field{ &ast.Field{
				Names: []*ast.Ident{ident},
				Type: typ,
			} }
			if !p.nextExpect(token.RPAREN) {
				return nil, nil, false
			}
		}
		if !p.nextExpect(token.IDENT) {
			return nil, nil, false
		}
	}
	if p.tok == token.IDENT {
		name = &ast.Ident{
			NamePos: p.pos,
			Name: p.lit,
		}
	}
	ok = true
	return
}

func (p *Parser)parseAssign()(*ast.AssignStmt){
	var lhs, rhs []ast.Expr
	for {
		lhs = append(lhs, &ast.Ident{
			NamePos: p.pos,
			Name: p.lit,
		})
		if !p.nextExpect(token.COMMA, token.ASSIGN, token.DEFINE) {
			return nil 
		}
		if p.tok != token.COMMA {
			break
		}
		if !p.nextExpect(token.IDENT) {
			return nil 
		}
	}
	tkp, tok := p.pos, p.tok
	rhs = make([]ast.Expr, 0, len(lhs))
	for {
		if !p.next() {
			return nil
		}
		expr := p.parseExpr(false)
		if expr == nil {
			return nil
		}
		rhs = append(rhs, expr)
		if len(rhs) >= len(lhs) {
			break
		}
		if !p.nextExpect(token.COMMA) {
			return nil 
		}
	}
	return &ast.AssignStmt{
		Lhs: lhs,
		TokPos: tkp,
		Tok: tok,
		Rhs: rhs,
	}
}

func (p *Parser)ParseExpr()(expr ast.Expr){
	if !p.next() {
		return nil
	}
	return p.parseExpr(false)
}

func (p *Parser)parseExpr(parsing2 bool)(expr ast.Expr){
	for {
		pos, tok := p.pos, p.tok
		switch tok {
		case token.FUNC:
			fc, bd := p.parseFuncExpr()
			if fc == nil {
				return nil
			}
			if bd == nil {
				expr = fc
			}else{
				expr = &ast.FuncLit{
					Type: fc,
					Body: bd,
				}
			}
		case token.LBRACK, token.STRUCT, token.INTERFACE, token.MAP, token.CHAN, token.ARROW:
			return p.parseType()
		case token.IDENT:
			if expr != nil {
				p.unexpectErr(p.pos, p.tok, p.lit, "")
				return
			}
			expr = &ast.Ident{
				NamePos: p.pos,
				Name: p.lit,
			}
		case token.INT, token.FLOAT, token.IMAG, token.CHAR, token.STRING:
			if expr != nil {
				p.unexpectErr(p.pos, p.tok, p.lit, "")
				return
			}
			expr = &ast.BasicLit{
				ValuePos: p.pos,
				Kind: p.tok,
				Value: p.lit,
			}
		case token.NOT:
			exp := p.ParseExpr()
			if exp == nil {
				return
			}
			expr = &ast.UnaryExpr{
				OpPos: pos,
				Op: tok,
				X: exp,
			}
			break
		case token.MUL, token.AND, token.ADD, token.SUB, token.XOR:
			if expr == nil {
				exp := p.ParseExpr()
				if exp == nil {
					return
				}
				if tok == token.MUL {
					expr = &ast.StarExpr{
						Star: pos,
						X: exp,
					}
				}else{
					expr = &ast.UnaryExpr{
						OpPos: pos,
						Op: tok,
						X: exp,
					}
				}
				break
			}
			fallthrough
		case token.QUO, token.REM, token.OR,
				token.SHL, token.SHR, token.AND_NOT,
				token.LAND, token.LOR, token.EQL, token.LSS, token.GTR,
				token.NEQ, token.LEQ, token.GEQ:
			if expr == nil {
				p.unexpectErr(p.pos, p.tok, p.lit, "expr")
				return nil
			}
			if parsing2 {
				return
			}
			if expr = p.parseExpr2([]ast.Expr{expr}, []tokT{{p.pos, p.tok, p.lit}}, 0); expr == nil {
				return nil
			}
			return
		case token.LPAREN:
			if expr == nil {
				if expr = p.ParseExpr(); expr == nil || !p.nextExpect(token.RPAREN) {
					return nil
				}
				expr = &ast.ParenExpr{
					Lparen: pos,
					X: expr,
					Rparen: p.pos,
				}
			}else if isFuncExpr(expr) {
				var args []ast.Expr
				if !p.next() {
					return
				}
				for p.tok != token.RPAREN {
					arg := p.parseExpr(false)
					if arg == nil {
						return nil
					}
					if !p.nextExpect(token.COMMA, token.RPAREN) {
						return nil
					}
					args = append(args, arg)
					if p.tok == token.RPAREN {
						break
					}
					if !p.next() {
						return
					}
				}
				expr = &ast.CallExpr{
					Fun: expr,
					Lparen: pos,
					Args: args,
					Rparen: p.pos,
				}
			}else{
				p.unexpectErr(p.pos, p.tok, p.lit, "")
				return nil
			}
		default:
			if expr == nil {
				p.unexpectErr(p.pos, p.tok, p.lit, "expr")
				return nil
			}
			return
		}
		switch _, tok, _ = p.peek(); tok {
		case token.EOF, token.RPAREN, token.RBRACK, token.RBRACE, token.LBRACE, token.COMMA, token.SEMICOLON:
			if expr == nil {
				p.unexpectErr(p.pos, p.tok, p.lit, "expr")
				return nil
			}
			return
		}
		p.next()
	}
}

func zipExprs(l []ast.Expr, ops []tokT, i int)(expr ast.Expr){
	expr = l[i]
	for j, o := range ops[i:] {
		expr = &ast.BinaryExpr{
			X: expr,
			OpPos: o.pos,
			Op: o.tok,
			Y: l[i + j + 1],
		}
	}
	return
}

func (p *Parser)parseExpr2(l []ast.Expr, ops []tokT, i int)(expr ast.Expr){
	for {
		if !p.next() {
			return nil
		}
		if expr = p.parseExpr(true); expr == nil {
			return nil
		}
		switch p.tok {
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM, token.AND, token.OR,
				token.XOR, token.SHL, token.SHR, token.AND_NOT,
				token.LAND, token.LOR, token.EQL, token.LSS, token.GTR,
				token.NEQ, token.LEQ, token.GEQ:
			if x, y := ops[len(ops) - 1].tok.Precedence(), p.tok.Precedence(); x > y {
				// if it's (... a * b + c) format, calc (a * b)
				l = append(l, expr)
				expr = zipExprs(l, ops, i)
				l, ops = l[:i], ops[:i]
				if i != 0 {
					// if it's (x + (a * b) + c) format, calc (x + (a * b))
					if x = ops[i - 1].tok.Precedence(); x >= y {
						expr = zipExprs(append(l, expr), ops, 0)
						i, l, ops = 0, l[:0], ops[:0]
					}
				}
			}else if x < y {
				// if it's (x + a + b + c * d) format, calc (x + a + b)
				o0 := ops[len(ops) - 1]
				l[i], ops[i] = zipExprs(l, ops[:len(ops) - 1], i), o0
				i += 1
				l, ops = l[:i], ops[:i]
			}
			l, ops = append(l, expr), append(ops, tokT{p.pos, p.tok, p.lit})
			continue
		}
		expr = zipExprs(append(l, expr), ops, i)
		if i > 0 {
			x := zipExprs(l, ops[:i - 1], 0)
			expr = &ast.BinaryExpr{
				X: x,
				OpPos: ops[i - 1].pos,
				Op: ops[i - 1].tok,
				Y: expr,
			}
		}
		return
	}
}

func (p *Parser)parseFuncExpr()(fc *ast.FuncType, bd *ast.BlockStmt){
	fc = p.parseFuncType()
	if fc == nil {
		return nil, nil
	}
	if _, t, _ := p.peek(); t == token.LBRACE {
		p.next()
		p.parsing = append(p.parsing, token.FUNC)
		bd = p.parseBlock()
		p.parsing = p.parsing[:len(p.parsing) - 1]
		if bd == nil {
			return nil, nil
		}
	}
	return
}

func (p *Parser)ParseType()(ast.Expr){
	if !p.next() {
		return nil
	}
	return p.parseType()
}

func (p *Parser)parseType()(ast.Expr){
	pos := p.pos
	switch p.tok {
	case token.LBRACK:
		return p.parseArrayType()
	case token.STRUCT:
		return p.parseStructType()
	case token.INTERFACE:
		return p.parseInterface()
	case token.FUNC:
		return p.parseFuncType()
	case token.MAP:
		return p.parseMapType()
	case token.CHAN, token.ARROW:
		return p.parseChan()
	case token.MUL:
		t := p.ParseType()
		if t == nil {
			return nil
		}
		return &ast.StarExpr{
			Star: pos,
			X: t,
		}
	case token.LPAREN:
		t := p.parseExpr(false)
		if t == nil || !p.expect(token.RPAREN) {
			return nil
		}
		return &ast.ParenExpr{
			Lparen: pos,
			X: t,
			Rparen: p.pos,
		}
	case token.IDENT:
		return &ast.Ident{
			NamePos: pos,
			Name: p.lit,
		}
	default:
		p.unexpectErr(p.pos, p.tok, p.lit, "a type")
		return nil
	}
}

func (p *Parser)parseArrayType()(*ast.ArrayType){
	pos := p.pos
	var ln ast.Expr
	_, tok, _ := p.peek()
	switch tok {
	case token.ELLIPSIS:
		p.next()
		ln = &ast.Ellipsis{
			Ellipsis: p.pos,
		}
	case token.RBRACK:
	case token.EOF:
		return nil
	default:
		if ln = p.ParseExpr(); ln == nil {
			return nil
		}
	}
	if !p.next() {
		return nil
	}
	val := p.ParseType()
	if val == nil {
		return nil
	}
	return &ast.ArrayType{
		Lbrack: pos,
		Len: ln,
		Elt: val,
	}
}

func (p *Parser)parseStructType()(*ast.StructType){
	pos := p.pos
	if !p.nextExpect(token.LBRACE) {
		return nil
	}
	var fields []*ast.Field
	for p.nextExpect(token.IDENT, token.RBRACE) {
		if p.tok == token.RBRACE {
			return &ast.StructType{
				Struct: pos,
				Fields: &ast.FieldList{Opening: pos, List: fields, Closing: p.pos},
			}
		}
	}
	return nil
}

func (p *Parser)parseInterface()(*ast.InterfaceType){
	pos := p.pos
	if !p.nextExpect(token.LBRACE) {
		return nil
	}
	var fields []*ast.Field
	for p.nextExpect(token.IDENT, token.RBRACE) {
		if p.tok == token.RBRACE {
			return &ast.InterfaceType{
				Interface: pos,
				Methods: &ast.FieldList{Opening: pos, List: fields, Closing: p.pos},
			}
		}
	}
	return nil
}

func (p *Parser)parseFuncFields()(*ast.FieldList){
	start := p.pos
	var (
		idents []*ast.Ident
		list []*ast.Field
		// flag of type parms
		flag bool = false
		// flag of parms
		flg2 bool = false
	)
	L: for {
		if !p.nextExpect(token.RPAREN, token.IDENT,
			token.MUL, token.STRUCT, token.INTERFACE, token.FUNC, token.MAP, token.CHAN, token.ARROW) {
			return nil
		}
		if p.tok == token.RPAREN {
			break
		}
		if p.tok == token.IDENT {
			idents = append(idents, &ast.Ident{
				NamePos: p.pos,
				Name: p.lit,
			})
		}else{
			if flg2 {
				p.errs.Add(p.file.Position(p.pos), "syntax error: mixed named and unnamed parameters")
				return nil
			}
			flag = true
			for _, d := range idents {
				list = append(list, &ast.Field{
					Type: d,
				})
			}
			idents = nil
			typ := p.parseType()
			if typ == nil {
				return nil
			}
			list = append(list, &ast.Field{
				Type: typ,
			})
		}
		if !p.next() {
			return nil
		}
		switch p.tok {
		case token.IDENT, token.MUL, token.STRUCT, token.INTERFACE, token.FUNC, token.MAP, token.CHAN, token.ARROW:
			if flag {
				p.errs.Add(p.file.Position(p.pos), "syntax error: mixed named and unnamed parameters")
				return nil
			}
			flg2 = true
			typ := p.parseType()
			if typ == nil {
				return nil
			}
			list = append(list, &ast.Field{
				Names: idents,
				Type: typ,
			})
			idents = nil
			if !p.nextExpect(token.COMMA, token.RPAREN) {
				return nil
			}
			if p.tok == token.RPAREN {
				break L
			}
		case token.COMMA:
		case token.RPAREN:
			break L
		default:
			p.unexpectErr(p.pos, p.tok, p.lit, "a type expr, comma, or ')'")
			return nil
		}
	}
	if len(idents) > 0 {
		if flg2 {
			p.errs.Add(p.file.Position(p.pos), "syntax error: mixed named and unnamed parameters")
			return nil
		}
		for _, d := range idents {
			list = append(list, &ast.Field{
				Type: d,
			})
		}
	}
	return &ast.FieldList{
		Opening: start,
		List: list,
		Closing: p.pos,
	}
}

func (p *Parser)parseFuncType()(*ast.FuncType){
	pos := p.pos
	t, _, ok := p.peekExpect(token.LPAREN, token.LBRACE)
	if !ok {
		return nil
	}
	if t == token.LBRACE {
		return &ast.FuncType{
			Func: pos,
		}
	}
	p.next()
	if t, _, ok = p.peekExpect(token.RPAREN, token.IDENT,
		token.MUL, token.STRUCT, token.INTERFACE, token.FUNC, token.MAP, token.CHAN, token.ARROW); !ok {
		return nil
	}
	var parm, res *ast.FieldList
	if p.tok == token.RPAREN {
		if t != token.LPAREN && t != token.IDENT {
			return &ast.FuncType{
				Func: pos,
			}
		}
		p.next()
		if p.tok == token.LPAREN {
			if res = p.parseFuncFields(); res == nil {
				return nil
			}
		}else{
			typ := p.parseType()
			if typ == nil {
				return nil
			}
			res = &ast.FieldList{
				List: []*ast.Field{ &ast.Field{
					Type: typ,
				} },
			}
		}
	}else{
		if parm = p.parseFuncFields(); parm == nil {
			return nil
		}
		_, t, _ = p.peek()
		switch t {
		case token.IDENT, token.MUL, token.STRUCT, token.INTERFACE, token.FUNC, token.MAP, token.CHAN, token.ARROW:
			p.next()
			typ := p.parseType()
			if typ == nil {
				return nil
			}
			res = &ast.FieldList{
				List: []*ast.Field{ &ast.Field{
					Type: typ,
				} },
			}
		case token.LPAREN:
			p.next()
			if res = p.parseFuncFields(); res == nil {
				return nil
			}
		}
	}
	return &ast.FuncType{
		Func: pos,
		Params: parm,
		Results: res,
	}
}

func (p *Parser)parseMapType()(*ast.MapType){
	pos := p.pos
	if !p.nextExpect(token.LBRACK) {
		return nil
	}
	key := p.ParseType()
	if key == nil || !p.nextExpect(token.RBRACK) {
		return nil
	}
	val := p.ParseType()
	if val == nil {
		return nil
	}
	return &ast.MapType{
		Map: pos,
		Key: key,
		Value: val,
	}
}

func (p *Parser)parseChan()(*ast.ChanType){
	pos := p.pos
	if p.tok == token.ARROW {
		if !p.nextExpect(token.CHAN) { return nil }
		val := p.ParseType()
		if val == nil { return nil }
		return &ast.ChanType{
			Begin: pos,
			Arrow: pos,
			Dir: ast.RECV,
			Value: val,
		}
	}
	var (
		arrow token.Pos = token.NoPos
		dir ast.ChanDir = ast.SEND
	)
	_, tok, _ := p.peek()
	if tok == token.ARROW {
		p.next()
		arrow = p.pos
	}else{
		dir |= ast.RECV
	}
	val := p.ParseType()
	if val == nil { return nil }
	return &ast.ChanType{
		Begin: pos,
		Arrow: arrow,
		Dir: dir,
		Value: val,
	}
}

func (p *Parser)parseBlock()(blk *ast.BlockStmt){
	pos := p.pos
	var list []ast.Stmt
	for p.next() {
		if p.tok == token.RBRACE {
			return &ast.BlockStmt{
				Lbrace: pos,
				List: list,
				Rbrace: p.pos,
			}
		}
		stmt := p.parseStmt()
		if stmt == nil {
			return nil
		}
		list = append(list, stmt)
	}
	p.unexpectErr(p.pos, p.tok, p.lit, token.RBRACE.String())
	return nil
}

func (p *Parser)parseIf()(*ast.IfStmt){
	pos := p.pos
	var (
		st ast.Stmt
		cond ast.Expr
	)
	if p.hasTokBefore(token.SEMICOLON, token.LBRACE) {
		if st = p.parseStmt(); st == nil {
			return nil
		}
	}
	if p.tok == token.SEMICOLON && !p.next() {
		return nil
	}
	if cond = p.parseExpr(false); cond == nil || !p.nextExpect(token.LBRACE) {
		return nil
	}
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	var elseb ast.Stmt
	if _, tok, _ := p.peek(); tok == token.ELSE {
		p.next()
		if !p.nextExpect(token.LBRACE, token.IF) {
			return nil
		}
		if p.tok == token.IF {
			if !p.next() {
				return nil
			}
			elseb = p.parseIf()
		}else{
			elseb = p.parseBlock()
		}
		if elseb == nil {
			return nil
		}
	}
	return &ast.IfStmt{
		If: pos,
		Init: st,
		Cond: cond,
		Body: body,
		Else: elseb,
	}
}

func (p *Parser)parseSelect()(*ast.SelectStmt){
	return nil
}

func (p *Parser)parseSwitch()(*ast.SwitchStmt){
	return nil
}

func (p *Parser)parseFor()(ast.Stmt){
	pos := p.pos
	var (
		s1, s3 ast.Stmt
		cond ast.Expr
	)

	if p.tok != token.LBRACE {
		if p.hasTokBefore(token.RANGE, token.LBRACE) {
			return p.parseRange(pos)
		}
		for3 := p.tok == token.SEMICOLON || p.hasTokBefore(token.SEMICOLON, token.LBRACE)
		if for3 {
			if s1 = p.parseStmt(); s1 == nil || !p.next() {
				return nil
			}
		}
		if !for3 || p.tok != token.SEMICOLON {
			if cond = p.parseExpr(false); cond == nil || !p.next() {
				return nil
			}
		}
		if for3 {
			if !p.expect(token.SEMICOLON) || !p.next() {
				return nil
			}
			if p.tok != token.LBRACE {
				if s3 = p.parseStmt(); s3 == nil {
					return nil
				}
			}
		}
		if !p.expect(token.LBRACE) {
			return nil
		}
	}
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	return &ast.ForStmt{
		For: pos,
		Init: s1,
		Cond: cond,
		Post: s3,
		Body: body,
	}
}

func (p *Parser)parseRange(pos token.Pos)(*ast.RangeStmt){
	return nil
}

func isFuncExpr(expr ast.Expr)(bool){
	if expr != nil {
		switch x := expr.(type) {
		case *ast.FuncLit:
			return true
		case *ast.Ident:
			return true
		case *ast.ParenExpr:
			return isFuncExpr(x.X)
		}
	}
	return false
}
