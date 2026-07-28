# Contributing to NOVA

We welcome contributions to **NOVA**! Please read the following developer guidelines to understand our architecture, development setup, and how to add new features, tools, personas, or LLM providers.

---

## Development Setup

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/awanmh/Nova.git
   cd Nova
   ```

2. **Verify Go Environment (Go 1.22+)**:
   ```bash
   go version
   ```

3. **Run Unit Tests Across All Modules**:
   ```bash
   go test -v ./...
   ```

4. **Run Static Analysis & Vet**:
   ```bash
   go vet ./...
   ```

5. **Build and Test CLI Locally**:
   ```bash
   go build -o nova ./cmd/nova
   ./nova version
   ./nova doctor
   ```

---

## Architecture & Coding Standards

- **Hexagonal Architecture**: Keep domain logic inside `internal/` decoupled from external frameworks.
- **Interface-Driven**: All providers, tools, and storage backends must implement their respective interfaces (`llm.Provider`, `tools.Tool`, `memory.Engine`).
- **Security First**:
  - Never hardcode credentials, secrets, or API tokens.
  - Always pass tool outputs through `security.Redact` or use the built-in `tools.Executor` which automatically redacts secrets.
  - Check file access against `security.ValidateWorkspacePath` to prevent directory traversal outside the workspace.

---

## Adding a New Tool

To add a new tool to NOVA's standard toolset:
1. Create a new struct implementing `tools.Tool` in `internal/tools/<toolname>.go`:
   ```go
   type Tool interface {
       Name() string
       Description() string
       Parameters() string // JSON Schema definition
       Execute(ctx context.Context, argsJSON string) (*Response, error)
   }
   ```
2. Classify its safety risk in `internal/permission/classifier.go` (`RiskLevelRead`, `RiskLevelLow`, `RiskLevelHigh`, `RiskLevelCritical`).
3. Register your tool in `tools.RegisterStandardTools()` inside `internal/tools/standard.go`.
4. Add unit tests in `internal/tools/<toolname>_test.go`.

---

## Adding a New Engineering Persona

### Custom Workspace Persona (Markdown/JSON)
Create a `.md` file inside any workspace at `.nova/personas/<name>.md`:
```markdown
# Cloud Infrastructure Architect
You are an expert in Terraform, AWS, and Kubernetes deployment pipelines.
Always verify IAM roles and security groups before deploying.
```

### Built-in Canonical Persona (Go)
1. Edit `Builtin()` in `internal/persona/persona.go`.
2. Define its name, description, system rules, default tools, and `TokenBudgetModifier`.
3. Add a verification test in `internal/persona/persona_test.go`.

---

## Adding a New LLM Provider

1. Implement the `llm.Provider` interface in `internal/llm/<provider>/<provider>.go`:
   ```go
   type Provider interface {
       Name() string
       Health(ctx context.Context) error
       ListModels(ctx context.Context) ([]Model, error)
       Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
       ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)
   }
   ```
2. Register the provider in `internal/llm/registry.go`.
3. Add mock-based unit tests for your provider in `internal/llm/<provider>/<provider>_test.go`.

---

## Pull Request Process

1. Ensure `go test -v ./...` and `go vet ./...` pass with 0 errors.
2. Update documentation (`README.md`, docstrings) if adding user-facing features or CLI commands.
3. Submit your PR with a clear description of the problem solved and verification steps.
