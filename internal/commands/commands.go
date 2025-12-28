package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Disneyjr/dcm/internal/workspace"
	"github.com/Disneyjr/dcm/utils"
)

var DryRun = false

func UpService(workspace *workspace.Workspace, serviceSpec string, verbose bool, extraArgs ...string) error {
	parts := strings.Split(serviceSpec, ":")
	projectName := parts[0]
	targetService := ""
	if len(parts) > 1 {
		targetService = parts[1]
	}

	project, exists := workspace.Projects[projectName]
	if !exists {
		return fmt.Errorf("projeto '%s' não encontrado", projectName)
	}

	if verbose {
		statusMsg := projectName
		if targetService != "" {
			statusMsg = fmt.Sprintf("%s:%s", projectName, targetService)
		}
		fmt.Printf("%s Iniciando %s\n", utils.Colorize("blue", "🚀"), statusMsg)
	}

	args := []string{"up", "-d"}
	args = append(args, extraArgs...)
	if targetService != "" {
		args = append(args, targetService)
	}

	if err := runCommand(project.Path, "docker-compose", args, !verbose); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("%s ✅ %s pronto!\n", utils.Colorize("green", ""), projectName)
	}
	return nil
}

func resolveGroupServices(ws *workspace.Workspace, groupName string, visited map[string]bool) ([]string, bool, error) {
	if visited[groupName] {
		return nil, true, fmt.Errorf("ciclo de herança detectado no grupo '%s'", groupName)
	}
	visited[groupName] = true

	group, exists := ws.Groups[groupName]
	if !exists {
		return nil, true, fmt.Errorf("grupo '%s' não encontrado", groupName)
	}

	var allServices []string
	parallel := true
	if group.Parallel != nil {
		parallel = *group.Parallel
	}

	if group.Extends != "" {
		inheritedServices, inheritedParallel, err := resolveGroupServices(ws, group.Extends, visited)
		if err != nil {
			return nil, true, err
		}
		allServices = append(allServices, inheritedServices...)
		// Se qualquer grupo na cadeia for sequencial, mantemos a intenção (opcional, mas seguro)
		if !inheritedParallel {
			parallel = false
		}
	}

	allServices = append(allServices, group.Services...)
	return allServices, parallel, nil
}

func UpGroup(workspace *workspace.Workspace, groupName string, extraArgs ...string) error {
	services, parallel, err := resolveGroupServices(workspace, groupName, make(map[string]bool))
	if err != nil {
		return err
	}

	fmt.Printf("%s Iniciando grupo '%s' (parallel=%v)...\n\n", utils.Colorize("cyan", "🔄"), groupName, parallel)

	if !parallel {
		for _, serviceSpec := range services {
			if err := UpService(workspace, serviceSpec, true, extraArgs...); err != nil {
				fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			}
		}
	} else {
		var wg sync.WaitGroup
		errChan := make(chan error, len(services))

		for _, s := range services {
			wg.Add(1)
			go func(spec string) {
				defer wg.Done()
				if err := UpService(workspace, spec, true, extraArgs...); err != nil {
					errChan <- fmt.Errorf("%s %v", utils.Colorize("red", "❌"), err)
				}
			}(s)
		}

		wg.Wait()
		close(errChan)

		hasError := false
		for err := range errChan {
			fmt.Println(err)
			hasError = true
		}

		if hasError {
			return fmt.Errorf("alguns serviços falharam ao iniciar")
		}
	}

	fmt.Printf("\n%s ✨ Grupo pronto!\n", utils.Colorize("green", ""))
	return nil
}

func DownAll(workspace *workspace.Workspace, removeVolumes bool) error {
	volumeMsg := ""
	if removeVolumes {
		volumeMsg = " e removendo volumes"
	}
	fmt.Printf("%s Parando todos os serviços%s...\n\n", utils.Colorize("cyan", "⏹️"), volumeMsg)

	for projectName, project := range workspace.Projects {
		fmt.Printf("%s Parando %s\n", utils.Colorize("blue", "🚀"), projectName)

		args := []string{"down"}
		if removeVolumes {
			args = append(args, "-v")
		}

		if err := runCommand(project.Path, "docker-compose", args, true); err != nil {
			fmt.Printf("%s Erro em %s: %v\n", utils.Colorize("red", "❌"), projectName, err)
		}
	}

	fmt.Printf("\n%s ✨ Todos parados!\n\n", utils.Colorize("green", ""))
	return nil
}

