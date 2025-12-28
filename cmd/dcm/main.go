package main

import (
	"flag"
	"fmt"

	"github.com/Disneyjr/dcm/internal/workspace"
	"github.com/Disneyjr/dcm/utils"
	"github.com/Disneyjr/dcm/utils/messages"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		messages.PrintHelp()
		return
	}

	// Comandos que não precisam de workspace e não devem mostrar ExitMessage
	if args[0] == "version" {
		handleVersionCommand()
		return
	}

	if args[0] == "init" {
		handleInitCommand()
		return
	}

	// Para todos os outros comandos, mostrar ExitMessage ao final
	defer messages.ExitMessage()

	ws := workspace.NewWorkspace()
	if err := workspace.LoadWorkspace(ws); err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		return
	}

	switch args[0] {
	case "up":
		handleUpCommand(ws, args)

	case "down":
		handleDownCommand(ws, args)

	case "restart":
		handleRestartCommand(ws)

	case "logs":
		handleLogsCommand(ws)

	case "status":
		handleStatusCommand(ws)

	case "list":
		handleListCommand(ws)

	case "inspect":
		handleInspectCommand(ws, args)

	case "validate":
		handleValidateCommand(ws)

	default:
		fmt.Printf("%s Comando desconhecido: %s\n", utils.Colorize("yellow", "⚠️"), args[0])
		return
	}
}
