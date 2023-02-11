
package main

import (
	"go/constant"
)

type Variable struct{
	cst bool
	name string
	typ string
	value interface{}
}

type Context struct{
	Vars []Variable
	Consts []constant.Value
}

type File struct{
	Package string
	Imports []string

	Context
}
