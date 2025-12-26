package comands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Disneyjr/dcm/internal/workspace"
	"github.com/Disneyjr/dcm/utils"
)

func UpService(workspace *workspace.Workspace, projectName string, verbose bool) error {
	project, exists := workspace.Projects[projectName]
	if !exists {
		return fmt.Errorf("projeto '%s' não encontrado", projectName)
	}

	if verbose {
		fmt.Printf("%s Iniciando %s\n",
			utils.Colorize("blue", "🚀"), projectName)
	}

	cmd := "docker-compose up -d"

	if err := runCommand(project.Path, cmd, !verbose); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("%s ✅ %s pronto!\n", utils.Colorize("green", ""), projectName)
	}
	return nil
}

func UpGroup(workspace *workspace.Workspace, groupName string) error {
	group, exists := workspace.Groups[groupName]
	if !exists {
		return fmt.Errorf("grupo '%s' não encontrado", groupName)
	}

	services, ok := group["services"].([]interface{})
	if !ok {
		return fmt.Errorf("grupo '%s' não tem field 'services'", groupName)
	}

	fmt.Printf("%s Iniciando grupo '%s'...\n\n", utils.Colorize("cyan", "🔄"), groupName)

	for _, s := range services {
		serviceSpec := s.(string)

		parts := strings.Split(serviceSpec, ":")
		projectName := parts[0]

		if err := UpService(workspace, projectName, true); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		}
	}

	fmt.Printf("\n%s ✨ Grupo pronto!\n", utils.Colorize("green", ""))
	return nil
}

func DownAll(workspace *workspace.Workspace) error {
	fmt.Printf("%s Parando todos os serviços...\n\n", utils.Colorize("cyan", "⏹️"))

	for projectName, project := range workspace.Projects {
		fmt.Printf("%s Parando %s\n", utils.Colorize("blue", "🚀"), projectName)
		if err := runCommand(project.Path, "docker-compose down", true); err != nil {
			fmt.Printf("%s Erro em %s: %v\n", utils.Colorize("red", "❌"), projectName, err)
		}
	}

	fmt.Printf("\n%s ✨ Todos parados!\n\n", utils.Colorize("green", ""))
	return nil
}

func RestartAll(workspace *workspace.Workspace) error {
	fmt.Printf("%s Reiniciando todos os serviços...\n\n", utils.Colorize("cyan", "🔄"))

	for projectName, project := range workspace.Projects {
		fmt.Printf("%s Reiniciando %s\n", utils.Colorize("blue", "🚀"), projectName)
		if err := runCommand(project.Path, "docker-compose restart", true); err != nil {
			fmt.Printf("%s Erro em %s: %v\n", utils.Colorize("red", "❌"), projectName, err)
		}
	}

	fmt.Printf("\n%s ✨ Todos reiniciados!\n\n", utils.Colorize("green", ""))
	return nil
}

func StatusAll(workspace *workspace.Workspace) error {
	fmt.Printf("%s Status de todos os serviços:\n\n", utils.Colorize("cyan", "📊"))

	for projectName, project := range workspace.Projects {
		fmt.Printf("%s %s:\n", utils.Colorize("blue", "📌"), projectName)
		if err := runCommand(project.Path, "docker-compose ps", false); err != nil {
			fmt.Printf("%s Erro: %v\n", utils.Colorize("red", "❌"), err)
		}
		fmt.Println()
	}

	return nil
}

func LogsAll(workspace *workspace.Workspace) error {
	fmt.Printf("%s Logs de todos os serviços:\n\n", utils.Colorize("cyan", "📋"))

	for projectName, project := range workspace.Projects {
		fmt.Printf("%s %s:\n", utils.Colorize("blue", "📌"), projectName)
		if err := runCommand(project.Path, "docker-compose logs", false); err != nil {
			fmt.Printf("%s Erro: %v\n", utils.Colorize("red", "❌"), err)
		}
		fmt.Println()
	}

	return nil
}

func ListAll(workspace *workspace.Workspace) {
	fmt.Printf("%s Projetos:\n", utils.Colorize("cyan", "📌"))
	for name, proj := range workspace.Projects {
		fmt.Printf("  - %s: %s\n", name, proj.Description)
	}
	fmt.Printf("\n%s Grupos:\n", utils.Colorize("cyan", "📌"))
	for name := range workspace.Groups {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Println()
}

func runCommand(projectPath string, cmd string, parallel bool) error {
	parts := strings.Fields(cmd)
	c := exec.Command(parts[0], parts[1:]...)
	c.Dir = projectPath

	if !parallel {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	} else {
		c.Stdout = io.Discard
		c.Stderr = io.Discard
	}

	if err := c.Run(); err != nil {
		return fmt.Errorf("erro: %w", err)
	}

	return nil
}
