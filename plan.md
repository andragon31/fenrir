# Plan de Desarrollo - Fenrir (MCP Plugin)

Este documento detalla el plan paso a paso para desarrollar el producto **Fenrir** cumpliendo con los requisitos del PRD v1.0. El plan se divide en fases ordenadas lógicamente, priorizando la creación de la infraestructura principal y luego construyendo las capacidades de cada módulo progresivamente.

## Fase 0: Inicialización, Setup y Esqueleto (Semanas 1-2)

**Objetivo:** Establecer la base del proyecto, las herramientas de desarrollo y la configuración base que permitirá implementar el resto de los módulos. El servidor MCP debe ser capaz de inicializar y responder sin herramientas.

1. **Setup del Proyecto Go:**
   - Inicializar el módulo Go (`go mod init github.com/TU_ORG/fenrir`).
   - Crear la estructura de directorios propuesta en el PRD (`cmd/`, `internal/`, `graph/`, `mcp/`, etc.).
   - Configurar `Makefile` para facilitar la construcción, testing y ejecución local.
2. **Setup de DB (SQLite + FTS5):**
   - Instalar el driver `modernc.org/sqlite` (libre de CGO).
   - Crear los archivos de migración inicial con el esquema de datos documentado (`nodes`, `edges`, `sessions`, `audit_log`, `pkg_cache`, `drift_scores`).
   - Implementar la lógica de conexión (`WAL mode`, PRAGMAs de rendimiento y seguridad atómica).
3. **Setup del Servidor MCP:**
   - Integrar `github.com/mark3labs/mcp-go`.
   - Implementar la estructura base que inicie el servidor via stdio para que los agentes clientes puedan conectarse a él.
   - Crear un helper genérico de logging usando `charmbracelet/log`.

## Fase 1: Motor del Knowledge Graph y Memory Module (Semanas 3-5)

**Objetivo:** Permitir a los agentes registrar observaciones estructuradas, buscar conocimiento con contexto, e iniciar y terminar sesiones generando "Session DNA".

1. **Data Access Layer (Grafo):**
   - Crear las funciones de escritura y lectura sobre la tabla `nodes` y las relaciones causales en la tabla `edges`.
   - Implementar la inserción de nodos con soporte FTS5 para búsquedas full-text.
2. **Tools Base de Memoria (CRUD & Search):**
   - `mem_save`: Lógica de validación, sanitización (stripping `<private>`) y guardado.
   - `mem_find`: Búsqueda usando la tabla virtual FTS5.
   - `mem_timeline`: Navegación iterativa del grafo para obtener la historia cronológica.
3. **Gestión de Sesiones (Living Memory):**
   - `mem_session_start`: Registrar la sesión activa y obtener contexto/alertas predictivas iniciales.
   - `mem_session_end`: Generar y persistir el fingerping (Session DNA), y calcular métricas de calidad.
   - `mem_context` y `mem_dna`: Consultas sobre memoria contextual en base a módulos actuales y la historia del componente.

## Fase 2: Implementación de Módulo Validator (Semanas 6-7)

**Objetivo:** Proveer a los agentes la habilidad de validar paquetes antes de agregarlos como dependencia, integrando APIs gratuitas y con caché robusto.

1. **Gestión de Caché (`pkg_cache`):**
   - Crear abstracción de guardado/lectura en caché SQLite con TTL y estado offline gracefully degradado.
2. **Integración con Registries:**
   - Ecosistemas: Llamadas HTTP a npm, PyPI, crates.io, y api.nuget.org.
   - Implementar detección rudimentaria de "typosquatting" de acuerdo con lógica local.
3. **Integración de Seguridad y Licencias:**
   - Llamadas a `api.osv.dev` para mitigación de CVEs conocidos.
   - Llamadas a `api.deps.dev` para resoluciones de la licencia.
4. **Tools de Validación:**
   - Exponer `pkg_check`, `pkg_license` y `pkg_audit`.

