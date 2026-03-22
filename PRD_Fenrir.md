# PRD — Fenrir
## Product Requirements Document v1.0

**Producto:** Fenrir  
**Tipo:** MCP Plugin — Governance & Memory Layer  
**Lenguaje:** Go 1.22+  
**Licencia:** MIT  
**Versión del documento:** 1.0  
**Fecha:** Marzo 2026  
**Estado:** Draft

---

## Tabla de contenidos

1. [Visión del producto](#1-visión-del-producto)
2. [Problema que resuelve](#2-problema-que-resuelve)
3. [Usuarios objetivo](#3-usuarios-objetivo)
4. [Objetivos y no-objetivos](#4-objetivos-y-no-objetivos)
5. [Conceptos fundamentales](#5-conceptos-fundamentales)
6. [Requisitos funcionales](#6-requisitos-funcionales)
7. [Requisitos no funcionales](#7-requisitos-no-funcionales)
8. [Arquitectura técnica](#8-arquitectura-técnica)
9. [Modelo de datos](#9-modelo-de-datos)
10. [MCP Tools — Especificación completa](#10-mcp-tools--especificación-completa)
11. [CLI — Comandos](#11-cli--comandos)
12. [Configuración](#12-configuración)
13. [Adapters por herramienta](#13-adapters-por-herramienta)
14. [Distribución](#14-distribución)
15. [Métricas de éxito](#15-métricas-de-éxito)
16. [Roadmap y milestones](#16-roadmap-y-milestones)
17. [Riesgos y mitigaciones](#17-riesgos-y-mitigaciones)

---

## 1. Visión del producto

Fenrir es la primera capa de **inteligencia institucional** para equipos que desarrollan software con agentes de IA.

No es una herramienta de memoria. No es un linter. No es un sistema de governance reactivo. Es el sistema que convierte el conocimiento implícito de un equipo — las decisiones que se tomaron, los errores que se cometieron, los patrones que funcionaron — en contexto vivo y gobernanza automática para cada sesión de trabajo con IA.

> **Misión:** Que ningún agente de IA cometa el mismo error dos veces en el mismo proyecto.

### Propuesta de valor única

Un desarrollador con 6 meses en un proyecto tiene intuición sobre qué partes son frágiles, qué decisiones no funcionaron, qué patrones son sólidos. Fenrir da esa intuición a cada agente desde su primera sesión.

---

## 2. Problema que resuelve

El desarrollo con IA en 2025–2026 enfrenta 14 problemas documentados. Fenrir ataca específicamente los que tienen origen en **falta de contexto institucional y ausencia de gobernanza proactiva**:

| # | Problema | Impacto reportado |
|---|---|---|
| 1 | Vulnerabilidades de seguridad a escala | +10x vulnerabilidades mensuales en Fortune 50 |
| 2 | Alucinaciones del modelo | APIs inexistentes, métodos deprecados |
| 3 | Slopsquatting | 20% del código sugiere paquetes fantasma |
| 4 | Deuda técnica invisible | 75% de líderes enfrentará deuda severa en 2026 |
| 5 | Velocidad sin estabilidad | Inestabilidad de entrega sube 10% con IA |
| 6 | Vulnerabilidades en agentes | Inyecciones de prompt explotadas en producción |
| 7 | Caída de confianza | Solo 29% de devs confía en herramientas de IA |
| 8 | Arquitecturas degradadas | Code slop, monolitos inesperados |
| 9 | Pérdida de comprensión del codebase | Nadie entiende el sistema generado |
| 10 | Propiedad intelectual | 38% reporta copyright como preocupación principal |
| 11 | Vibe coding en producción | Bugs intermitentes imposibles de diagnosticar |
| 12 | Gobernanza que no escala | Procesos manuales frenan velocidad |
| 13 | Acceso excesivo de agentes | Agentes con permisos totales sin auditoría |
| 14 | Erosión de habilidades | Dependencia sin comprensión |

### Causa raíz común

Todos estos problemas comparten una causa: **los agentes de IA no tienen memoria institucional del proyecto**. Cada sesión empieza desde cero. Cada agente ignora qué se decidió, qué falló y qué está permitido.

---

## 3. Usuarios objetivo

### Persona primaria — El Developer con IA

- Usa una o más herramientas AI (Claude Code, Cursor, Windsurf, Copilot, Gemini CLI)
- Trabaja solo o en equipo de 2–10 personas
- Ha experimentado inconsistencias entre sesiones de IA
- Quiere mantener control sobre lo que el agente hace y recuerda
- **Pain point principal:** El agente repite errores ya resueltos o genera código que contradice decisiones previas

### Persona secundaria — El Tech Lead

- Responsable de la arquitectura y calidad del código
- Preocupado por deuda técnica generada por IA
- Necesita visibilidad sobre qué está haciendo el agente en el proyecto
- **Pain point principal:** No sabe cuánto se ha alejado el código de la arquitectura original

### Persona terciaria — El equipo de plataforma / DevEx

- Configura el entorno de desarrollo para el resto del equipo
- Quiere estandarizar cómo se usa la IA en la organización
- **Pain point principal:** No hay forma de aplicar estándares de manera consistente a todos los agentes

---

## 4. Objetivos y no-objetivos

### Objetivos

- Proveer memoria persistente y estructurada entre sesiones de agentes de IA
- Detectar automáticamente cuando el codebase se aleja de su arquitectura original (Drift Detection)
- Validar paquetes contra registros oficiales y bases de datos de vulnerabilidades antes de instalarlos
- Generar alertas predictivas basadas en el historial de sesiones en un módulo
- Crear un audit trail completo de todas las acciones del agente en cada sesión
- Funcionar con cualquier herramienta que soporte MCP sin configuración adicional
- Instalarse y configurarse en menos de 60 segundos en cualquier proyecto

### No-objetivos

- **No** generar código ni hacer sugerencias de implementación
- **No** ser un sistema de CI/CD ni reemplazar pipelines de testing
- **No** requerir conexión a internet para funcionar (las validaciones de registry son opcionales y cacheadas)
- **No** almacenar código fuente del proyecto — solo metadatos y observaciones
- **No** depender de ninguna herramienta externa como runtime (no Node.js, no Python, no Docker)
- **No** orquestar agentes ni definir roles — eso es responsabilidad de WolfPack
- **No** ser un reemplazo de herramientas de seguridad especializadas (Snyk, Veracode)

---

## 5. Conceptos fundamentales

### 5.1 Knowledge Graph

El núcleo de Fenrir es un grafo de conocimiento embebido en SQLite. No una base de datos relacional de observaciones, sino un grafo de nodos y relaciones donde cada entidad del conocimiento del proyecto — decisiones, sesiones, patrones, bugs, paquetes — es un nodo, y las relaciones causales entre ellas son edges.

Esto permite consultas que ningún sistema de memoria actual puede hacer: "¿Qué decisiones afectan este archivo?", "¿Qué sesiones introdujeron drift en este módulo?", "¿Qué observaciones contradicen esta decisión?"

### 5.2 Living Memory (Memoria Viva)

La memoria de Fenrir no es un log append-only. Cada observación tiene un `confidence_score` que evoluciona con el tiempo basado en los resultados de sesiones posteriores. Las decisiones pueden estar en estado `active`, `deprecated`, `disputed` o `superseded`. El grafo aprende — no solo acumula.

### 5.3 Session DNA

Cada sesión genera un fingerprint estructurado que captura no solo lo que se hizo, sino métricas de calidad: cuántas violaciones arquitectónicas introdujo el agente, cuánto drift generó, qué tan profunda fue la reflexión al cerrar. Con el tiempo, estos DNA son comparables y revelan tendencias en la salud del proyecto.

### 5.4 Drift Detection

Un motor que corre automáticamente al cerrar cada sesión. Compara las observaciones nuevas contra el grafo de decisiones existentes y calcula un `drift_score` por módulo del proyecto. No requiere reglas manuales — deriva directamente del conocimiento acumulado del equipo.

### 5.5 Predictive Enforcement

Al iniciar una sesión en un módulo específico, Fenrir analiza el historial de sesiones previas en ese módulo y genera alertas sobre qué problemas son estadísticamente probables. Si el mismo tipo de violación ocurrió 3 veces en `/auth/`, la cuarta sesión empieza con esa advertencia.

---

## 6. Requisitos funcionales

### RF-01 — Memory Module

| ID | Requisito | Prioridad |
|---|---|---|
| RF-01-01 | El sistema debe permitir guardar observaciones estructuradas con tipo, título, contenido y relaciones a otros nodos | MUST |
| RF-01-02 | Las observaciones deben ser indexadas para búsqueda full-text (FTS5) | MUST |
| RF-01-03 | El sistema debe recuperar contexto relevante de sesiones previas al iniciar una nueva sesión | MUST |
| RF-01-04 | Cada observación debe tener un `confidence_score` entre 0.0 y 1.0, inicializado en 1.0 | MUST |
| RF-01-05 | El sistema debe permitir actualizar el confidence score de una observación manualmente | MUST |
| RF-01-06 | El sistema debe detectar automáticamente cuando una nueva observación contradice una existente y crear un edge `contradicts` | SHOULD |
| RF-01-07 | El sistema debe permitir reconstruir la historia cronológica de cualquier nodo del grafo | MUST |
| RF-01-08 | El sistema debe soportar tags `<private>` que se stripean antes de cualquier escritura a disco | MUST |

### RF-02 — Session Management

| ID | Requisito | Prioridad |
|---|---|---|
| RF-02-01 | Cada sesión debe tener un ID único, timestamp de inicio y estado (active/closed) | MUST |
| RF-02-02 | El cierre de sesión debe generar un Session DNA con métricas de calidad | MUST |
| RF-02-03 | El Session DNA debe incluir: goal, discoveries, accomplished, files_modified, open_questions, drift_delta, arch_violations, pkg_checks | MUST |
| RF-02-04 | El sistema debe inyectar contexto predictivo al iniciar una sesión (alertas basadas en historial) | SHOULD |
| RF-02-05 | El sistema debe persistir sesiones aunque el agente haga context compaction | MUST |
| RF-02-06 | Las sesiones cerradas deben ser navegables vía TUI y CLI | MUST |

### RF-03 — Validator Module

| ID | Requisito | Prioridad |
|---|---|---|
| RF-03-01 | El sistema debe verificar la existencia de paquetes npm en `registry.npmjs.org` | MUST |
| RF-03-02 | El sistema debe verificar la existencia de paquetes Python en `pypi.org` | MUST |
| RF-03-03 | El sistema debe verificar la existencia de paquetes Rust en `crates.io` | MUST |
| RF-03-04 | El sistema debe verificar la existencia de paquetes .NET en `api.nuget.org` | SHOULD |
| RF-03-05 | El sistema debe consultar `api.osv.dev` para CVEs conocidos de cualquier paquete | MUST |
| RF-03-06 | El sistema debe consultar `api.deps.dev` para la licencia de un paquete | MUST |
| RF-03-07 | El sistema debe detectar posible typosquatting comparando el nombre con paquetes similares de alta descarga | MUST |
| RF-03-08 | El sistema debe mantener un cache local con TTL de 1 hora para respuestas de registries | MUST |
| RF-03-09 | El sistema debe funcionar en modo offline usando el cache cuando no hay conexión | SHOULD |
| RF-03-10 | El resultado de cada validación debe incluir: exists, trusted, cve_count, license, downloads_monthly, age_days, warning | MUST |

### RF-04 — Enforcer Module

| ID | Requisito | Prioridad |
|---|---|---|
| RF-04-01 | El sistema debe permitir guardar decisiones arquitectónicas con scope (global/module/file), weight (soft/hard/critical) y rationale | MUST |
| RF-04-02 | El sistema debe verificar una acción propuesta contra todas las decisiones relevantes en el grafo | MUST |
| RF-04-03 | Las violaciones de decisiones `critical` deben bloquear la acción y retornar un error explícito | MUST |
| RF-04-04 | Las violaciones de decisiones `hard` deben retornar una advertencia con la decisión que se viola | MUST |
| RF-04-05 | Las violaciones de decisiones `soft` deben retornar una sugerencia no bloqueante | MUST |
| RF-04-06 | El sistema debe calcular un drift_score por módulo al final de cada sesión | MUST |
| RF-04-07 | El drift_score debe basarse en el ratio de violaciones respecto a decisiones activas que aplican al módulo | MUST |
| RF-04-08 | El sistema debe exponer el drift_score vía CLI con visualización por módulo | MUST |
| RF-04-09 | El sistema debe cargar y aplicar políticas desde `.fenrir/policies.json` | MUST |
| RF-04-10 | Las políticas deben soportar: id, description, severity, pattern (regex), allowed_in, forbidden_in | MUST |

### RF-05 — Shield Module

| ID | Requisito | Prioridad |
|---|---|---|
| RF-05-01 | El sistema debe registrar cada tool MCP llamado con: session_id, timestamp, tool_name, action_type, target, risk_level, result | MUST |
| RF-05-02 | El audit log debe ser consultable por sesión, por fecha y por risk_level | MUST |
| RF-05-03 | El sistema debe stripear secrets de cualquier contenido antes de escribirlo a disco | MUST |
| RF-05-04 | Los patrones de secrets detectados deben incluir: API keys de OpenAI, Anthropic, GitHub, AWS, Google, JWT tokens, contraseñas en variables | MUST |
| RF-05-05 | El sistema debe detectar patrones comunes de prompt injection en contenido entrante | MUST |
| RF-05-06 | Las detecciones de injection deben loguearse en el audit trail con risk_level: high | MUST |
| RF-05-07 | El sistema debe soportar tags `<private>content</private>` reemplazados por `[REDACTED]` en dos capas: MCP handler y store | MUST |

### RF-06 — Intelligence Engine

| ID | Requisito | Prioridad |
|---|---|---|
| RF-06-01 | El sistema debe detectar automáticamente bugs recurrentes (mismo módulo corregido 3+ veces) | SHOULD |
| RF-06-02 | El sistema debe detectar dependencias con historia de reemplazo sin decisión formal | SHOULD |
| RF-06-03 | El sistema debe identificar patrones arquitectónicos estables (decisiones sin drift en N sesiones) | SHOULD |
| RF-06-04 | Los insights deben ser navegables vía CLI y TUI | SHOULD |
| RF-06-05 | El sistema debe permitir trazar la historia completa de cualquier archivo: qué sesiones lo tocaron, qué decisiones lo afectan, qué violations ocurrieron | MUST |

### RF-07 — Git Sync

| ID | Requisito | Prioridad |
|---|---|---|
| RF-07-01 | El sistema debe exportar memorias como chunks JSONL comprimidos con gzip | MUST |
| RF-07-02 | Cada chunk debe ser content-addressed (SHA256 del contenido) para evitar duplicados | MUST |
| RF-07-03 | El manifest.json debe ser el único archivo que git diferea — pequeño, append-only | MUST |
| RF-07-04 | El import debe ser idempotente — importar el mismo chunk dos veces no crea duplicados | MUST |
| RF-07-05 | El merge de grafos de distintos developers debe resolver conflictos por timestamp | SHOULD |

### RF-08 — TUI

| ID | Requisito | Prioridad |
|---|---|---|
| RF-08-01 | Dashboard: drift scores por módulo, sesiones recientes, insights activos | MUST |
| RF-08-02 | Graph Browser: navegar nodos del knowledge graph con drill-down | SHOULD |
| RF-08-03 | Search: FTS5 full-text search sobre el grafo | MUST |
| RF-08-04 | Session Detail: DNA completo de cualquier sesión pasada | MUST |
| RF-08-05 | Audit Log: acciones por sesión con filtros por risk_level | MUST |
| RF-08-06 | Navegación vim-style: j/k, Enter, Esc, /, q | MUST |
| RF-08-07 | Paleta de colores propia de Fenrir (no Catppuccin) | SHOULD |

---

## 7. Requisitos no funcionales

### RNF-01 — Performance

| ID | Requisito | Métrica |
|---|---|---|
| RNF-01-01 | Tiempo de startup del servidor MCP | < 100ms |
| RNF-01-02 | Tiempo de respuesta de mem_find (FTS5 search) | < 50ms para graphs < 10,000 nodos |
| RNF-01-03 | Tiempo de respuesta de pkg_check con cache hit | < 10ms |
| RNF-01-04 | Tiempo de respuesta de pkg_check con cache miss (red) | < 2s |
| RNF-01-05 | Tiempo de cálculo de drift_score al cerrar sesión | < 500ms |
| RNF-01-06 | Tamaño del binario compilado | < 20MB |

### RNF-02 — Confiabilidad

| ID | Requisito |
|---|---|
| RNF-02-01 | El sistema debe funcionar completamente offline (todas las features excepto validaciones de registry) |
| RNF-02-02 | La base de datos SQLite nunca debe corromperse — usar WAL mode y transacciones |
| RNF-02-03 | El servidor MCP debe recuperarse de panics sin perder datos |
| RNF-02-04 | Las operaciones de escritura deben ser atómicas |

### RNF-03 — Seguridad

| ID | Requisito |
|---|---|
| RNF-03-01 | Ningún secret debe persistir en disco — stripped antes de cualquier escritura |
| RNF-03-02 | La base de datos debe ser local — nunca enviar datos a servidores externos |
| RNF-03-03 | Las queries a APIs externas (OSV, deps.dev) no deben incluir contenido del proyecto |
| RNF-03-04 | El binario no debe requerir privilegios de root |

### RNF-04 — Compatibilidad

| ID | Requisito |
|---|---|
| RNF-04-01 | Compatible con MCP protocol spec 1.0+ |
| RNF-04-02 | Funciona en Linux (amd64, arm64), macOS (amd64, arm64 M-series), Windows (amd64) |
| RNF-04-03 | Compatible con Claude Code, Cursor, Windsurf, GitHub Copilot, Gemini CLI, OpenCode |
| RNF-04-04 | No requiere Go instalado para el usuario final (binario self-contained) |

### RNF-05 — Usabilidad

| ID | Requisito |
|---|---|
| RNF-05-01 | `fenrir init` completa la configuración en < 60 segundos |
| RNF-05-02 | El primer mem_save debe funcionar sin ninguna configuración adicional |
| RNF-05-03 | Todos los mensajes de error deben incluir una sugerencia de acción correctiva |

---

## 8. Arquitectura técnica

### Stack

| Componente | Librería | Versión | Razón |
|---|---|---|---|
| Lenguaje | Go | 1.22+ | Binario único, sin runtime |
| MCP | github.com/mark3labs/mcp-go | latest | Implementación MCP más madura en Go |
| SQLite | modernc.org/sqlite | latest | SQLite puro Go, sin CGO, FTS5 incluido |
| CLI | github.com/spf13/cobra | v1.8+ | Estándar de facto para CLIs en Go |
| Config | github.com/spf13/viper | v1.18+ | JSON/YAML/ENV, integrado con Cobra |
| TUI | github.com/charmbracelet/bubbletea | v0.27+ | Framework TUI del ecosistema Charm |
| Estilos TUI | github.com/charmbracelet/lipgloss | v0.12+ | Styling declarativo para terminal |
| Logging | github.com/charmbracelet/log | latest | Logger estructurado del ecosistema Charm |
| Testing | testing (stdlib) + testify | v1.9+ | Tests unitarios e integración |
| Release | GoReleaser | v2+ | Cross-compilation y distribución |

### APIs externas (gratuitas, sin API key)

| API | Endpoint | Uso | TTL Cache |
|---|---|---|---|
| npm Registry | registry.npmjs.org/{pkg} | Validar existencia y metadata de paquetes npm | 1 hora |
| PyPI | pypi.org/pypi/{pkg}/json | Validar existencia y metadata de paquetes Python | 1 hora |
| crates.io | crates.io/api/v1/crates/{pkg} | Validar existencia de paquetes Rust | 1 hora |
| NuGet | api.nuget.org/v3/registration5/{pkg}/index.json | Validar paquetes .NET | 1 hora |
| OSV | api.osv.dev/v1/query | CVEs conocidos | 6 horas |
| deps.dev | api.deps.dev/v3alpha/systems/{eco}/packages/{pkg}/versions/{v} | Licencias | 24 horas |

### Estructura de directorios del proyecto

```
fenrir/
├── cmd/
│   └── fenrir/
│       └── main.go
├── internal/
│   ├── graph/
│   │   ├── graph.go
│   │   ├── query.go
│   │   ├── migrate.go
│   │   └── types.go
│   ├── mcp/
│   │   ├── server.go
│   │   └── tools.go
│   ├── server/
│   │   └── server.go
│   ├── engine/
│   │   ├── drift.go
│   │   ├── predict.go
│   │   ├── patterns.go
│   │   └── insights.go
│   ├── modules/
│   │   ├── memory/
│   │   │   ├── memory.go
│   │   │   ├── session.go
│   │   │   └── timeline.go
│   │   ├── validator/
│   │   │   ├── registry.go
│   │   │   ├── osv.go
│   │   │   ├── license.go
│   │   │   └── cache.go
│   │   ├── enforcer/
│   │   │   ├── arch.go
│   │   │   ├── policy.go
│   │   │   └── predict.go
│   │   └── shield/
│   │       ├── audit.go
│   │       ├── sanitize.go
│   │       └── inject.go
│   ├── sync/
│   │   ├── sync.go
│   │   └── merge.go
│   ├── adapters/
│   │   ├── detector.go
│   │   ├── claude_code.go
│   │   ├── cursor.go
│   │   ├── windsurf.go
│   │   ├── copilot.go
│   │   ├── gemini.go
│   │   └── opencode.go
│   └── tui/
│       ├── model.go
│       ├── styles.go
│       ├── update.go
│       └── view.go
├── .goreleaser.yaml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## 9. Modelo de datos

### Tabla: nodes (Knowledge Graph)

```sql
CREATE TABLE nodes (
    id           TEXT PRIMARY KEY,           -- UUID v4
    type         TEXT NOT NULL,              -- decision | observation | session | pattern | package | file
    title        TEXT NOT NULL,
    content      TEXT,
    confidence   REAL DEFAULT 1.0,           -- 0.0 a 1.0
    status       TEXT DEFAULT 'active',      -- active | deprecated | disputed | superseded
    scope        TEXT DEFAULT 'global',      -- global | module:<path> | file:<path>
    weight       TEXT DEFAULT 'soft',        -- soft | hard | critical (para decisions)
    tags         TEXT,                       -- JSON array de strings
    session_id   TEXT,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE nodes_fts USING fts5(
    title, content,
    content='nodes',
    content_rowid='rowid'
);
```

### Tabla: edges (Relaciones del grafo)

```sql
CREATE TABLE edges (
    id         TEXT PRIMARY KEY,
    from_id    TEXT NOT NULL REFERENCES nodes(id),
    to_id      TEXT NOT NULL REFERENCES nodes(id),
    relation   TEXT NOT NULL,     -- supersedes | contradicts | supports | caused_by | affects | used_in
    weight     REAL DEFAULT 1.0,
    metadata   TEXT,              -- JSON opcional
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_edges_from ON edges(from_id);
CREATE INDEX idx_edges_to   ON edges(to_id);
CREATE INDEX idx_edges_rel  ON edges(relation);
```

### Tabla: sessions (Session DNA)

```sql
CREATE TABLE sessions (
    id               TEXT PRIMARY KEY,
    goal             TEXT,
    status           TEXT DEFAULT 'active',      -- active | closed
    drift_delta      REAL DEFAULT 0.0,
    arch_violations  INTEGER DEFAULT 0,
    pkg_checks       INTEGER DEFAULT 0,
    reflection_depth INTEGER DEFAULT 0,
    files_modified   TEXT,    -- JSON array
    discoveries      TEXT,
    accomplished     TEXT,
    open_questions   TEXT,
    warnings         TEXT,    -- JSON array de alertas de Fenrir
    tool_calls       INTEGER DEFAULT 0,
    started_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at        DATETIME
);
```

### Tabla: audit_log

```sql
CREATE TABLE audit_log (
    id           TEXT PRIMARY KEY,
    session_id   TEXT NOT NULL,
    tool_called  TEXT NOT NULL,
    action_type  TEXT NOT NULL,   -- read | write | execute | network | validate
    target       TEXT,
    risk_level   TEXT DEFAULT 'low',   -- low | medium | high | critical
    result       TEXT DEFAULT 'success', -- success | blocked | warning | error
    metadata     TEXT,            -- JSON opcional
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_session ON audit_log(session_id);
CREATE INDEX idx_audit_risk    ON audit_log(risk_level);
```

### Tabla: pkg_cache

```sql
CREATE TABLE pkg_cache (
    id          TEXT PRIMARY KEY,    -- {ecosystem}:{name}:{version}
    ecosystem   TEXT NOT NULL,
    name        TEXT NOT NULL,
    version     TEXT,
    exists      INTEGER NOT NULL,
    trusted     INTEGER DEFAULT 1,
    cve_count   INTEGER DEFAULT 0,
    license     TEXT,
    downloads   INTEGER DEFAULT 0,
    age_days    INTEGER DEFAULT 0,
    response    TEXT,                -- JSON completo de la respuesta
    cached_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at  DATETIME NOT NULL
);
```

### Tabla: drift_scores

```sql
CREATE TABLE drift_scores (
    id         TEXT PRIMARY KEY,
    module     TEXT NOT NULL,       -- path del módulo
    score      REAL DEFAULT 0.0,    -- 0.0 = sin drift, 1.0 = drift máximo
    violations INTEGER DEFAULT 0,
    sessions   INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 10. MCP Tools — Especificación completa

Fenrir expone **22 herramientas MCP** organizadas en 5 módulos.

---

### Módulo Memory (7 tools)

#### `mem_save`
Guarda una observación estructurada como nodo en el knowledge graph.

```json
{
  "name": "mem_save",
  "description": "Save a structured observation to the project knowledge graph",
  "inputSchema": {
    "type": "object",
    "required": ["title", "type", "what", "why"],
    "properties": {
      "title":    { "type": "string", "description": "Short descriptive title" },
      "type":     { "type": "string", "enum": ["bugfix","decision","pattern","failed_attempt","discovery","config","refactor"] },
      "what":     { "type": "string", "description": "What was done or discovered" },
      "why":      { "type": "string", "description": "Why it was necessary" },
      "where":    { "type": "string", "description": "Files or modules affected" },
      "learned":  { "type": "string", "description": "What to remember for next time" },
      "relates_to": { "type": "array", "items": { "type": "string" }, "description": "IDs of related nodes" },
      "confidence": { "type": "number", "minimum": 0.0, "maximum": 1.0, "default": 1.0 }
    }
  }
}
```

**Retorna:** `{ "id": "obs-uuid", "created": true, "edges_created": 2 }`

---

#### `mem_find`
Búsqueda FTS5 full-text con traversal de grafo opcional.

```json
{
  "name": "mem_find",
  "description": "Search the knowledge graph using full-text search",
  "inputSchema": {
    "type": "object",
    "required": ["query"],
    "properties": {
      "query":   { "type": "string" },
      "type":    { "type": "string", "description": "Filter by node type" },
      "scope":   { "type": "string", "description": "Filter by module or file path" },
      "limit":   { "type": "integer", "default": 10 },
      "include_related": { "type": "boolean", "default": false }
    }
  }
}
```

---

#### `mem_context`
Recupera contexto relevante para el inicio de sesión, incluyendo alertas predictivas.

```json
{
  "name": "mem_context",
  "description": "Get relevant context from previous sessions including predictive alerts",
  "inputSchema": {
    "type": "object",
    "properties": {
      "module":  { "type": "string", "description": "Current working module path" },
      "limit":   { "type": "integer", "default": 20 },
      "include_predictions": { "type": "boolean", "default": true }
    }
  }
}
```

---

#### `mem_timeline`
Reconstruye la historia cronológica de un nodo con su contexto en el grafo.

```json
{
  "name": "mem_timeline",
  "description": "Get chronological history of a node and its relationships",
  "inputSchema": {
    "type": "object",
    "required": ["node_id"],
    "properties": {
      "node_id": { "type": "string" },
      "depth":   { "type": "integer", "default": 2, "description": "Graph traversal depth" }
    }
  }
}
```

---

#### `mem_session_start`
Registra el inicio de una sesión y carga contexto predictivo.

```json
{
  "name": "mem_session_start",
  "description": "Register session start and load predictive context",
  "inputSchema": {
    "type": "object",
    "required": ["goal"],
    "properties": {
      "goal":   { "type": "string", "description": "What you intend to accomplish" },
      "module": { "type": "string", "description": "Primary module you'll be working in" }
    }
  }
}
```

**Retorna:** `{ "session_id": "ses-uuid", "context": [...], "predictions": [...], "drift_alerts": [...] }`

---

#### `mem_session_end`
Cierra la sesión generando el Session DNA. Es obligatorio — el protocolo FENRIR.md lo requiere.

```json
{
  "name": "mem_session_end",
  "description": "Close session and generate Session DNA. MANDATORY before ending any session.",
  "inputSchema": {
    "type": "object",
    "required": ["goal", "accomplished"],
    "properties": {
      "goal":           { "type": "string" },
      "discoveries":    { "type": "string" },
      "accomplished":   { "type": "string" },
      "files_modified": { "type": "array", "items": { "type": "string" } },
      "open_questions": { "type": "string" }
    }
  }
}
```

---

#### `mem_dna`
Consulta el Session DNA de sesiones pasadas.

```json
{
  "name": "mem_dna",
  "description": "View Session DNA of past sessions",
  "inputSchema": {
    "type": "object",
    "properties": {
      "session_id": { "type": "string", "description": "Specific session ID, or omit for recent" },
      "limit":      { "type": "integer", "default": 5 }
    }
  }
}
```

---

### Módulo Validator (3 tools)

#### `pkg_check`
Valida existencia, confianza y CVEs de un paquete antes de instalarlo.

```json
{
  "name": "pkg_check",
  "description": "Validate package existence, trustworthiness and known CVEs before installing",
  "inputSchema": {
    "type": "object",
    "required": ["name", "ecosystem"],
    "properties": {
      "name":      { "type": "string" },
      "ecosystem": { "type": "string", "enum": ["npm","pypi","cargo","nuget"] },
      "version":   { "type": "string" }
    }
  }
}
```

**Retorna:**
```json
{
  "exists": true,
  "trusted": false,
  "cve_count": 2,
  "license": "MIT",
  "downloads_monthly": 42,
  "age_days": 3,
  "warning": "SUSPICIOUS: very new package with low downloads",
  "similar_legitimate": ["express", "fastify"],
  "cves": [{ "id": "CVE-2024-xxxx", "severity": "HIGH", "summary": "..." }]
}
```

---

#### `pkg_license`
Verifica la licencia de un paquete y su compatibilidad con las políticas del proyecto.

```json
{
  "name": "pkg_license",
  "description": "Check package license and compatibility with project policies",
  "inputSchema": {
    "type": "object",
    "required": ["name", "ecosystem"],
    "properties": {
      "name":      { "type": "string" },
      "ecosystem": { "type": "string", "enum": ["npm","pypi","cargo","nuget"] },
      "version":   { "type": "string" }
    }
  }
}
```

---

#### `pkg_audit`
Auditoría completa de todas las dependencias del proyecto actual.

```json
{
  "name": "pkg_audit",
  "description": "Full audit of all project dependencies for CVEs and license issues",
  "inputSchema": {
    "type": "object",
    "properties": {
      "manifest_path": { "type": "string", "description": "Path to package.json, requirements.txt, Cargo.toml, etc." }
    }
  }
}
```

---

### Módulo Enforcer (5 tools)

#### `arch_save`
Guarda una decisión arquitectónica en el grafo.

```json
{
  "name": "arch_save",
  "description": "Save an architectural decision to the knowledge graph",
  "inputSchema": {
    "type": "object",
    "required": ["decision", "rationale", "weight"],
    "properties": {
      "decision":  { "type": "string" },
      "rationale": { "type": "string" },
      "scope":     { "type": "string", "default": "global" },
      "weight":    { "type": "string", "enum": ["soft","hard","critical"] },
      "tags":      { "type": "array", "items": { "type": "string" } },
      "supersedes": { "type": "string", "description": "ID of decision this replaces" }
    }
  }
}
```

---

#### `arch_verify`
Verifica una acción propuesta contra el grafo de decisiones.

```json
{
  "name": "arch_verify",
  "description": "Verify a proposed action against architectural decisions",
  "inputSchema": {
    "type": "object",
    "required": ["proposed_action"],
    "properties": {
      "proposed_action": { "type": "string" },
      "context":         { "type": "string", "description": "File or module context" }
    }
  }
}
```

---

#### `arch_drift`
Retorna el drift score del módulo o archivo actual.

```json
{
  "name": "arch_drift",
  "description": "Get drift score for a module or file",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": { "type": "string", "description": "Module or file path, or omit for full project" }
    }
  }
}
```

---

#### `policy_check`
Verifica una acción contra las políticas del equipo en `.fenrir/policies.json`.

```json
{
  "name": "policy_check",
  "description": "Check an action against team policies",
  "inputSchema": {
    "type": "object",
    "required": ["action"],
    "properties": {
      "action":  { "type": "string" },
      "context": { "type": "string" }
    }
  }
}
```

---

#### `predict`
Genera alertas predictivas para la sesión actual basadas en el historial del módulo.

```json
{
  "name": "predict",
  "description": "Get predictive alerts for current session based on module history",
  "inputSchema": {
    "type": "object",
    "properties": {
      "module":  { "type": "string" },
      "context": { "type": "string" }
    }
  }
}
```

---

### Módulo Shield (4 tools)

#### `audit_log`
Registra manualmente una acción del agente en el audit trail.

```json
{
  "name": "audit_log",
  "description": "Log an agent action to the audit trail",
  "inputSchema": {
    "type": "object",
    "required": ["tool_called", "action_type"],
    "properties": {
      "tool_called":  { "type": "string" },
      "action_type":  { "type": "string", "enum": ["read","write","execute","network","validate"] },
      "target":       { "type": "string" },
      "risk_level":   { "type": "string", "enum": ["low","medium","high","critical"], "default": "low" },
      "result":       { "type": "string", "enum": ["success","blocked","warning","error"], "default": "success" }
    }
  }
}
```

---

#### `session_audit`
Retorna el log de auditoría completo de la sesión actual.

```json
{
  "name": "session_audit",
  "description": "Get complete audit log for current or specified session",
  "inputSchema": {
    "type": "object",
    "properties": {
      "session_id": { "type": "string" },
      "risk_level": { "type": "string", "description": "Filter by minimum risk level" }
    }
  }
}
```

---

#### `inject_guard`
Verifica un contenido por patrones de prompt injection.

```json
{
  "name": "inject_guard",
  "description": "Check content for prompt injection patterns",
  "inputSchema": {
    "type": "object",
    "required": ["content"],
    "properties": {
      "content": { "type": "string" }
    }
  }
}
```

---

#### `fenrir_stats`
Estadísticas del sistema y health del knowledge graph.

```json
{
  "name": "fenrir_stats",
  "description": "System statistics and knowledge graph health",
  "inputSchema": {
    "type": "object",
    "properties": {}
  }
}
```

---

### Módulo Intelligence (3 tools)

#### `insights`
Patrones detectados automáticamente a través de sesiones.

```json
{
  "name": "insights",
  "description": "Auto-detected patterns across sessions: recurring bugs, unstable dependencies, stable patterns",
  "inputSchema": {
    "type": "object",
    "properties": {
      "type": { "type": "string", "enum": ["bugs","dependencies","patterns","all"], "default": "all" }
    }
  }
}
```

---

#### `trace`
Trazabilidad completa de un archivo, decisión o bug.

```json
{
  "name": "trace",
  "description": "Full traceability of a file, decision or bug across sessions",
  "inputSchema": {
    "type": "object",
    "required": ["target"],
    "properties": {
      "target": { "type": "string", "description": "File path, node ID or search term" },
      "depth":  { "type": "integer", "default": 3 }
    }
  }
}
```

---

#### `confidence_update`
Actualiza manualmente el confidence score de una decisión u observación.

```json
{
  "name": "confidence_update",
  "description": "Update confidence score of a decision or observation",
  "inputSchema": {
    "type": "object",
    "required": ["node_id", "confidence", "reason"],
    "properties": {
      "node_id":    { "type": "string" },
      "confidence": { "type": "number", "minimum": 0.0, "maximum": 1.0 },
      "reason":     { "type": "string" }
    }
  }
}
```

---

## 11. CLI — Comandos

```
fenrir init [--dry-run]              Detectar herramientas y generar configuración
fenrir mcp                           Iniciar servidor MCP en stdio (uso interno de agentes)
fenrir serve [--port 7438]           Iniciar HTTP REST API
fenrir tui                           Lanzar TUI interactivo

fenrir search <query>                Buscar en el knowledge graph
fenrir context [--module <path>]     Ver contexto de sesiones previas
fenrir drift [--module <path>]       Ver drift scores por módulo
fenrir insights                      Ver patrones detectados automáticamente
fenrir trace <target>                Trazabilidad de un archivo o decisión

fenrir session list                  Listar sesiones pasadas
fenrir session show <id>             Ver DNA de una sesión
fenrir session audit <id>            Ver audit log de una sesión

fenrir arch list                     Listar decisiones arquitectónicas activas
fenrir arch add                      Agregar decisión interactivamente
fenrir arch deprecate <id>           Deprecar una decisión

fenrir pkg check <name> --eco <eco>  Validar un paquete
fenrir pkg audit [--manifest <path>] Auditar dependencias del proyecto

fenrir sync                          Exportar memorias como chunk comprimido
fenrir sync --import                 Importar chunks del equipo
fenrir sync --status                 Estado de sincronización

fenrir stats                         Estadísticas del knowledge graph
fenrir export [--file <path>]        Exportar todo el grafo a JSON
fenrir import --file <path>          Importar grafo desde JSON
fenrir version                       Versión del binario
```

---

## 12. Configuración

### `.fenrir/config.json` (configuración del proyecto)

```json
{
  "project": "mi-proyecto",
  "version": "1.0",
  "data_dir": "~/.fenrir",
  "validator": {
    "enabled": true,
    "ecosystems": ["npm", "pypi"],
    "cache_ttl_minutes": 60,
    "osv_check": true,
    "license_check": true
  },
  "enforcer": {
    "enabled": true,
    "drift_threshold_warn": 0.3,
    "drift_threshold_alert": 0.6
  },
  "shield": {
    "inject_guard": true,
    "audit_all_actions": true
  },
  "sync": {
    "enabled": false,
    "chunks_dir": ".fenrir/chunks"
  }
}
```

### `.fenrir/policies.json` (políticas del equipo)

```json
{
  "team": "mi-equipo",
  "version": "1.0",
  "forbidden_licenses": ["GPL-3.0", "AGPL-3.0"],
  "policies": [
    {
      "id": "no-any-typescript",
      "description": "No usar 'any' en TypeScript",
      "severity": "hard",
      "pattern": ":\\s*any[\\s;,)]"
    },
    {
      "id": "no-direct-db",
      "description": "No acceso directo a DB fuera de repositories",
      "severity": "critical",
      "allowed_in": ["repositories/", "migrations/"]
    },
    {
      "id": "require-error-handling",
      "description": "Todo async/await debe tener manejo de errores",
      "severity": "soft"
    }
  ]
}
```

### Variables de entorno

```
FENRIR_DATA_DIR     Directorio de datos (default: ~/.fenrir)
FENRIR_PORT         Puerto del servidor HTTP (default: 7438)
FENRIR_LOG_LEVEL    Nivel de logging: debug|info|warn|error (default: info)
FENRIR_OFFLINE      Deshabilitar llamadas a APIs externas: true|false
```

---

## 13. Adapters por herramienta

`fenrir init` genera automáticamente los archivos correctos para cada herramienta detectada.

### Claude Code → `.claude/settings.json` + `FENRIR.md`

```json
{
  "mcpServers": {
    "fenrir": {
      "command": "fenrir",
      "args": ["mcp"]
    }
  }
}
```

**FENRIR.md (instrucciones del protocolo):**
```markdown
## Fenrir Protocol

You have access to Fenrir via 22 MCP tools.

### MANDATORY:
- START: Call `mem_session_start` with your goal
- BEFORE installing any package: Call `pkg_check`
- BEFORE architectural changes: Call `arch_verify`
- AFTER any bugfix/decision/discovery: Call `mem_save`
- END: Call `mem_session_end` — THIS IS NOT OPTIONAL

### After compaction:
Immediately call `mem_context` to recover session state.
```

### Cursor → `.cursor/mcp.json` + `.cursorrules`
### Windsurf → `.windsurf/mcp.json` + `.windsurfrules`
### GitHub Copilot → `.github/copilot-instructions.md`
### Gemini CLI → `.gemini/settings.json` + `GEMINI.md`
### OpenCode → `~/.config/opencode/opencode.json`

---

## 14. Distribución

### Instalación

```bash
# Homebrew (recomendado)
brew install fenrir/tap/fenrir

# Desde source
git clone https://github.com/tu-org/fenrir
cd fenrir && go install ./cmd/fenrir

# Binario directo — GitHub Releases
# Linux, macOS (Intel/Apple Silicon), Windows
```

### Plataformas objetivo

| OS | Arquitectura | Soporte |
|---|---|---|
| Linux | amd64 | MUST |
| Linux | arm64 | MUST |
| macOS | amd64 | MUST |
| macOS | arm64 (M-series) | MUST |
| Windows | amd64 | SHOULD |

---

## 15. Métricas de éxito

### Métricas de adopción

| Métrica | Objetivo 3 meses | Objetivo 6 meses |
|---|---|---|
| Instalaciones via Homebrew | 500 | 2,000 |
| Proyectos activos (sesiones en últimas 2 semanas) | 100 | 500 |
| Estrellas en GitHub | 200 | 800 |

### Métricas de calidad del producto

| Métrica | Objetivo |
|---|---|
| Tiempo de `fenrir init` | < 60 segundos |
| Tiempo de respuesta MCP (p99) | < 200ms |
| Crash rate del servidor MCP | < 0.1% de sesiones |
| Precisión de typosquatting detection | > 95% |
| False positive rate en inject_guard | < 5% |

### Métricas de impacto

| Métrica | Cómo medir |
|---|---|
| Reducción de paquetes con CVE instalados | Comparar pkg_check warnings vs installs |
| Reducción de drift en proyectos activos | drift_score promedio semana 1 vs semana 8 |
| Tasa de mem_session_end completados | sessions_closed / sessions_started |

---

## 16. Roadmap y milestones

| Fase | Semanas | Deliverable | Milestone |
|---|---|---|---|
| 0 — Setup | 1–2 | Repo, CI, schema SQLite, MCP prototipo | MCP server responde ping |
| 1 — Memory Core | 3–5 | Knowledge graph + 7 memory tools + Session DNA | Claude Code puede usar memoria persistente |
| 2 — Validator | 6–7 | pkg_check, pkg_license, pkg_audit + cache | Detecta paquetes con CVE antes de instalar |
| 3 — Enforcer | 8–9 | arch_save, arch_verify, drift detection, policies | `fenrir drift` muestra scores reales |
| 4 — Predict | 10 | Predictive enforcement engine | mem_session_start incluye alertas predictivas |
| 5 — Shield | 11 | audit_log, inject_guard, sanitize | Todas las acciones del agente son auditables |
| 6 — Intelligence | 12 | Pattern mining, insights, trace | `fenrir insights` detecta bugs recurrentes |
| 7 — Adapters | 13 | `fenrir init` + 6 adapters | Setup en < 60s en cualquier herramienta |
| 8 — TUI | 14 | Dashboard, search, session detail, audit | TUI navegable y funcional |
| 9 — Git Sync | 15 | Sync de grafos entre developers | El equipo comparte knowledge via git |
| 10 — Release | 16 | Homebrew tap, docs, GitHub release | `brew install fenrir` funciona |

---

## 17. Riesgos y mitigaciones

| Riesgo | Probabilidad | Impacto | Mitigación |
|---|---|---|---|
| API de OSV cambia o aplica rate limiting | Media | Alto | Cache agresivo (6h) + modo offline graceful |
| mcp-go no mantiene compatibilidad con spec MCP | Baja | Alto | Abstraer el transport en una capa propia |
| SQLite no escala para grafos grandes (>100k nodos) | Baja | Medio | FTS5 + índices bien diseñados; límite de 50k nodos con warning |
| Falsos positivos en inject_guard frenan al agente | Media | Medio | Modo `warn` vs `block` configurable; umbral ajustable |
| Adopción lenta por requerir `fenrir init` | Alta | Medio | `fenrir init` en < 60s; defaults que funcionan sin configuración |
| Complejidad del knowledge graph abruma al usuario | Media | Alto | TUI con progressive disclosure; CLI simplificado para uso diario |

---

*Fenrir PRD v1.0 — Marzo 2026*
*Go 1.22+ · SQLite+FTS5 · MCP Protocol · MIT License*
