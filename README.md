# Go Rain Module Library

### Installation

To install this library:

    $ go get github.com/raindevteam/gorml

### Usage

First, make sure you have installed gorml via go get. Once you have done that, you can begin writing
your own module. Here's a simple example:

```go
package main

import (
    "os"
    "strings"

    "github.com/raindevteam/gorml"
    "gopkg.in/sorcix/irc.v1"
)

func main() {
    m := NewModule("EchoM", "A one command echo module")

    m.AddCommand("echo", &module.Command{
		Help: "Repeats what you give it",
		Fun: func(msg *irc.Message, args []string) {
            m.Say(msg.Params[0], strings.Join(args, " "))
        },
	})

    m.register(os.Args)
}
```

Adding commands this way disallows you from providing help text. We hope to change this in the 
future. It also isn't much less code than the next example. If you want a smaller footprint for
simpler commands and don't mind using Python, we recommend [pyrml](https://github.com/raindevteam/pyrml)

And here's a slightly more complex one that uses a struct

```go
package main

import (
	"os"
    "strings"

	"github.com/raindevteam/gorml"
	"gopkg.in/sorcix/irc.v1"
)

type Echo struct { *module.Module }

func (m *Echo) Echo(msg *irc.Message, args []string) {
	m.Say(msg.Args[0], strings.Join(args, " "))
}

func main() {
	m := &Echo{module.NewModule("Echo", "An echo module")}

	m.AddCommand("echo", &module.Command{
		Help: ",
		Fun:  m.Info,
	})

	m.Register(os.Args)
}
```

After writing your module, you may install it with go install. You can then reference the module in
your Rain powered Bot as so:

```yaml
modules:
  go:
    echo: path/to/folder/holding/srcfolder # i.e. github.com/youruser
```