package module

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/sorcix/irc.v1"
)

func execName() string {
	return strings.TrimSuffix(filepath.Base(os.Args[0]),
		filepath.Ext(filepath.Base(os.Args[0])))
}

// A CommandFun denotes a function used for irc commands.
type CommandFun func(*irc.Message, []string)

// A TriggerFun denotes a function used to check whether a Listener should fire or not.
type TriggerFun func(*irc.Message) bool

// A Listener denotes a function to fire on a registered event.
type Listener func(*irc.Message)

// A Command struct contains information about a certain command.
type Command struct {
	// A help string displayed when a user calls help on the command
	Help string

	// The function to execute when this command is called
	Fun CommandFun

	// If PM is true, the command can be used for private messages
	PM bool

	// if CM is true, the command can be used for channel messages
	CM bool
}

// The Trigger struct contains information about a certain trigger
type Trigger struct {
	// A function to run to check whether this listener can fire or not.
	Check TriggerFun

	// The listener to fire if Check returns true.
	Fun Listener
}

// A Module struct holds all necessary information about a module.
type Module struct {
	Name string
	Desc string

	// The bot RPC client
	Master *rpc.Client

	// The port that bot must connect to
	RPCPort string

	Listeners map[string][]Listener
	Commands  map[string]*Command

	// The module's RPC server
	Provider net.Listener
}

// NewModule returns a new module given a name and description.
func NewModule(name string, desc string) *Module {
	m := &Module{
		Name:      name,
		Desc:      desc,
		Listeners: make(map[string][]Listener),
		Commands:  make(map[string]*Command),
	}

	m.createRPCServer()

	return m
}

// JoinChannel is used to join a channel on IRC. The first parameter is used for sending a join
// confirmation. This should usually be the channel if called from there, or a nick if called
// privately. The second parameter is the channel to join. The third parameter is the password to
// the channel if it has one, otherwise an empty string ("") should be passed.
func (m *Module) JoinChannel(caller string, channel string, password string) (result string) {
	JoinData := struct {
		Caller   string
		Channel  string
		Password string
	}{caller, channel, ""}

	m.Master.Call("Master.JoinChannel", JoinData, &result)
	return
}

// GetName returns the module's name
func (m *Module) GetName() string {
	return m.Name
}

// GetBotVersion returns the bot's version as a string.
func (m *Module) GetBotVersion() (result string) {
	m.Master.Call("Master.GetVersion", m.Name, &result)
	return result
}

// Say will send a messagge to a channel. The first parameter is the channel, and the second
// parameter is the message to send.
func (m *Module) Say(ch string, text string) {
	result := ""
	m.Master.Call("Master.Send", ch+" :"+text, &result)
}

// Listener adds a Listener struct to the module. It will be registerd when Register() is called.
func (m *Module) Listener(event string, l func(*irc.Message)) bool {
	m.Listeners[event] = append(m.Listeners[event], l)
	return true
}

// AddCommand adds a Command struct to the module. It will be registered when Register() is called.
func (m *Module) AddCommand(name string, c *Command) {
	m.Commands[name] = c
}

// Register will register the module with bot for use. You must past your program's arguments
// to Register! The bot passes its server port via the program's arguments.
func (m *Module) Register(cargs []string) (result string, err error) {
	// Connect to the Bot's RPC server
	m.Master = rpc.NewClientWithCodec(RpcCodecClient(cargs[1]))

	for name := range m.Commands {
		m.registerCommand(name)
	}

	data := struct {
		Port       string
		ModuleName string
	}{m.RPCPort, execName()}

	err = m.Master.Call("Master.Register", data, &result)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Start serving
	m.startRPCServer()
	return result, nil
}

func (m *Module) createRPCServer() (err error) {
	port := getOpenPort()
	if port == "" {
		return errors.New("Couldn't find a port to listen on")
	}

	rpc.RegisterName(string(execName()), ModuleApi{m})

	m.Provider, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	m.RPCPort = port
	return nil
}

func (m *Module) startRPCServer() {
	conn, err := m.Provider.Accept()
	if err != nil {
		panic(err)
	}
	rpc.ServeCodec(RpcCodecServer(conn))
}

func (m *Module) registerCommand(name string) (result string) {
	data := struct {
		CommandName string
		ModuleName  string
	}{name, execName()}

	err := m.Master.Call("Master.RegisterCommand", data, &result)
	if err != nil {
		panic(err)
	}

	return
}
