# Fenrir 🐺: AI Governance & Memory Layer
> **"Dándole conciencia y memoria a tu Agente de IA."**

## 1. ¿Qué es Fenrir?
Fenrir es una capa de **Gobernanza e Inteligencia** diseñada para operar entre el desarrollador y el Agente de IA (Claude Code, OpenCode, Cursor, etc.). Actúa como una **memoria de largo plazo** y un **motor de reglas arquitectónicas**, asegurando que la IA no pierda contexto ni tome decisiones que rompan la integridad del proyecto.

## 2. ¿Cómo lo hace? (El Protocolo Fenrir)
Fenrir utiliza el **Model Context Protocol (MCP)** para conectarse directamente con el cerebro de la IA. A través de este canal:
- **Indexación Continua**: Cada descubrimiento, decisión o cambio se guarda en un grafo de conocimiento (SQLite).
- **Verificación en Tiempo Real**: Antes de que la IA instale un paquete o proponga un cambio, Fenrir valida si cumple con las políticas de seguridad y diseño.
- **ADN de Sesión**: Al final de cada sesión de trabajo, Fenrir destila lo aprendido para que la próxima vez que el agente se active, sepa exactamente dónde se quedó.

## 3. ¿Qué Mitiga? (Riesgos que resuelve)
Fenrir es una armadura contra los fallos comunes de la IA:
- **Alucinaciones Arquitectónicas**: Evita que la IA proponga patrones de diseño que ya se decidieron evitar (ej. usar `global state` cuando se prohibió).
- **Drift (Deriva) de Proyecto**: Mitiga que el código pierda coherencia con el tiempo al alertar cuando el agente se desvía de los patrones establecidos.
- **Riesgos de Seguridad (Supply Chain)**: Bloquea paquetes maliciosos, licencias incompatibles o versiones vulnerables mediante auditorías heurísticas (`pkg_audit`).
- **Pérdida de Contexto**: Elimina la necesidad de repetir instrucciones a la IA cada vez que abres un chat nuevo.

## 4. Mejoras de Implementarlo (Beneficios)
- **Eficiencia (>30%)**: Menos tiempo corrigiendo errores de la IA y menos repetición de contexto.
- **Seguridad Garantizada**: Auditoría automática de cada dependencia antes de ser instalada.
- **Visibilidad Total**: Dashboard en TUI y diagramas **Mermaid** generados automáticamente para visualizar la evolución del proyecto.
- **Consistencia de Equipo**: Asegura que diferentes desarrolladores (o diferentes IAs) sigan las mismas reglas de diseño.

## 5. Ecosistema y Compatibilidad
Fenrir es **agnóstico** y funciona en cualquier IDE que soporte agentes MCP:
- **Agentes**: Claude Code, OpenCode, Antigravity, Gemini CLI.
- **IDEs**: Cursor, VS Code, Windsurf, Zed.
- **Llenguajes**: Soporte universal (Go, TS, Python, Rust, etc.).

## 6. Instalación y Despliegue 🚀
Fenrir se instala como un binario único que actúa como servidor MCP.

### 6.1 Instalación Rápida (PowerShell)
```powershell
powershell -ExecutionPolicy Bypass -File .\install.ps1
```
*Este comando descarga/actualiza el binario de Fenrir y lo agrega automáticamente al PATH de tu sistema.*

### 6.2 Configuración en Herramientas (Agents/IDEs)
- **OpenCode**: Ejecuta `fenrir setup opencode`. Esto inyecta automáticamente la configuración en tu `opencode.json`.
- **Claude Code**: La configuración se realiza agregando el binario al archivo de configuración de MCP de Claude (típicamente en `~/.config/claude/mcp.json`).
- **Cursor / VS Code**: Ve a la configuración de "Project Rules" o MCP y agrega un nuevo servidor con el comando `fenrir mcp`.

## 7. Guía de Uso para Desarrolladores 🛠️

### 7.1 Comandos Clave (CLI)
| Comando | Propósito |
| :--- | :--- |
| `fenrir version` | Verifica la versión instalada (v0.6.0). |
| `fenrir stats` | Resumen rápido de nodos, sesiones y auditorías. |
| `fenrir tui` / `main` | Abre el Dashboard interactivo para ver el Drift Analysis. |
| `fenrir setup [tool]` | Autoconfiguración para agentes específicos. |

### 7.2 Cómo Interactuar vía IA
Una vez configurado, no tienes que usar la CLI manualmente. Simplemente habla con tu agente:
- *"Inicia sesión de memoria"*
- *"¿Cómo está la salud arquitectónica del proyecto?"*
- *"Analiza el impacto de refactorizar el módulo X"*

---

## 8. Detalles Técnicos (Deep Dive para Devs)

### 8.1 Gestión de Memoria: El Grafo Semántico
Fenrir construye un **Grafo de Conocimiento** en una base de datos **SQLite** altamente optimizada:
- **Nodos & Aristas**: Cada descubrimiento es un `Node` (tipo: `observation`, `decision`, `arch`). Las relaciones (`Edge`) conectan estos nodos para entender la jerarquía.
- **FTS5 (Full Text Search)**: Búsqueda por texto completo sobre los nodos para recuperación instantánea (<10ms).
- **DNA de Sesión**: Proceso de destilación de contexto al cerrar sesiones.

### 8.2 Motor de Gobernanza y Verificadores
- **Caché de Políticas**: `policyCache` protegido por `sync.RWMutex` para verificaciones en tiempo real sin latencia de disco.
- **Concurrencia**: Uso de **Goroutines** para cálculos de Drift y logs de auditoría sin bloquear la respuesta al agente.

---
*Fenrir es el guardián de la integridad técnica de tu código.*
