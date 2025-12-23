package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Disneyjr/dcm/utils"
)

var Version = "dev"

type Profile struct {
	Description string   `json:"description"`
	Services    []string `json:"services"`
}

type Project struct {
	Path           string             `json:"path"`
	Type           string             `json:"type"`
	Description    string             `json:"description"`
	Profiles       map[string]Profile `json:"profiles,omitempty"`
	DefaultProfile string             `json:"defaultProfile,omitempty"`
}

type Config struct {
	Version  string                            `json:"version"`
	Projects map[string]Project                `json:"projects"`
	Groups   map[string]map[string]interface{} `json:"groups"`
}

var config *Config

func loadConfig() error {
	data, err := os.ReadFile("services.json")
	if err != nil {
		return fmt.Errorf("services.json não encontrado.")
	}

	config = &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("erro ao parsear services.json: %w", err)
	}

	return nil
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

func upService(projectName string, profile string, verbose bool) error {
	project, exists := config.Projects[projectName]
	if !exists {
		return fmt.Errorf("projeto '%s' não encontrado", projectName)
	}

	if profile == "" {
		profile = project.DefaultProfile
		if profile == "" {
			profile = "default"
		}
	}

	if verbose {
		fmt.Printf("%s Iniciando %s (profile: %s)\n",
			utils.Colorize("blue", "🚀"), projectName, profile)
	}

	cmd := "docker-compose up -d"

	if project.Profiles != nil {
		cmd = fmt.Sprintf("docker-compose --profile %s up -d", profile)
	}

	if err := runCommand(project.Path, cmd, !verbose); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("%s ✅ %s pronto!\n", utils.Colorize("green", ""), projectName)
	}
	return nil
}

func upGroup(groupName string) error {
	group, exists := config.Groups[groupName]
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
		profile := ""
		if len(parts) > 1 {
			profile = parts[1]
		}

		if err := upService(projectName, profile, true); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		}
	}

	fmt.Printf("\n%s ✨ Grupo pronto!\n", utils.Colorize("green", ""))
	return nil
}

func downAll() error {
	fmt.Printf("%s Parando todos os serviços...\n\n", utils.Colorize("cyan", "⏹️"))

	for projectName, project := range config.Projects {
		fmt.Printf("%s Parando %s\n", utils.Colorize("blue", "🚀"), projectName)
		if err := runCommand(project.Path, "docker-compose down", true); err != nil {
			fmt.Printf("%s Erro em %s: %v\n", utils.Colorize("red", "❌"), projectName, err)
		}
	}

	fmt.Printf("\n%s ✨ Todos parados!\n\n", utils.Colorize("green", ""))
	return nil
}

func restartAll() error {
	fmt.Printf("%s Reiniciando todos os serviços...\n\n", utils.Colorize("cyan", "🔄"))

	for projectName, project := range config.Projects {
		fmt.Printf("%s Reiniciando %s\n", utils.Colorize("blue", "🚀"), projectName)
		if err := runCommand(project.Path, "docker-compose restart", true); err != nil {
			fmt.Printf("%s Erro em %s: %v\n", utils.Colorize("red", "❌"), projectName, err)
		}
	}

	fmt.Printf("\n%s ✨ Todos reiniciados!\n\n", utils.Colorize("green", ""))
	return nil
}

func statusAll() error {
	fmt.Printf("%s Status de todos os serviços:\n\n", utils.Colorize("cyan", "📊"))

	for projectName, project := range config.Projects {
		fmt.Printf("%s %s:\n", utils.Colorize("blue", "📌"), projectName)
		if err := runCommand(project.Path, "docker-compose ps", false); err != nil {
			fmt.Printf("%s Erro: %v\n", utils.Colorize("red", "❌"), err)
		}
		fmt.Println()
	}

	return nil
}

func logsAll() error {
	fmt.Printf("%s Logs de todos os serviços:\n\n", utils.Colorize("cyan", "📋"))

	for projectName, project := range config.Projects {
		fmt.Printf("%s %s:\n", utils.Colorize("blue", "📌"), projectName)
		if err := runCommand(project.Path, "docker-compose logs", false); err != nil {
			fmt.Printf("%s Erro: %v\n", utils.Colorize("red", "❌"), err)
		}
		fmt.Println()
	}

	return nil
}
func installSuccessful() error {
	fmt.Printf("%s DCM instalado com sucesso!\n", utils.Colorize("green", "✅"))
	fmt.Printf("%s Você pode usar 'dcm' em qualquer terminal/pasta.\n\n", utils.Colorize("green", "✨"))
	fmt.Printf("Exemplo:\n")
	fmt.Printf("  dcm list\n")
	fmt.Printf("  dcm up dev\n")
	fmt.Printf("  dcm version\n\n")

	return nil
}

func listAll() {
	fmt.Printf("%s Projetos:\n", utils.Colorize("cyan", "📌"))
	for name, proj := range config.Projects {
		fmt.Printf("  - %s: %s\n", name, proj.Description)
		if proj.Profiles != nil {
			for profName := range proj.Profiles {
				fmt.Printf("      • %s\n", profName)
			}
		}
	}
	fmt.Printf("\n%s Grupos:\n", utils.Colorize("cyan", "📌"))
	for name := range config.Groups {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Println()
}

func printHelp() {
	fmt.Printf("%s DCM - Docker Compose Manager\n\n", utils.Colorize("cyan", "📌"))
	fmt.Printf("Versão: %s\n\n", Version)
	fmt.Println("Uso:")
	fmt.Println("  dcm up <projeto> [profile]    - Inicia projeto com profile")
	fmt.Println("  dcm up <grupo>                - Inicia grupo")
	fmt.Println("  dcm down                      - Para todos os serviços")
	fmt.Println("  dcm restart                   - Reinicia todos")
	fmt.Println("  dcm logs                      - Mostra logs")
	fmt.Println("  dcm status                    - Status dos serviços")
	fmt.Println("  dcm list                      - Lista projetos e grupos")
	fmt.Println("  dcm version                   - Mostra versão\n")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		printHelp()
		return
	}

	if err := loadConfig(); err != nil {
		if args[0] != "install" {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}
	}

	switch args[0] {
	case "install":
		if err := installSuccessful(); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}
	case "up":
		if len(args) < 2 {
			fmt.Println(utils.Colorize("red", "❌ Especifique um projeto ou grupo"))
			os.Exit(1)
		}

		projectOrGroup := args[1]
		profile := ""

		if len(args) > 2 {
			profile = args[2]
		}

		if groupErr := upGroup(projectOrGroup); groupErr == nil {
			return
		}

		if err := upService(projectOrGroup, profile, true); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "down":
		if err := downAll(); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "restart":
		if err := restartAll(); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "logs":
		if err := logsAll(); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "status":
		if err := statusAll(); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			os.Exit(1)
		}

	case "list":
		listAll()

	case "version":
		fmt.Printf("dcm v%s\n", Version)

	default:
		fmt.Printf("%s Comando desconhecido: %s\n", utils.Colorize("yellow", "⚠️"), args[0])
		os.Exit(1)
	}
}
