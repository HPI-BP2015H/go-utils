package cli

import "sync"

var appSingleton *App
var once sync.Once

// App saves registered subcommands, flags and a bit more
type App struct {
	DefaultCommandName string
	Version            string
	commands           map[string]Command
	Fallback           func(c *Cmd, cmdName string) ExitValue
	Before             func(c *Cmd, cmdName string) string
	flags              map[string]*Flag
}

// Command is the struct for one subcommand of the app including help text and flags
type Command struct {
	Name     string
	Help     string
	Function func(*Cmd) ExitValue
	flags    map[string]*Flag
}

// AppInstance returns the singleton instance of App
func AppInstance() *App {
	once.Do(func() {
		appSingleton = &App{}
	})
	return appSingleton
}

// Commands returns all registered Commands
func (a *App) Commands() map[string]Command {
	return a.commands
}

// Flags returns all registered application wide Flags
func (a *App) Flags() map[string]*Flag {
	return a.flags
}

// RegisterCommand registers a Command
func (a *App) RegisterCommand(c Command) {
	if a.commands == nil {
		a.commands = make(map[string]Command)
	}
	a.commands[c.Name] = c
}

// RegisterFlag registers a Flag
func (a *App) RegisterFlag(f Flag) {
	if a.flags == nil {
		a.flags = make(map[string]*Flag)
	}
	a.flags[f.Long] = &f
}

// RegisterFlag registers a Flag
func (c *Command) RegisterFlag(f Flag) {
	if c.flags == nil {
		c.flags = make(map[string]*Flag)
	}
	c.flags[f.Long] = &f
}

// Flags returns all registered Flags for the Command
func (c *Command) Flags() map[string]*Flag {
	return c.flags
}

// Run executs the App with the given arguments
func (a *App) Run(arguments []string) ExitValue {
	args := NewArgs(arguments)
	cmdName := args.Peek(0)
	if cmdName == "" {
		cmdName = a.DefaultCommandName
	}
	command := a.commands[cmdName]
	cmdFunc := command.Function
	parameters := &Parameters{}
	var parameter *Parameter
	for _, flag := range a.flags {
		parameter, args = args.Extract(*flag)
		parameters.AddParameter(parameter)
	}
	for _, flag := range command.flags {
		parameter, args = args.Extract(*flag)
		parameters.AddParameter(parameter)
	}
	cmd := NewCmd(args, parameters)
	if a.Before != nil {
		res := a.Before(cmd, cmdName)
		if res != "" {
			cmdName = res
			cmdFunc = a.commands[cmdName].Function
		}
	}
	if cmdFunc != nil {
		return cmdFunc(cmd)
	}
	return a.Fallback(cmd, cmdName)
}
