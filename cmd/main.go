package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Disneyjr/dcm/internal/comands"
	"github.com/Disneyjr/dcm/internal/workspace"
	"github.com/Disneyjr/dcm/utils"
	"github.com/Disneyjr/dcm/utils/messages"
)

func main() {
	flag.Parse()
	args := flag.Args()
	defer messages.ExitMessage()
	if len(args) == 0 {
		messages.PrintHelp()
		return
	}
	ws := workspace.NewWorkspace()
	if err := workspace.LoadWorkspace(ws); err != nil {
		if args[0] != "install" {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}
	}

	switch args[0] {
	case "up":
		if len(args) < 2 {
			fmt.Println(utils.Colorize("red", "❌ Especifique um projeto ou grupo"))
			os.Exit(1)
		}

		projectOrGroup := args[1]

		if groupErr := comands.UpGroup(ws, projectOrGroup); groupErr == nil {
			return
		}

	case "down":
		if err := comands.DownAll(ws); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "restart":
		if err := comands.RestartAll(ws); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "logs":
		if err := comands.LogsAll(ws); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "status":
		if err := comands.StatusAll(ws); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "list":
		comands.ListAll(ws)

	case "version":
		messages.VersionMessage()

	default:
		fmt.Printf("%s Comando desconhecido: %s\n", utils.Colorize("yellow", "⚠️"), args[0])
		os.Exit(1)
	}
}
