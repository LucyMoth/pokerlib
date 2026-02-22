package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const version = "1.0.0"

func main() {
	cli := NewCLI()
	cli.PrintWelcome()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		cli.PrintPrompt()
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if !cli.Execute(line) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

type CLI struct {
	session  *Session
	commands map[string]Command
}

type Command struct {
	Name        string
	Description string
	Usage       string
	Handler     func(args []string)
}

func NewCLI() *CLI {
	cli := &CLI{
		session:  NewSession(),
		commands: make(map[string]Command),
	}
	cli.registerCommands()
	return cli
}

func (c *CLI) registerCommands() {
	c.commands["help"] = Command{
		Name:        "help",
		Description: "Show available commands",
		Usage:       "help [command]",
		Handler:     c.cmdHelp,
	}
	c.commands["newgame"] = Command{
		Name:        "newgame",
		Description: "Start a new game session",
		Usage:       "newgame [blinds] (e.g., newgame 25/50)",
		Handler:     c.cmdNewGame,
	}
	c.commands["addplayer"] = Command{
		Name:        "addplayer",
		Description: "Add a player to the game",
		Usage:       "addplayer <name> <chips> <strategy>",
		Handler:     c.cmdAddPlayer,
	}
	c.commands["modplayer"] = Command{
		Name:        "modplayer",
		Description: "Modify an existing player",
		Usage:       "modplayer <name> [chips] [strategy]",
		Handler:     c.cmdModPlayer,
	}
	c.commands["removeplayer"] = Command{
		Name:        "removeplayer",
		Description: "Remove a player from the game",
		Usage:       "removeplayer <name>",
		Handler:     c.cmdRemovePlayer,
	}
	c.commands["randplayers"] = Command{
		Name:        "randplayers",
		Description: "Add random players",
		Usage:       "randplayers [count] [chips]",
		Handler:     c.cmdRandPlayers,
	}
	c.commands["players"] = Command{
		Name:        "players",
		Description: "List all players",
		Usage:       "players",
		Handler:     c.cmdPlayers,
	}
	c.commands["strategies"] = Command{
		Name:        "strategies",
		Description: "List available strategies",
		Usage:       "strategies",
		Handler:     c.cmdStrategies,
	}
	c.commands["run"] = Command{
		Name:        "run",
		Description: "Simulate hands",
		Usage:       "run <count>",
		Handler:     c.cmdRun,
	}
	c.commands["slowsim"] = Command{
		Name:        "slowsim",
		Description: "Simulate hands with step-by-step output",
		Usage:       "slowsim <count>",
		Handler:     c.cmdSlowSim,
	}
	c.commands["results"] = Command{
		Name:        "results",
		Description: "Show simulation results",
		Usage:       "results",
		Handler:     c.cmdResults,
	}
	c.commands["details"] = Command{
		Name:        "details",
		Description: "Show detailed hand history",
		Usage:       "details [hand_number] or details all",
		Handler:     c.cmdDetails,
	}
	c.commands["clear"] = Command{
		Name:        "clear",
		Description: "Clear session and start fresh",
		Usage:       "clear",
		Handler:     c.cmdClear,
	}
	c.commands["status"] = Command{
		Name:        "status",
		Description: "Show current game status",
		Usage:       "status",
		Handler:     c.cmdStatus,
	}
	c.commands["exit"] = Command{
		Name:        "exit",
		Description: "Exit the program",
		Usage:       "exit",
		Handler:     nil,
	}
	c.commands["quit"] = Command{
		Name:        "quit",
		Description: "Exit the program",
		Usage:       "quit",
		Handler:     nil,
	}
}

func (c *CLI) Execute(line string) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return true
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	if cmdName == "exit" || cmdName == "quit" {
		PrintInfo("Goodbye! :3")
		return false
	}

	cmd, exists := c.commands[cmdName]
	if !exists {
		PrintError("Unknown command: %s. Type 'help' for available commands.", cmdName)
		return true
	}

	if cmd.Handler != nil {
		cmd.Handler(args)
	}
	return true
}

func (c *CLI) PrintWelcome() {
	fmt.Println()
	PrintHeader("╔══════════════════════════════════════════╗")
	PrintHeader("║       POKERLIB CLI v%-21s║", version)
	PrintHeader("║   Poker Simulation & Analysis Tool       ║")
	PrintHeader("╚══════════════════════════════════════════╝")
	fmt.Println()
	PrintInfo("Type 'help' for available commands")
	fmt.Println()
}

func (c *CLI) PrintPrompt() {
	PrintPrompt("poker> ")
}
