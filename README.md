# DCM - Docker Compose Manager

Gerencie múltiplos serviços Docker com **um único comando**.

## O Problema

```bash
# ❌ Sem DCM: Múltiplos terminais
Terminal 1: cd services/service-a && docker-compose up
Terminal 2: cd services/service-b && docker-compose up
Terminal 3: cd services/service-c && docker-compose up
# ... e assim vai
```

## A Solução

```bash
# ✅ Com DCM: Um comando
dcm up dev
```

## Características

- 🚀 **Comando único** - Inicie todos os serviços de uma vez
- 📦 **Sem dependências** - Binário standalone
- 🎯 **Grupos** - Organize serviços em combinações
- 🖥️ **Cross-platform** - Linux, macOS, Windows

## Quick Start

### 1. Instalar

```bash
# Baixe o binário em: https://github.com/Disneyjr/dcm/releases

# Linux/macOS
chmod +x install
./install

# Windows
double-click install.exe
```

### 2. Configurar

Crie `workspace.json` na raiz do projeto:

```json
{
  "version": "1.0",
  "projects": {
    "database": { "path": "./infra/db", "type": "simple" },
    "api": { "path": "./services/api", "type": "simple" },
    "web": { "path": "./services/web", "type": "simple" }
  },
  "groups": {
    "dev": { "services": ["database", "api", "web"] }
  }
}
```

### 3. Usar

```bash
dcm list      # Ver projetos e grupos
dcm up dev    # Iniciar grupo completo
dcm down      # Parar tudo
```

## Exemplos Rápidos

**Iniciar um grupo:**
```bash
dcm up dev          # Todos os serviços do grupo 'dev'
```

**Gerenciar serviços:**
```bash
dcm logs            # Ver logs de tudo
dcm status          # Status dos containers
dcm restart         # Reiniciar tudo
```

## Estrutura do Projeto

```
dcm/
├── cmd/                           # Código-fonte
│   ├── main.go                   # CLI principal (dcm)
│   └── install.go                # Instalador
├── utils/                        # Utilitários
├── .github/workflows/            # CI/CD
│   └── release.yml
├── .goreleaser.yaml             # Config para releases automáticos
├── .gitignore                   # Ignore patterns
├── DEVELOPMENT.md               # Guia para desenvolvedores
├── README.md                    # Este arquivo
├── LICENSE
└── go.mod / go.sum             # Dependências Go
```

## Contribuindo

1. Fork o repositório
2. Crie uma branch: `git checkout -b feature/minha-feature`
3. Faça seus commits: `git commit -m "feat: descrição"`
4. Push: `git push origin feature/minha-feature`
5. Abra um Pull Request

## Licença

MIT - Use livremente!

## Suporte

Dúvidas ou problemas? Abra uma [issue](https://github.com/Disneyjr/dcm/issues) 🚀