# Tareas de Desarrollo - Fenrir (MCP Plugin)

## Fase 0: Inicialización y Setup
- [x] Ejecutar `go mod init github.com/andragon31/fenrir`
- [x] Crear estructura de directorios base (`cmd`, `internal`, `internal/graph`, `internal/mcp`, `internal/modules/memory`, etc.)
- [x] Configurar `.goreleaser.yaml` inicial y `Makefile`
- [x] Instalar dependencia `modernc.org/sqlite`
- [x] Crear scripts de esquema y migraciones SQLite (`nodes`, `edges`, `sessions`, `audit_log`, `pkg_cache`, `drift_scores`)
- [x] Implementar tabla `drift_scores` (módulo, score, violations, sessions, updated_at)
- [x] Implementar conexión base a SQLite con WAL mode activado
- [x] Agregar warning a 50k nodos, query de conteo en startup
- [x] Implementar capa de abstracción de transport para aislamiento de mcp-go
- [x] Integrar `github.com/mark3labs/mcp-go` e implementar inicialización de servidor por standard I/O (stdio)
- [x] Configurar log estructurado con `charmbracelet/log`
- [x] Implementar manejo de rate limiting para OSV API con backoff exponencial
- [x] Implementar modo offline graceful cuando FENRIR_OFFLINE=true

## Fase 1: Motor del Knowledge Graph y Memory Module
- [x] Crear consultas (CRUD) de la tabla `nodes` y `edges`
- [x] Implementar FTS5 tabla virtual para indexación full-text
- [x] Implementar Tool MCP: `mem_save` (sanitizado de tags `<private>`)
- [x] Implementar Tool MCP: `mem_find`
- [x] Implementar Tool MCP: `mem_timeline` (profundidad de grafo)
- [x] Implementar Tool MCP: `mem_session_start` (retorno de contexto predictivo base)
- [x] Implementar Tool MCP: `mem_session_end` (creación "Session DNA" y métricas)
- [x] Implementar Tool MCP: `mem_context` y `mem_dna`

## Fase 2: Módulo Validator
- [x] Implementar abstracción de memoria en caché SQLite para paquetes (`pkg_cache`)
- [x] Integrar API contra npm (`registry.npmjs.org`)
- [x] Integrar API contra PyPI (`pypi.org/pypi`)
- [x] Integrar API contra crates.io, api.nuget.org
- [x] Implementar consultas de CVE vulnerables contra `api.osv.dev`
- [x] Implementar consultas de compatibilidad de licencia via `api.deps.dev`
- [x] Implementar Tool MCP: `pkg_check`
- [x] Implementar Tool MCP: `pkg_license`
- [x] Implementar Tool MCP: `pkg_audit` (analizando un `package.json`, `requirements.txt`, etc.)

## Fase 3: Módulo Enforcer y Policy Engine
- [x] Crear parser/cargador de JSON para el archivo de políticas `.fenrir/policies.json`
- [x] Implementar lógica Regex y matcheo de reglas arquitectónicas de severity `soft`, `hard` o `critical`
- [x] Implementar cálculo de Drift (Drift_Score) en los módulos/archivos en base a reglas violadas vs aceptadas
- [x] Ejecutar re-cálculos de Drift al finalizado de sesiones (`mem_session_end`)
- [x] Implementar Tool MCP: `arch_save`
- [x] Implementar Tool MCP: `arch_verify`
- [x] Implementar Tool MCP: `arch_drift`
- [x] Implementar Tool MCP: `policy_check`
- [x] Implement Tool MCP: `predict` (basado en historial del módulo)

## Fase 4: Módulo Shield y Auditoría
- [x] Crear middleware/sanitizer que remueva tokens via RegEx pre-persistencia (OpenAI, AWS, JWT, etc.)
- [x] Crear capa de stripping para contenido `<private>...</private>`
- [x] Implementar heurística sencilla para detección de patrones Prompt Injection
- [x] Extender Tool `mem_session_end` u otras para que guarden sus acciones dentro de la tabla `audit_log`
- [x] Implementar Tool MCP: `audit_log`
- [x] Implementar Tool MCP: `session_audit`
- [x] Implementar Tool MCP: `inject_guard`
- [x] Implementar Tool MCP: `fenrir_stats` (vía `statsCmd`)

## Fase 5: Intelligence Engine & Insights
- [x] Escribir consultas SQL agrupadoras agregadas sobre histórico para inferir frecuencia y recurrencia de bugs por path
- [x] Implementar Tool MCP: `insights`
- [x] Implementar Tool MCP: `trace` usando subqueries CTE Recursive para navegar descendencias de decisiones a un fichero
- [x] Implementar Tool MCP: `confidence_update`

## Fase 6: UI, CLI y Adapters
- [x] Configurar CLI root commands con `github.com/spf13/cobra`
- [x] Implementar COMANDOS CLI básicos:
  - [x] `fenrir init`
  - [x] `fenrir mcp` (servidor stdio)
  - [x] `fenrir setup opencode`
  - [x] `fenrir search <query>`
  - [x] `fenrir context [--module <path>]`
  - [x] `fenrir stats`
  - [x] `fenrir version`
  - [x] `fenrir pkg check <name> --eco <eco>`
  - [x] `fenrir pkg license <name>` (stub)
- [x] Implementar comandos avanzados:
  - [x] `fenrir drift [--module <path>]`
  - [x] `fenrir serve [--port 7438]`
  - [x] `fenrir tui`
  - [x] `fenrir insights`
  - [x] `fenrir trace <target>`
  - [x] `fenrir session list`
  - [x] `fenrir session show <id>`
  - [x] `fenrir session audit <id>`
  - [x] `fenrir arch list`
  - [x] `fenrir arch add`
  - [x] `fenrir arch deprecate <id>`
  - [x] `fenrir sync`
  - [x] `fenrir export [--file <path>]`
  - [x] `fenrir import --file <path>`
- [x] **Soporte de Adapters Universales (Inspirado en Engram)**:
  - [x] `fenrir setup claude-code`
  - [x] `fenrir setup gemini-cli`
  - [x] `fenrir setup cursor`
  - [x] `fenrir setup windsurf`
  - [x] `fenrir setup antigravity`
  - [x] `fenrir setup opencode` (con plugin TypeScript)
- [x] Implementar TUI Dashboard completo

## Fase 7: Mecanismo Git Sync y Distribución Final (En Progreso)
- [x] Crear utilidades de serializar contenido exportado vía GZIP estructurado JSONL
- [x] Crear rutina para identificar hash addressable de estado para sync idempotente
- [x] Lógica para manejar comandos `fenrir sync` e import local a través de disco
- [x] **Implementar lógica de merge idempotente** (último timestamp gana)
- [x] **Resolver conflictos de merge** por timestamp
- [ ] Finalizar `.goreleaser.yaml` para distribución cross-platform

## Fase 8: Métricas, Testing y Validación
- [ ] **Implementar métricas de adopción**: contador de instalaciones, sesiones activas
- [ ] **Verificar tiempo de startup < 100ms**
- [ ] **Verificar binario < 20MB**
- [ ] **Verificar tiempo de fenrir init < 60 segundos**
