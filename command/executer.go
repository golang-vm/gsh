
package sh_command

type Executer struct{
	cmds CommandMap
}

func NewExecuter()(*Executer){
	return &Executer{
		cmds: make(CommandMap),
	}
}

func (e *Executer)Register(name string, cmd Command)(*Executer){
	e.cmds.Add(name, cmd)
	return e
}

func (e *Executer)Execute(src Source, name string, args ...string)(err error){
	cmd, ok := e.cmds.Get(name)
	if !ok {
		return &UnknownCommand{name}
	}
	return cmd.Execute(src, args...)
}

func Execute(src Source, name string, args ...string)(err error){
	return DefaultExecuter.Execute(src, name, args...)
}

func ExecuteCommand(src Source, cmd string)(err error){
	var args []string
	args, err = ParseCommand(cmd)
	if err != nil {
		return
	}
	if len(args) == 0 {
		return nil
	}
	return Execute(src, args[0], args[1:]...)
}

var DefaultExecuter = makeDefaultExecuter()

func makeDefaultExecuter()(e *Executer){
	return NewExecuter().
		Register("exit", ExitCmd{}).
		Register("sh",   ShellCmd{}).
		Register("echo", EchoCmd{})
}
