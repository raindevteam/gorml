package module

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/sorcix/irc.v1"
)

type CommandFun func(*irc.Message, []string)
type TriggerFun func(*irc.Message) bool
type Listener func(*irc.Message)

type Command struct {
	Help string
	Fun  CommandFun
	PM   bool
	CM   bool
}

type Trigger struct {
	Check TriggerFun
	Fun   Listener
}

type Module struct {
	Name      string
	Desc      string
	Master    *rpc.Client
	RpcPort   string
	Listeners map[string][]Listener
	Commands  map[string]*Command
	Provider  net.Listener
}

func NewModule(name string, desc string) *Module {
	m := &Module{
		Name:      name,
		Desc:      desc,
		Listeners: make(map[string][]Listener),
		Commands:  make(map[string]*Command),
		Master:    rpc.NewClientWithCodec(RpcCodecClient()), // Connect to master
	}
	// Start Provider server
	m.initRpcServer()
	return m
}

func execName() string {
	return strings.TrimSuffix(filepath.Base(os.Args[0]),
		filepath.Ext(filepath.Base(os.Args[0])))
}

func (m *Module) initRpcServer() {
	port := getOpenPort()
	if port == "" {
		return // Handle
	}
	rpc.RegisterName(string(execName()), ModuleApi{m})
	var err error
	m.Provider, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return // Handle
	}
	m.RpcPort = port
}

func (m *Module) startRpcServer() {
	conn, _ := m.Provider.Accept()
	rpc.ServeCodec(RpcCodecServer(conn))
}

func (m *Module) GetName() string {
	return m.Name
}

func (m *Module) GetBotVersion() (result string) {
	m.Master.Call("Master.GetVersion", m.Name, &result)
	return result
}

func (m *Module) Say(ch string, text string) {
	result := ""
	m.Master.Call("Master.Send", ch+" :"+text, &result)
}

func (m *Module) RawListener(event string, l func(*irc.Message)) bool {
	m.Listeners[event] = append(m.Listeners[event], l)
	return true
}

func (m *Module) AddCommand(name string, c *Command) {
	result := ""
	data := struct {
		CommandName string
		ModuleName  string
	}{name, execName()}
	err := m.Master.Call("Master.RegisterCommand", data, &result)
	if err != nil {
		return
	}

	m.Commands[name] = c
}

func (m *Module) Register() (result string, err error) {
	data := struct {
		Port       string
		ModuleName string
	}{m.RpcPort, execName()}
	fmt.Println("registering")
	err = m.Master.Call("Master.Register", data, &result)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("done")
	m.startRpcServer()
	return result, nil
}

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