func DownGroup(workspace *workspace.Workspace, groupName string, removeVolumes bool) error {
	services, _, err := resolveGroupServices(workspace, groupName, make(map[string]bool))
	if err != nil {
		return err
	}

	volumeMsg := ""
	if removeVolumes {
		volumeMsg = " e removendo volumes"
	}
	fmt.Printf("%s Parando grupo '%s'%s...\n\n", utils.Colorize("cyan", "⏹️"), groupName, volumeMsg)

	for _, serviceSpec := range services {
		parts := strings.Split(serviceSpec, ":")
		projectName := parts[0]

		project, exists := workspace.Projects[projectName]
		if !exists {
			fmt.Printf("%s Projeto '%s' não encontrado\n", utils.Colorize("red", "❌"), projectName)
			continue
		}

		fmt.Printf("%s Parando %s\n", utils.Colorize("blue", "🚀"), projectName)

		args := []string{"down"}
		if removeVolumes {
			args = append(args, "-v")
		}

		if err := runCommand(project.Path, "docker-compose", args, true); err != nil {
			fmt.Printf("%s Erro em %s: %v\n", utils.Colorize("red", "❌"), projectName, err)
		}
	}

	fmt.Printf("\n%s ✨ Grupo '%s' parado!\n\n", utils.Colorize("green", ""), groupName)
	return nil
}