## Fase 3: Módulo Enforcer y Policy Engine (Semanas 8-9)

**Objetivo:** Analizar reglas arquitectónicas y políticas duras pre-definidas para regular el comportamiento del agente y calcular los drifts de los proyectos.

1. **Carga y Verificación de Políticas:**
   - Funciones para cargar `.fenrir/policies.json` en memoria.
   - Evaluar expresiones (RegEx, file matching) para validaciones de código (severity `soft`, `hard`, `critical`).
2. **Knowledge Graph para Arquitectura:**
   - Implementar la lógica para guardar decisiones arquitectónicas e indexarlas.
   - Tool `arch_save` y `arch_verify` (revisar propuesta de un agente contra conocimiento de Fenrir).
3. **Evaluación Continua (Drift Detection):**
   - Motor y background job (o ejecutable al finalizar cada sesión) para el re-cálculo del score de drift.
   - Tools complementarias: `arch_drift`, `policy_check` y `predict`.

## Fase 4: Módulo Shield y Auditoría de Seguridad (Semanas 10-11)

**Objetivo:** Generar un log persistente de auditoría y evitar que secrets sean expuestos o leídos o insertados como conocimiento.

1. **Protección y Sanitización de Inputs:**
   - Implementar regex para remover keys de OpenAI, Anthropic, AWS, Github Tokens y JWT tokens clásicos pre-inserción.
   - Sanitización para soportar la tag `<private>...</private>` a capa de MCP handler.
2. **Registro de logs (Audit Trail):**
   - Insertar logs de operación en tabla `audit_log` con diferentes niveles de risk (risk_level).
3. **Tools Anti-Vulnerabilidad:**
   - Exponer `audit_log` (tool de MCP manual para el agente), `session_audit`, y `inject_guard` para el análisis de prompt injections usando heurística.
4. **System Statistics:**
   - Implementar herramienta `fenrir_stats` para estadísticas del knowledge graph y health del sistema.

## Fase 5: Intelligence Engine & Insights (Semana 12)

**Objetivo:** El AI del sistema en sí mismo. Capacidad estandarizada para analizar patrones entre multi-sesiones.

1. **Búsqueda Autómata de Patrones:**
   - Engine local usando SQLite queries agregadas para inferir qué bugs/ficheros tienen mucha rotación en N sesiones.
2. **Tools Analíticas:**
   - Exponer herramienta `insights` que devuelve resultados formatados agregados.
   - Implementar herramienta de rastreo vertical `trace` usando CTE recursive queries con SQLite en el Grafo.
   - Herramienta `confidence_update` para retroalimentación humana/agente del sistema.

## Fase 6: UI, CLI y Adapters (Semanas 13-14)

**Objetivo:** Desarrollar los puntos de entrada para los usuarios humanos del proyecto y el comando rápido que integra plugins a cualquier IDE u CLI.

1. **Framework del CLI (`cobra` + `viper`):**
   - Montar el binario final en `main.go`.
   - Lógica de configuración `.fenrir/config.json`.
   - Implementar TODOS los comandos documentados en PRD sección 11:
     - Comandos principales: `init`, `mcp`, `serve`, `tui`
     - Comandos de búsqueda: `search`, `context`, `drift`, `insights`, `trace`
     - Comandos de sesión: `session list`, `session show`, `session audit`
     - Comandos de arquitectura: `arch list`, `arch add`, `arch deprecate`
     - Comandos de paquetes: `pkg check`, `pkg audit`
     - Comandos de sync: `sync`, `sync --import`, `sync --status`
     - Comandos utilitarios: `stats`, `export`, `import`, `version`

