# Plan de Desarrollo - Fenrir (MCP Plugin)

Este documento detalla el plan paso a paso para desarrollar el producto **Fenrir** cumpliendo con los requisitos del PRD v1.0. 

> [!NOTE]
> **Estado Actual:** Fase 2 finalizada (Módulo Validator core). Iniciando Fase 3 e Integración Universal (Adapters).

---

## Fase 0: Inicialización, Setup y Esqueleto (FINALIZADO)
**Objetivo:** Establecer la base del proyecto y el servidor MCP Stdiod.
1. **Setup del Proyecto Go:** Estructura de directorios, Makefile y go.mod. (Hecho)
2. **Setup de DB (SQLite + FTS5):** Esquema inicial, conexión WAL y FTS5. (Hecho)
3. **Setup del Servidor MCP:** Integración de `mcp-go` y logging con `charmbracelet/log`. (Hecho)

## Fase 1: Motor del Knowledge Graph y Memory Module (FINALIZADO)
**Objetivo:** Registro de observaciones estructuradas y gestión de sesiones.
1. **Data Access Layer (Grafo):** CRUD de `nodes` y `edges`. (Hecho)
2. **Tools Base de Memoria:** `mem_save`, `mem_find`, `mem_timeline`. (Hecho)
3. **Gestión de Sesiones:** `mem_session_start` y `mem_session_end` con Session DNA. (Hecho)

## Fase 2: Módulo Validator (FINALIZADO - Core)
**Objetivo:** Validación de paquetes y seguridad de dependencias.
1. **Caché de Paquetes:** Implementación de `pkg_cache` en SQLite. (Hecho)
2. **Integración con Registries:** Consumo de APIs de npm, PyPI, Cargo y NuGet para existencia. (Hecho)
3. **Seguridad (CVEs):** Integración con **OSV.dev API** para detección de vulnerabilidades. (Hecho)
4. **Heurística:** Detección básica de **Typosquatting**. (Hecho)
5. **CLI:** Comandos `pkg check` y `pkg license`. (Hecho)

## Fase 3: Integración Universal y Adapters (EN PROGRESO)
**Objetivo:** Hacer que Fenrir sea "plug-and-play" en todas las herramientas populares (Cursor, Windsurf, Claude Code, etc.) imitando las mejores prácticas de **Engram**.

1. **Adapter OpenCode:** Plugin TypeScript bridge para tracking de sesiones y compactación. (Hecho)
2. **Adapter Claude Code:** Instalación vía Marketplace (`claude plugin install`) y registro MCP persistente. (Pendiente)
3. **Adapter Gemini CLI:** Registro en `settings.json` e inyección de `system.md`. (Pendiente)
4. **Adapter Universal (Cursor/Windsurf/Antigravity):** Registro en los archivos de configuración JSON de estos IDEs y generación del archivo `FENRIR.md` (Memory Protocol) en la raíz del proyecto. (Pendiente)

## Fase 4: Módulo Enforcer y Policy Engine
**Objetivo:** Regulación del comportamiento del agente y cálculo de Drifts.
1. **Carga de Políticas:** Parser para `.fenrir/policies.json`.
2. **Evaluación de Reglas:** Validaciones Regex (soft/hard/critical).
3. **Drift Detection:** Cálculo de `drift_score` por módulo al finalizar sesiones.
4. **Predictive Alerts:** Generar advertencias basadas en el historial del módulo al iniciar sesión.

## Fase 5: Módulo Shield y Auditoría de Seguridad
**Objetivo:** Sanitización de secrets y protección contra Prompt Injections.
1. **Sanitización:** Regex para remover API Keys y Secrets pre-persistencia.
2. **Stealing Protection:** Soporte para tags `<private>...</private>`.
3. **Audit Trail:** Persistencia de todas las acciones en `audit_log`.
4. **Inject Guard:** Heurísticas para detectar patrones de Prompt Injection.

## Fase 6: Intelligence Engine & Insights
**Objetivo:** Análisis multisesión y grafos de decisión.
1. **Patrones Recurrentes:** Consultas SQL para detectar "hotspots" de bugs.
2. **Visualización de Trazas:** `trace` para navegar descendencias de decisiones.
3. **Insights Automáticos:** Generación de sugerencias proactivas.

## Fase 7: UI (TUI) y Sync
**Objetivo:** Dashboard interactivo y sincronización vía Git.
1. **TUI (Bubbletea):** Dashboard de drift, navegador de grafo y visor de sesiones.
2. **Git Sync:** Exportación por chunks (JSONL Gzipped) content-addressed.
3. **Merge Idempotente:** Resolución de conflictos por timestamp.

## Fase 8: Distribución y Validación Final
1. **Builds:** Configuración final de GoReleaser (Windows, MacOS, Linux).
2. **Benchmarking:** Verificar startup < 100ms y binario < 20MB.
3. **Validación E2E:** Pruebas reales en todos los adapters soportados.
