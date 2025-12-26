package workspace

import (
	"encoding/json"
	"fmt"
	"os"
)

type Project struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type Workspace struct {
	Version  string                            `json:"version"`
	Projects map[string]Project                `json:"projects"`
	Groups   map[string]map[string]interface{} `json:"groups"`
}

func NewWorkspace() *Workspace {
	return &Workspace{}
}

func LoadWorkspace(ws *Workspace) error {
	data, err := os.ReadFile("workspace.json")
	if err != nil {
		return fmt.Errorf("workspace.json não encontrado.")
	}

	if err := json.Unmarshal(data, ws); err != nil {
		return fmt.Errorf("erro ao parsear workspace.json: %w", err)
	}

	return nil
}
