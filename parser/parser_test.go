
package parser_test

import (
	"testing"
	. "github.com/golang-vm/gsh/parser"
)

var parser = &Parser{}

func TestVar(t *testing.T){
	type T struct {
		src string
		err bool
	}
	stmts := []T{
		{`var`, true},
		{`var 1`, true},
		{`var 1.0`, true},
		{`var "1"`, true},
		{`var a = 0`, false},
		{`var a = 0,`, true},
		{`var a = 0 1`, true},
		{`var a = 0, 1`, true},
		{`var a, = 0, 1`, true},
		{`var a, b = 0`, true},
		{`var a, b = 0, 1`, false},
		{`var a, b int = 0, 1`, false},
		{"var (a = 0; b = 1)", true},
		{"var (a = 0\nb = 1)", true},
		{"var (a = 0; b = 1;)", false},
		{"var (a = 0\nb = 1\n)", false},
		{"var (a, b = 0; c = 1;)", true},
		{"var (a, b = 0, 1; c = 1;)", false},
		{`var (a, b int = 0, 1; c string = "1";)`, false},

		{`a :=`, true},
		{`a := 0`, false},
		{`a := 0 1`, true},
		{`a := 0,`, true},
		{`a := 0, 1`, true},
		{`a, b := 0, 1`, false},
		{`a, b = 0, 1`, false},
	}
	for _, s := range stmts {
		t.Logf("Parsing: %s", s.src)
		parser.Init(([]byte)(s.src))
		_ = parser.ParseStmt()
		err := parser.Errs()
		if s.err {
			if err == nil {
				t.Fatalf("Parsing err: got no error; expected one")
			}
		}else if err != nil {
			t.Fatalf("Parsing err: got %v; expected no error", err)
		}
	}
}

func TestIf(t *testing.T){
	type T struct {
		src string
		err bool
	}
	stmts := []T{
		{`if`, true},
		{`if a`, true},
		{`if {}`, true},
		{`if ; {}`, true},
		{`if a; {}`, true},
		{`if a {}`, false},
		{`if true {}`, false},
		{`if (true) {}`, false},
		{`if !true {}`, false},
		{`if (!true) {}`, false},
		{`if a; true {}`, false},
		{`if true {} else {}`, false},
		{`if true {} else a; {}`, true},
		{`if true {} else if true {}`, false},
		{`if true {} else if ; true {} else {}`, false},
		{`if true {} else if a; true {} else {}`, false},
	}
	for _, s := range stmts {
		t.Logf("Parsing: %s", s.src)
		parser.Init(([]byte)(s.src))
		_ = parser.ParseStmt()
		err := parser.Errs()
		if s.err {
			if err == nil {
				t.Fatalf("Parsing err: got no error; expected one")
			}
		}else if err != nil {
			t.Fatalf("Parsing err: got %v; expected no error", err)
		}
	}
}

func TestFor(t *testing.T){
	type T struct {
		src string
		err bool
	}
	stmts := []T{
		{`for {}`, false},
		{`for true {}`, false},
		{`for (true) {}`, false},
		{`for ; {}`, true},
		{`for ;; {}`, false},
		{`for a;; {}`, false},
		{`for ;b; {}`, false},
		{`for ;;c {}`, false},
		{`for a;b;c {}`, false},
		{`for a;b;c++ {}`, false},
		{`for a;b;c() {}`, false},
		{`for a;b;c(x) {}`, false},
	}
	for _, s := range stmts {
		t.Logf("Parsing: %s", s.src)
		parser.Init(([]byte)(s.src))
		_ = parser.ParseStmt()
		err := parser.Errs()
		if s.err {
			if err == nil {
				t.Fatalf("Parsing err: got no error; expected one")
			}
		}else if err != nil {
			t.Fatalf("Parsing err: got %v; expected no error", err)
		}
	}
}

