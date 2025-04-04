package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/nrbernard/gator/internal/config"
	"github.com/nrbernard/gator/internal/database"
)

type state struct {
	db     *database.Queries
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

	if _, err := s.db.GetUser(context.Background(), username); err != nil {
		return fmt.Errorf("user not found: %s", username)
	}

	if err := s.config.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %s", err)
	}

	fmt.Println("User set to:", username)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteUsers(context.Background()); err != nil {
		return fmt.Errorf("failed to delete users: %s", err)
	}

	fmt.Println("Users deleted")

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}

	username := cmd.args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:   uuid.New(),
		Name: username,
	})
	if err != nil {
		fmt.Printf("failed to create user: %s\n", err)
		os.Exit(1)
	}

	s.config.SetUser(username)

	fmt.Printf("User created: %+v\n", user)
	return nil
}

func main() {
	configFile, err := config.Read()
	if err != nil {
		fmt.Printf("Failed to read config: %s\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", configFile.DBUrl)
	if err != nil {
		fmt.Printf("Failed to connect to database: %s\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	appState := &state{
		db:     dbQueries,
		config: configFile,
	}

	commands := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)

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
