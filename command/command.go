
package sh_command

import (
	"strings"
)

type UnknownCommand struct{
	Name string
}

func (e *UnknownCommand)Error()(string){
	return "Unknown command: " + e.Name
}

type Command interface{
	Execute(src Source, args ...string)(error)
}

type CommandMap map[string]Command

func (m CommandMap)Len()(int){
	return len(m)
}

func (m CommandMap)Add(name string, cmd Command){
	m[name] = cmd
}

func (m CommandMap)Get(name string)(cmd Command, ok bool){
	cmd, ok = m[name]
	return
}

type exitErr struct{}
func (exitErr)Error()(string){ return "exit" }

var ExitErr error = exitErr{}

type ExitCmd struct{}
func (ExitCmd)Execute(src Source, args ...string)(error){ return ExitErr }

type EchoCmd struct{}
func (EchoCmd)Execute(src Source, args ...string)(err error){
	_, err = src.Println(strings.Join(args, " "))
	return
}
