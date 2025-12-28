package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Project struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type Group struct {
	Services []string `json:"services"`
	Extends  string   `json:"extends,omitempty"`
	Parallel *bool    `json:"parallel,omitempty"` // Use pointer to distinguish between false and not set
}

type Workspace struct {
	Version  string             `json:"version"`
	Projects map[string]Project `json:"projects"`
	Groups   map[string]Group   `json:"groups"`
	BaseDir  string             `json:"-"` // Diretório base do workspace (onde o workspace.json foi encontrado)
}

func NewWorkspace() *Workspace {
	return &Workspace{}
}

func findWorkspaceFile() (string, string, error) {
	curr, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	for {
		path := filepath.Join(curr, "workspace.json")
		if _, err := os.Stat(path); err == nil {
			return path, curr, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	return "", "", fmt.Errorf("workspace.json não encontrado nos diretórios pais")
}

func LoadWorkspace(ws *Workspace) error {
	path, baseDir, err := findWorkspaceFile()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("não foi possível ler %s: %w", path, err)
	}

	if err := json.Unmarshal(data, ws); err != nil {
		return fmt.Errorf("erro ao parsear workspace.json: %w", err)
	}

	ws.BaseDir = baseDir

	// Resolver caminhos dos projetos relativos ao BaseDir
	for name, proj := range ws.Projects {
		if !filepath.IsAbs(proj.Path) {
			proj.Path = filepath.Clean(filepath.Join(baseDir, proj.Path))
			ws.Projects[name] = proj
		}
	}

	return nil
}
