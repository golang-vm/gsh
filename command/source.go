
package sh_command

import (
	"io"
	"fmt"
)

type Source interface{
	io.Reader
	io.Writer

	Print(args ...interface{})(int, error)
	Println(args ...interface{})(int, error)
	Printf(format string, args ...interface{})(int, error)
}

type source struct{
	io.Reader
	io.Writer
}

var _ Source = (*source)(nil)

func NewSource(r io.Reader, w io.Writer)(Source){
	return &source{
		Reader: r,
		Writer: w,
	}
}

func (s *source)Print(args ...interface{})(n int, err error){
	return fmt.Fprint(s, args...)
}

func (s *source)Println(args ...interface{})(n int, err error){
	return fmt.Fprintln(s, args...)
}

func (s *source)Printf(format string, args ...interface{})(n int, err error){
	return fmt.Fprintf(s, format, args...)
}
