
package sh_command

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"github.com/creack/pty"
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

type ShellCmd struct{}
func (ShellCmd)Execute(src Source, args ...string)(err error){
	var carg []string
	if len(args) == 0 {
		carg = []string{"-i"}
	} else {
		carg = []string{"-c", strings.Join(args, " ")}
	}
	cmd := exec.Command("bash", carg...)
	var f *os.File
	f, err = pty.Start(cmd)
	go io.Copy(f, src)
	go io.Copy(src, f)
	err = cmd.Wait()
	return
}

type EchoCmd struct{}
func (EchoCmd)Execute(src Source, args ...string)(err error){
	_, err = src.Println(strings.Join(args, " "))
	return
}
