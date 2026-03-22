# Tareas de Desarrollo - Fenrir (MCP Plugin)

## Fase 0: Inicialización y Setup
- [ ] Ejecutar `go mod init github.com/TU_ORG/fenrir`
- [ ] Crear estructura de directorios base (`cmd`, `internal`, `internal/graph`, `internal/mcp`, `internal/modules/memory`, etc.)
- [ ] Configurar `.goreleaser.yaml` inicial y `Makefile`
- [ ] Instalar dependencia `modernc.org/sqlite`
- [ ] Crear scripts de esquema y migraciones SQLite (`nodes`, `edges`, `sessions`, `audit_log`, `pkg_cache`, `drift_scores`)
- [ ] Implementar tabla `drift_scores` (módulo, score, violations, sessions, updated_at)
- [ ] Implementar conexión base a SQLite con WAL mode activado
- [ ] Agregar warning a 50k nodos, query de conteo en startup
- [ ] Implementar capa de abstracción de transport para aislamiento de mcp-go
- [ ] Integrar `github.com/mark3labs/mcp-go` e implementar inicialización de servidor por standard I/O (stdio)
- [ ] Configurar log estructurado con `charmbracelet/log`
- [ ] Implementar manejo de rate limiting para OSV API con backoff exponencial
- [ ] Implementar modo offline graceful cuando FENRIR_OFFLINE=true

## Fase 1: Motor del Knowledge Graph y Memory Module
- [ ] Crear consultas (CRUD) de la tabla `nodes` y `edges`
- [ ] Implementar FTS5 tabla virtual para indexación full-text
- [ ] Implementar Tool MCP: `mem_save` (sanitizado de tags `<private>`)
- [ ] Implementar Tool MCP: `mem_find`
- [ ] Implementar Tool MCP: `mem_timeline` (profundidad de grafo)
- [ ] Implementar Tool MCP: `mem_session_start` (retorno de contexto predictivo base)
- [ ] Implementar Tool MCP: `mem_session_end` (creación "Session DNA" y métricas)
- [ ] Implementar Tool MCP: `mem_context` y `mem_dna`

## Fase 2: Módulo Validator
- [ ] Implementar abstracción de memoria en caché SQLite para paquetes (`pkg_cache`)
- [ ] Integrar ping HTTP contra npm (`registry.npmjs.org`)
- [ ] Integrar ping HTTP contra PyPI (`pypi.org/pypi`)
- [ ] Integrar ping HTTP contra crates.io, api.nuget.org
- [ ] Implementar consultas de CVE vulnerables contra `api.osv.dev`
- [ ] Implementar consultas de compatibilidad de licencia via `api.deps.dev`
- [ ] Implementar Tool MCP: `pkg_check`
- [ ] Implementar Tool MCP: `pkg_license`
- [ ] Implementar Tool MCP: `pkg_audit` (analizando un `package.json`, `requirements.txt`, etc.)

## Fase 3: Módulo Enforcer y Policy Engine
- [ ] Crear parser/cargador de JSON para el archivo de políticas `.fenrir/policies.json`
- [ ] Implementar lógica Regex y matcheo de reglas arquitectónicas de severity `soft`, `hard` o `critical`
- [ ] Implementar cálculo de Drift (Drift_Score) en los módulos/archivos en base a reglas violadas vs aceptadas
- [ ] Ejecutar re-cálculos de Drift al finalizado de sesiones (`mem_session_end`)
- [ ] Implementar Tool MCP: `arch_save`
- [ ] Implementar Tool MCP: `arch_verify`
- [ ] Implementar Tool MCP: `arch_drift`
- [ ] Implementar Tool MCP: `policy_check`
- [ ] Implementar Tool MCP: `predict` (basado en historial del módulo)

## Fase 4: Módulo Shield y Auditoría
- [ ] Crear middleware/sanitizer que remueva tokens via RegEx pre-persistencia (OpenAI, AWS, JWT, etc.)
- [ ] Crear capa de stripping para contenido `<private>...</private>`
- [ ] Implementar heurística sencilla para detección de patrones Prompt Injection
- [ ] Extender Tool `mem_session_end` u otras para que guarden sus acciones dentro de la tabla `audit_log`
- [ ] Implementar Tool MCP: `audit_log`
- [ ] Implementar Tool MCP: `session_audit`
- [ ] Implementar Tool MCP: `inject_guard`
- [ ] Implementar Tool MCP: `fenrir_stats`

