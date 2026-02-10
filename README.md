# Monolito Modular (DDD/Clean/Hexagonal)

Objetivo: separar o dominio em contextos e reduzir acoplamento.

Estrutura sugerida:
- `cmd/api`
- `internal/domain` entidades e contratos
- `internal/usecase` casos de uso
- `internal/adapters` repositorios, filas, APIs
- `http/` handlers e rotas

Foco: mutabilidade e testabilidade nas ondas B e C.
