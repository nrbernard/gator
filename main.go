package main

import (
	"fmt"
	"os"

	"github.com/nrbernard/gator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	if f, ok := c.commandMap[cmd.name]; ok {
		return f(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.name)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}

	username := cmd.args[0]
	if err := s.config.SetUser(username); err != nil {
		return err
	}

	fmt.Println("User set to:", username)
	return nil
}

func main() {
	configFile, err := config.Read()
	if err != nil {
		fmt.Printf("Failed to read config: %s\n", err)
		os.Exit(1)
	}

	appState := &state{
		config: configFile,
	}

	commands := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please specify a command")
		os.Exit(1)
	}

	command := command{
		name: args[0],
		args: args[1:],
	}

	if err := commands.run(appState, command); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config: %+v\n", configFile)
}