2. **TUI interactiva (`bubbletea` + `lipgloss`):**
   - Dashboard: drift scores por módulo, sesiones recientes, insights activos (RF-08-01)
   - Graph Browser: navegar nodos del knowledge graph con drill-down (RF-08-02)
   - Search: FTS5 full-text search sobre el grafo (RF-08-03)
   - Session Detail: DNA completo de cualquier sesión pasada (RF-08-04)
   - Audit Log: acciones por sesión con filtros por risk_level (RF-08-05)
   - Navegación vim-style: j/k, Enter, Esc, /, q (RF-08-06)
   - Paleta de colores propia de Fenrir (no Catppuccin) (RF-08-07)

3. **Adapters Auto-Config (init):**
   - El subsistema del comando `fenrir init` capaz de leer qué .folders están (ej: `.cursor`, `.claude`) y escribir automáticamente configuraciones y las INSTRUCCIONES PROMPT (`FENRIR.md`).
   - Generar archivos de configuración para: Claude Code, Cursor, Windsurf, Copilot, Gemini CLI, OpenCode

## Fase 7: Mecanismo de Sincronización y Distribución (Semanas 15-16)

**Objetivo:** Git-first memory support, resolviendo compartición de grafo de forma idempotente entre equipos de desarrolladores y preparando binarios empaquetados.

1. **Motor de Hash e Idempotencia:**
   - Motor de exportación `.fenrir/chunks` (JSONL Gzipped) donde cada nombre de archivo es SHA256 content-addressable.
   - Lógica de la CLI `fenrir sync [--import]` uniendo datos respetando timestamps (último en DB gana).
   - **Merge idempotente**: importar el mismo chunk dos veces no crea duplicados (RF-07-04).
   - **Resolución de conflictos**: merge de grafos de distintos developers por timestamp.
2. **Pipeline de GoReleaser:**
   - Scripting de `.goreleaser.yaml` cubriendo targets de MacOS, Windwos, Linux en todas sus builds y arquetecturas documentadas.
3. **Refinamiento & Documentación:**
   - Aseguramiento de métricas técnicas (start in < 100ms, DB memory leak fixes, sizes <= 20MB).

## Fase 8: Métricas, Testing y Validación (Semanas 17-18)

**Objetivo:** Asegurar que el sistema cumple con las métricas de éxito del PRD y está listo para producción.

1. **Tracking de Métricas de Adopción:**
   - Contador de instalaciones via Homebrew (objetivo: 500 a 3 meses, 2000 a 6 meses)
   - Proyectos activos (sesiones en últimas 2 semanas): objetivo 100 a 3 meses, 500 a 6 meses
   - Estrellas en GitHub: objetivo 200 a 3 meses, 800 a 6 meses

2. **Tracking de Métricas de Calidad:**
   - Tiempo de `fenrir init` < 60 segundos
   - Tiempo de respuesta MCP (p99) < 200ms
   - Crash rate del servidor MCP < 0.1% de sesiones
   - Precisión de typosquatting detection > 95%
   - False positive rate en inject_guard < 5%

3. **Tracking de Métricas de Impacto:**
   - Reducción de paquetes con CVE instalados
   - Reducción de drift en proyectos activos (drift_score promedio semana 1 vs semana 8)
   - Tasa de mem_session_end completados (sessions_closed / sessions_started)

4. **Validación de Performance:**
   - Startup del servidor MCP < 100ms
   - Tiempo de respuesta mem_find (FTS5) < 50ms para graphs < 10,000 nodos
   - Tiempo de respuesta pkg_check con cache hit < 10ms
   - Tiempo de respuesta pkg_check con cache miss < 2s
   - Tamaño del binario compilado < 20MB

5. **Testing Integral:**
   - Tests End-To-End locales ejecutando CLI y el servidor MCP
   - Validación en las 6 herramientas objetivo (Claude Code, Cursor, Windsurf, Copilot, Gemini CLI, OpenCode)

---
*Este plan mapea una relación directa a los milestones descritos en el PRD asegurando la coherencia en el flujo de dependencias técnicas y requerimientos funcionales y no-funcionales.*
