package module

import (
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RyanPrintup/nimbus"
	"github.com/wolfchase/rainbot"
)

type CommandFun func(*nimbus.Message, []string)
type TriggerFun func(*nimbus.Message) bool
type Listener func(*nimbus.Message)

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
	Listeners map[rainbot.Event][]Listener
	Commands  map[rainbot.CommandName]*Command
}

func MakeModule(name string, desc string) *Module {
	m := &Module{
		Name:      name,
		Desc:      desc,
		Listeners: make(map[rainbot.Event][]Listener),
		Commands:  make(map[rainbot.CommandName]*Command),
		Master:    rpc.NewClientWithCodec(rainbot.RpcCodecClient()), // Connect to master
	}
	// Start Provider server
	m.startRpcServer()
	return m
}

func execName() rainbot.ModuleName {
	return rainbot.ModuleName(strings.TrimSuffix(filepath.Base(os.Args[0]),
		filepath.Ext(filepath.Base(os.Args[0]))))
}

func (m *Module) startRpcServer() {
	port := getOpenPort()
	if port == "" {
		return // Handle
	}
	rpc.RegisterName(string(execName()), ModuleApi{m})
	provider, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return // Handle
	}
	m.RpcPort = port
	go func() {
		for {
			conn, _ := provider.Accept()
			rpc.ServeCodec(rainbot.RpcCodecServer(conn))
		}
	}()
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

func (m *Module) RawListener(event rainbot.Event, l func(*nimbus.Message)) bool {
	m.Listeners[event] = append(m.Listeners[event], l)
	return true
}

func (m *Module) CommandHook(name rainbot.CommandName, c *rainbot.Command) {
	result := ""
	err := m.Master.Call("Master.RegisterCommand",
		rainbot.CommandRequest{name, execName()}, &result)
	if err != nil {
		return
	}

	m.Commands[name] = c
}

func (m *Module) Register() (result string, err error) {
	m.Master.Call("Master.Reg", rainbot.Ticket{m.RpcPort, execName()}, &result)
	m.Master.Call("Master.Loop", "", &result)
	return result, nil
}

type ModuleApi struct {
	M *Module
}

func (mpi ModuleApi) InvokeCommand(d *rainbot.CommandData, result *string) error {
	mpi.M.Commands[d.Name].Fun(d.Msg, d.Args)
	return nil
}

func (mpi ModuleApi) Dispatch(d *rainbot.IrcData, result *string) error {
	for _, listener := range mpi.M.Listeners[d.Event] {
		go listener(d.Msg)
	}
	return nil
}

func getOpenPort() string {
	for i := 65535; i >= 49152; i-- {
		conn, err := net.Listen("tcp", ":"+strconv.Itoa(i))
		if err == nil {
			conn.Close()
			return strconv.Itoa(i)
		}
	}
	return ""
}