## Fase 5: Intelligence Engine & Insights
- [ ] Escribir consultas SQL agrupadoras agregadas sobre histórico para inferir frecuencia y recurrencia de bugs por path
- [ ] Implementar Tool MCP: `insights`
- [ ] Implementar Tool MCP: `trace` usando subqueries CTE Recursive para navegar descendencias de decisiones a un fichero
- [ ] Implementar Tool MCP: `confidence_update`

## Fase 6: UI, CLI y Adapters
- [ ] Configurar CLI root commands con `github.com/spf13/cobra`
- [ ] Implementar TODOS los comandos CLI del PRD:
  - [ ] `fenrir init [--dry-run]`
  - [ ] `fenrir mcp` (servidor stdio)
  - [ ] `fenrir serve [--port 7438]`
  - [ ] `fenrir tui`
  - [ ] `fenrir search <query>`
  - [ ] `fenrir context [--module <path>]`
  - [ ] `fenrir drift [--module <path>]`
  - [ ] `fenrir insights`
  - [ ] `fenrir trace <target>`
  - [ ] `fenrir session list`
  - [ ] `fenrir session show <id>`
  - [ ] `fenrir session audit <id>`
  - [ ] `fenrir arch list`
  - [ ] `fenrir arch add`
  - [ ] `fenrir arch deprecate <id>`
  - [ ] `fenrir pkg check <name> --eco <eco>`
  - [ ] `fenrir pkg audit [--manifest <path>]`
  - [ ] `fenrir sync`
  - [ ] `fenrir sync --import`
  - [ ] `fenrir sync --status`
  - [ ] `fenrir stats`
  - [ ] `fenrir export [--file <path>]`
  - [ ] `fenrir import --file <path>`
  - [ ] `fenrir version`
- [ ] Parseo de configuración de equipo local desde `viper` conectado a `cobra`
- [ ] Construir Dashboard TUI con drift scores, sesiones recientes, insights (RF-08-01)
- [ ] Construir Graph Browser TUI con drill-down de nodos (RF-08-02)
- [ ] Construir Search view TUI con FTS5 full-text (RF-08-03)
- [ ] Construir Session Detail view TUI con DNA completo (RF-08-04)
- [ ] Construir Audit Log view TUI con filtros por risk_level (RF-08-05)
- [ ] Implementar navegación vim-style en TUI: j/k, Enter, Esc, /, q (RF-08-06)
- [ ] Crear paleta de colores propia de Fenrir (RF-08-07)
- [ ] Generar sub-comando `fenrir init` para inferir directorios `.cursor`, `.claude` e introducir `FENRIR.md`

## Fase 7: Mecanismo Git Sync y Distribución Final
- [ ] Crear utilidades de serializar contenido exportado vía GZIP estructurado JSONL
- [ ] Crear rutina para identificar hash addressable de estado para sync idempotente
- [ ] Lógica para manejar comandos `fenrir sync` e import local a través de disco
- [ ] **Implementar lógica de merge idempotente** (último timestamp gana) (RF-07-04)
- [ ] **Resolver conflictos de merge** por timestamp
- [ ] Finalizar `.goreleaser.yaml` para distribución cross-platform
- [ ] Tests End-To-End locales ejecutando CLI y el servidor MCP

## Fase 8: Métricas, Testing y Validación
- [ ] **Implementar métricas de adopción**: contador de instalaciones, sesiones activas
- [ ] **Implementar tracking de crash rate** del servidor MCP
- [ ] **Implementar métricas de impacto**: drift_score promedio, tasa de session_end
- [ ] **Verificar tiempo de startup < 100ms**
- [ ] **Verificar respuesta MCP p99 < 200ms**
- [ ] **Verificar binario < 20MB**
- [ ] **Verificar tiempo de fenrir init < 60 segundos**
