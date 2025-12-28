package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Disneyjr/dcm/internal/commands"
	"github.com/Disneyjr/dcm/utils"
	"github.com/Disneyjr/dcm/utils/messages"
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
func main() {
	fmt.Printf("\n%s DCM - Instalador Global\n\n", utils.Colorize("cyan", "📌"))
	defer messages.ExitMessage()
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		if !utils.IsAdmin() {
			fmt.Println("📌 DCM - Instalador Global")
			fmt.Println("❌ ERRO: Necessário privilégios de Administrador")
			fmt.Println("💡 Clique direito no install.exe > 'Executar como administrador'")
			return
		}
		sourcePath, err := findDCMBinary()
		if err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			fmt.Printf("%s\nUso: Coloque dcm no diretório atual e execute este instalador.\n\n", utils.Colorize("yellow", "💡"))
			return
		}

		fmt.Printf("%s Encontrado: %s\n", utils.Colorize("green", "✅"), sourcePath)

		var installedPath string
		var installErr error
		switch runtime.GOOS {
		case "linux", "darwin":
			installedPath, installErr = commands.InstallLinuxMacOS(sourcePath)
		case "windows":
			installedPath, installErr = commands.InstallWindows(sourcePath)
		default:
			installErr = fmt.Errorf("SO não suportado: %s", runtime.GOOS)
		}

		if installErr != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), installErr)
			return
		}

		if err := commands.VerifyInstallation(installedPath); err != nil {
			fmt.Printf("%s %v\n", utils.Colorize("red", "❌"), err)
			fmt.Printf("\n%s Tente executar manualmente:\n", utils.Colorize("yellow", "💡"))
			fmt.Printf("  Linux/macOS: sudo mv dcm /usr/local/bin/ && sudo chmod +x /usr/local/bin/dcm\n")
			fmt.Printf("  Windows: Mova dcm.exe para C:\\Windows\\System32\\ (execute como Admin)\n\n")
			return
		}

		messages.InstallSuccessful()
	}
}
