package module

import "gopkg.in/sorcix/irc.v1"

type ModuleApi struct {
	M *Module
}

type IrcData struct {
	Event string
	Msg   *irc.Message
}

type CommandData struct {
	Msg  *irc.Message
	Name string
	Args []string
}

func (mpi ModuleApi) InvokeCommand(d *CommandData, result *string) error {
	mpi.M.Commands[d.Name].Fun(d.Msg, d.Args)
	return nil
}

func (mpi ModuleApi) Dispatch(d *IrcData, result *string) error {
	for _, listener := range mpi.M.Listeners[d.Event] {
		go listener(d.Msg)
	}
	return nil
}

func (mpi ModuleApi) Cleanup(d interface{}, result *string) error {
	return nil
}
