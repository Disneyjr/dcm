package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Disneyjr/dcm/utils"
)

func findDCMBinary() (string, error) {
	baseName := "dcm"
	if runtime.GOOS == "windows" {
		baseName = "dcm.exe"
	}
	if _, err := os.Stat(baseName); err == nil {
		abs, _ := filepath.Abs(baseName)
		return abs, nil
	}

	return "", fmt.Errorf("binário '%s' não encontrado no diretório atual", baseName)
}

func installLinuxMacOS(sourcePath string) error {
	fmt.Printf("%s Detectado: %s\n", utils.Colorize("cyan", "🔍"), utils.GetSystemInfo())
	fmt.Printf("%s Instalando DCM globalmente...\n\n", utils.Colorize("blue", "🚀"))

	targetPath := "/usr/local/bin/dcm"

	fmt.Printf("%s Copiando binário para %s\n", utils.Colorize("cyan", "📁"), targetPath)

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
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
			return fmt.Errorf("erro ao criar pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("erro ao executar sudo: %w", err)
		}

		if _, err := io.Copy(stdinPipe, srcFile); err != nil {
			return fmt.Errorf("erro ao copiar arquivo: %w", err)
		}

		stdinPipe.Close()

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("erro ao finalizar cópia: %w", err)
		}

		fmt.Printf("%s Ajustando permissões...\n", utils.Colorize("cyan", "🔒"))
		chmodCmd := exec.Command("sudo", "chmod", "+x", targetPath)
		if err := chmodCmd.Run(); err != nil {
			return fmt.Errorf("erro ao ajustar permissões: %w", err)
		}
	} else {
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return fmt.Errorf("erro ao copiar conteúdo: %w", err)
		}

		fmt.Printf("%s Ajustando permissões...\n", utils.Colorize("cyan", "🔒"))
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("erro ao ajustar permissões: %w", err)
		}
	}

	return nil
}

func installWindows(sourcePath string) error {
	fmt.Printf("%s Detectado: %s\n", utils.Colorize("cyan", "🔍"), utils.GetSystemInfo())
	fmt.Printf("%s Instalando DCM globalmente...\n\n", utils.Colorize("blue", "🚀"))

	targetPath := filepath.Join(os.Getenv("WINDIR"), "System32", "dcm.exe")

	fmt.Printf("%s Copiando para: %s\n", utils.Colorize("cyan", "📁"), targetPath)

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("erro ao criar destino (pode precisar executar como Admin): %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("erro ao copiar: %w", err)
	}

	return nil
}

func verifyInstallation() error {
	fmt.Printf("\n%s Validando instalação...\n", utils.Colorize("cyan", "✓"))

	cmd := exec.Command("which", "dcm")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("where", "dcm.exe")
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("DCM não encontrado no PATH")
	}

	installedPath := strings.TrimSpace(string(output))
	fmt.Printf("%s Encontrado em: %s\n", utils.Colorize("green", "✅"), installedPath)

	testCmd := exec.Command("dcm", "install")
	if runtime.GOOS == "windows" {
		testCmd = exec.Command("dcm.exe", "install")
	}

	output, err = testCmd.Output()
	if err != nil {
		return fmt.Errorf("erro ao executar 'dcm install': %w", err)
	}

	return nil
}

func main() {
	fmt.Printf("\n%s DCM - Instalador Global\n\n", utils.Colorize("cyan", "📌"))

	sourcePath, err := findDCMBinary()
	if err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		fmt.Printf("%s\nUso: Coloque dcm no diretório atual e execute este instalador.\n\n", utils.Colorize("yellow", "💡"))
		os.Exit(1)
	}

	fmt.Printf("%s Encontrado: %s\n", utils.Colorize("green", "✅"), sourcePath)

	var installErr error
	switch runtime.GOOS {
	case "linux", "darwin":
		installErr = installLinuxMacOS(sourcePath)
	case "windows":
		installErr = installWindows(sourcePath)
	default:
		installErr = fmt.Errorf("SO não suportado: %s", runtime.GOOS)
	}

	if installErr != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), installErr)
		os.Exit(1)
	}

	if err := verifyInstallation(); err != nil {
		fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
		fmt.Printf("\n%s Tente executar manualmente:\n", utils.Colorize("yellow", "💡"))
		fmt.Printf("  Linux/macOS: sudo mv dcm /usr/local/bin/ && sudo chmod +x /usr/local/bin/dcm\n")
		fmt.Printf("  Windows: Move dcm.exe para C:\\Windows\\System32\\ (execute como Admin)\n\n")
		os.Exit(1)
	}

	fmt.Printf("\n%s Instalação concluída com sucesso!\n", utils.Colorize("green", "🎉"))
	fmt.Printf("%s Você pode usar 'dcm' em qualquer terminal/pasta.\n\n", utils.Colorize("green", "✨"))
	fmt.Printf("Exemplo:\n")
	fmt.Printf("  dcm list\n")
	fmt.Printf("  dcm up dev\n")
	fmt.Printf("  dcm version\n\n")
}