func RestartAll(workspace *workspace.Workspace) error {
	fmt.Printf("%s Reiniciando todos os serviços...\n\n", utils.Colorize("cyan", "🔄"))

	for projectName, project := range workspace.Projects {
		fmt.Printf("%s Reiniciando %s\n", utils.Colorize("blue", "🚀"), projectName)
		if err := runCommand(project.Path, "docker-compose", []string{"restart"}, true); err != nil {
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
		if err := runCommand(project.Path, "docker-compose", []string{"ps"}, false); err != nil {
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
		if err := runCommand(project.Path, "docker-compose", []string{"logs"}, false); err != nil {
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

func InspectGroup(ws *workspace.Workspace, groupName string) {
	services, parallel, err := resolveGroupServices(ws, groupName, make(map[string]bool))
	if err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		return
	}

	fmt.Printf("%s Inspeção do grupo: %s\n", utils.Colorize("cyan", "🔍"), groupName)
	fmt.Printf("Configuração: parallel=%v\n\n", parallel)
	fmt.Printf("Serviços na ordem de execução:\n")
	for i, spec := range services {
		parts := strings.Split(spec, ":")
		projectName := parts[0]
		targetService := "todos"
		if len(parts) > 1 {
			targetService = parts[1]
		}

		project := ws.Projects[projectName]
		fmt.Printf("%d. %s\n", i+1, utils.Colorize("blue", spec))
		fmt.Printf("   Caminho: %s\n", project.Path)
		fmt.Printf("   Serviço: %s\n", targetService)
	}
	fmt.Println()
}

func ValidateWorkspace(ws *workspace.Workspace) {
	fmt.Printf("%s Validando workspace.json...\n", utils.Colorize("cyan", "🔍"))
	hasError := false

	for name, proj := range ws.Projects {
		if _, err := os.Stat(proj.Path); os.IsNotExist(err) {
			fmt.Printf("%s Projeto '%s': caminho não encontrado: %s\n", utils.Colorize("red", "❌"), name, proj.Path)
			hasError = true
		}
	}

	for name, group := range ws.Groups {
		for _, spec := range group.Services {
			parts := strings.Split(spec, ":")
			if _, exists := ws.Projects[parts[0]]; !exists {
				fmt.Printf("%s Grupo '%s': projeto '%s' não definido\n", utils.Colorize("red", "❌"), name, parts[0])
				hasError = true
			}
		}
		if group.Extends != "" {
			if _, exists := ws.Groups[group.Extends]; !exists {
				fmt.Printf("%s Grupo '%s': estende grupo inexistente '%s'\n", utils.Colorize("red", "❌"), name, group.Extends)
				hasError = true
			}
		}
	}

	if !hasError {
		fmt.Printf("%s Workspace válido!\n", utils.Colorize("green", "✅"))
	}
}

func Uninstall() {
	targetPath := filepath.Join(os.Getenv("WINDIR"), "System32", "dcm.exe")
	_, err := os.Stat(targetPath)
	if err != nil {
		fmt.Println("dcm não encontrado!")
		fmt.Println(err.Error())
		fmt.Println("Uma pena que o dcm não atendeu o seu projeto!")
		return
	}

	fmt.Printf("🗑️  Removendo %s\n", targetPath)
	os.Remove(targetPath)

	fmt.Println("✅ dcm desinstalado!")
	fmt.Println("Uma pena que o dcm não atendeu o seu projeto!")
}

func InitWorkspace() error {
	filePath := "workspace.json"
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("workspace.json já existe")
	}

	content := `{
  "version": "1.0",
  "projects": {
    "exemplo": { "path": "./services/exemplo", "description": "Projeto de exemplo" }
  },
  "groups": {
    "dev": { "services": ["exemplo"] }
  }
}
`
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("erro ao criar workspace.json: %w", err)
	}

	fmt.Printf("%s workspace.json criado com sucesso!\n", utils.Colorize("green", "✅"))
	return nil
}

func runCommand(projectPath string, command string, args []string, parallel bool) error {
	if DryRun {
		fmt.Printf("%s [DRY-RUN] cd %s && %s %s\n", utils.Colorize("yellow", "🛠️"), projectPath, command, strings.Join(args, " "))
		return nil
	}

	c := exec.Command(command, args...)
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

func InstallLinuxMacOS(sourcePath string) (string, error) {
	fmt.Printf("%s Detectado: %s\n", utils.Colorize("cyan", "🔍"), utils.GetSystemInfo())
	fmt.Printf("%s Instalando DCM globalmente...\n\n", utils.Colorize("blue", "🚀"))

	targetPath := "/usr/local/bin/dcm"

	fmt.Printf("%s Copiando binário para %s\n", utils.Colorize("cyan", "📁"), targetPath)

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(targetPath)
	if err != nil {
		fmt.Printf("%s Permissão negada, tentando com sudo...\n", utils.Colorize("yellow", "⚠️"))

		cmd := exec.Command("sudo", "tee", targetPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("erro ao criar pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return "", fmt.Errorf("erro ao executar sudo: %w", err)
		}

		if _, err := io.Copy(stdinPipe, srcFile); err != nil {
			return "", fmt.Errorf("erro ao copiar arquivo: %w", err)
		}

		stdinPipe.Close()

		if err := cmd.Wait(); err != nil {
			return "", fmt.Errorf("erro ao finalizar cópia: %w", err)
		}

		fmt.Printf("%s Ajustando permissões...\n", utils.Colorize("cyan", "🔒"))
		chmodCmd := exec.Command("sudo", "chmod", "+x", targetPath)
		if err := chmodCmd.Run(); err != nil {
			return "", fmt.Errorf("erro ao ajustar permissões: %w", err)
		}
	} else {
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return "", fmt.Errorf("erro ao copiar conteúdo: %w", err)
		}

		fmt.Printf("%s Ajustando permissões...\n", utils.Colorize("cyan", "🔒"))
		if err := os.Chmod(targetPath, 0755); err != nil {
			return "", fmt.Errorf("erro ao ajustar permissões: %w", err)
		}
	}

	return targetPath, nil
}

func InstallWindows(sourcePath string) (string, error) {
	fmt.Printf("%s Detectado: %s\n", utils.Colorize("cyan", "🔍"), utils.GetSystemInfo())
	fmt.Printf("%s Instalando DCM globalmente...\n\n", utils.Colorize("blue", "🚀"))

	targetPath := filepath.Join(os.Getenv("WINDIR"), "System32", "dcm.exe")

	fmt.Printf("%s Copiando para: %s\n", utils.Colorize("cyan", "📁"), targetPath)

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("erro ao criar destino (pode precisar executar como Admin): %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("erro ao copiar: %w", err)
	}

	return targetPath, nil
}

func VerifyInstallation(installedPath string) error {
	fmt.Printf("\n%s Validando instalação...\n", utils.Colorize("cyan", "✓"))

	if installedPath == "" {
		cmdSearch := exec.Command("which", "dcm")
		if runtime.GOOS == "windows" {
			cmdSearch = exec.Command("where", "dcm.exe")
		}

		output, err := cmdSearch.Output()
		if err != nil {
			return fmt.Errorf("DCM não encontrado no PATH")
		}
		installedPath = strings.TrimSpace(string(output))
		// 'where' can return multiple paths on Windows, use the first one
		if runtime.GOOS == "windows" {
			paths := strings.Split(installedPath, "\r\n")
			if len(paths) > 0 {
				installedPath = paths[0]
			}
		}
	}

	fmt.Printf("%s Encontrado em: %s\n", utils.Colorize("green", "✅"), installedPath)

	// Use absolute path to run the version command to avoid Go's relative path execution security check
	testCmd := exec.Command(installedPath, "version")
	output, err := testCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao executar '%s version': %w\nSaída: %s", installedPath, err, string(output))
	}

	return nil
}