func TestExpr(t *testing.T){
	type T struct {
		src string
		err bool
	}
	exprs := []T{
		{`a`, false},
		{`a()`, false},
		{`a()(`, true},
		{`a()()`, false},
		{`a(0)`, false},
		{`a("", *)`, true},
		{`a("", *b)`, false},
		{`a(...)`, true},
		{`a(b...)`, false},
		{`a(...b)`, false}, // (...b) is a type, will check later
		{`a(b......)`, true},
		{`a(b...,)`, false},
		{"a(b...\n)", false},
		{"a(b...,\n)", false},
		{"a(b..., c)", true},
		{"a(b..., c...)", true},
		{"a[", true},
		{"a[]", true},
		{"a[0]", false},
		{"a[a + b]", false},
		{"a[]int", true},
		{"a[...]", true},
		{"a[...]int", true},
		{"a[0][1]", false},
		{"a[0:1]", false},
		{"a[0]1]", true},
		{"a[0:1:2]", false},
		{"a[0:1]2]", true},
		{"a[0:1:2:3]", true},
		{"a[:::]", true},
		{"a[:]", false},
		{"a[:][0]", false},
	}
	for _, s := range exprs {
		t.Logf("Parsing: %s", s.src)
		parser.Init(([]byte)(s.src))
		_ = parser.ParseExpr()
		err := parser.Errs()
		if s.err {
			if err == nil {
				t.Fatalf("Parsing err: got no error; expected one")
			}
		}else if err != nil {
			t.Fatalf("Parsing err: got %v; expected no error", err)
		}
	}
}

func TestMathExpr(t *testing.T){
	type T struct {
		src string
		err bool
	}
	exprs := []T{
		{`a +`, true},
		{`a + b`, false},
		{`a + b + c`, false},
		{`a + b * c`, false},
		{`a * b + c`, false},
		{`a + *b + c`, false},
		{`a + +b + c`, false},
		{`a + -b + c`, false},
		{`a + ^b + c`, false},
		{`a + /b + c`, true},
	}
	for _, s := range exprs {
		t.Logf("Parsing: %s", s.src)
		parser.Init(([]byte)(s.src))
		_ = parser.ParseExpr()
		err := parser.Errs()
		if s.err {
			if err == nil {
				t.Fatalf("Parsing err: got no error; expected one")
			}
		}else if err != nil {
			t.Fatalf("Parsing err: got %v; expected no error", err)
		}
	}
}

func TestType(t *testing.T){
	type T struct {
		src string
		err bool
	}
	exprs := []T{
		{`*`, true},
		{`int`, false},
		{`*int`, false},
		{`*(int)`, false},
		{`(*int)`, false},
		{`[]`, true},
		{`[int`, true},
		{`[]int`, false},
		{`*[]int`, false},
		{`[int]`, true},
		{`[0]int`, false},
		{`[a]int`, false},
		{`[[]]int`, true},
		{`[[]int]int`, false}, // Should ok to be parsed? or check later
		{`([]int + string)`, true},
		{`[...]int`, false},
		{`[...a]int`, true},
		{`[a...]int`, true},
		{`a []int`, true},
		{`map`, true},
		{`map[`, true},
		{`map[]int`, true},
		{`map[string]`, true},
		{`map[string]int`, false},
		{`chan`, true},
		{`chan int`, false},
		{`<-chan int`, false},
		{`chan<- int`, false},
		{`<-chan <-chan int`, false},
		{`chan<- chan int`, false},
		{`chan (<-chan int)`, false},
		{`chan<- chan<- int`, false},
		{`<-chan <-chan <- int`, true},
		{`func`, true},
		{`func()`, false},
		{`func(int)`, false},
		{`func(int, string)`, false},
		{`func(a int)`, false},
		{`func(a int, string)`, true},
		{`func(a int, b string)`, false},
		{`func(a, b string)`, false},
		{`func(struct{}, b string)`, true},
		{`func(a, b string, c)`, true},
		{`func()()`, false},
		{`func()int`, false},
		{`func()(int)`, false},
		{`func()(int,)`, false},
		{`func()(a int)`, false},
		{`func() a int`, true},
		{`func()(a, b int)`, false},
		{`func()(a int, string)`, true},
		{`func()(a int, b string)`, false},
		{`func()(a int, b string, c)`, true},
		{`func(int)(a int, b string, c *int)`, false},
	}
	for _, s := range exprs {
		t.Logf("Parsing: %s", s.src)
		parser.Init(([]byte)(s.src))
		_ = parser.ParseExpr()
		err := parser.Errs()
		if s.err {
			if err == nil {
				t.Fatalf("Parsing err: got no error; expected one")
			}
		}else if err != nil {
			t.Fatalf("Parsing err: got %v; expected no error", err)
		}
	}
}
