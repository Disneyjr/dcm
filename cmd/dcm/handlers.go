package main

import (
	"fmt"
	"strings"

	"github.com/Disneyjr/dcm/internal/commands"
	"github.com/Disneyjr/dcm/internal/workspace"
	"github.com/Disneyjr/dcm/utils"
	"github.com/Disneyjr/dcm/utils/messages"
)

func handleUpCommand(ws *workspace.Workspace, args []string) {
	if len(args) < 2 {
		fmt.Println(utils.Colorize("red", "❌ Especifique um projeto ou grupo"))
		return
	}

	projectOrGroup := args[1]
	extraArgs := []string{}
	for i := 2; i < len(args); i++ {
		if args[i] == "--build" {
			extraArgs = append(extraArgs, "--build")
		}
		if args[i] == "--dry-run" {
			commands.DryRun = true
		}
	}

	if groupErr := commands.UpGroup(ws, projectOrGroup, extraArgs...); groupErr == nil {
		return
	}
}

func handleDownCommand(ws *workspace.Workspace, args []string) {
	removeVolumes := false
	groupName := ""

	// Parse arguments
	for i := 1; i < len(args); i++ {
		if args[i] == "-v" {
			removeVolumes = true
		} else if !strings.HasPrefix(args[i], "-") {
			groupName = args[i]
		}
	}

	// If group is specified, use DownGroup
	if groupName != "" {
		if err := commands.DownGroup(ws, groupName, removeVolumes); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			return
		}
		return
	}

	// Otherwise, use DownAll
	if err := commands.DownAll(ws, removeVolumes); err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		return
	}
}

func handleRestartCommand(ws *workspace.Workspace) {
	if err := commands.RestartAll(ws); err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		return
	}
}

func handleLogsCommand(ws *workspace.Workspace) {
	if err := commands.LogsAll(ws); err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		return
	}
}

func handleStatusCommand(ws *workspace.Workspace) {
	if err := commands.StatusAll(ws); err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		return
	}
}

func handleListCommand(ws *workspace.Workspace) {
	commands.ListAll(ws)
}

func handleInspectCommand(ws *workspace.Workspace, args []string) {
	if len(args) < 2 {
		fmt.Println(utils.Colorize("red", "❌ Especifique um grupo para inspecionar"))
		return
	}
	commands.InspectGroup(ws, args[1])
}

func handleValidateCommand(ws *workspace.Workspace) {
	commands.ValidateWorkspace(ws)
}

func handleInitCommand() {
	if err := commands.InitWorkspace(); err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		return
	}
}

func handleVersionCommand() {
	messages.VersionMessage()
}
