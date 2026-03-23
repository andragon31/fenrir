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

## 6. Detalles Técnicos (Deep Dive para Devs)

### 6.1 Gestión de Memoria: El Grafo Semántico
Fenrir no guarda "texto plano", sino que construye un **Grafo de Conocimiento** en una base de datos **SQLite** altamente optimizada:
- **Nodos & Aristas**: Cada descubrimiento es un `Node` (tipo: `observation`, `decision`, `arch`). Las relaciones (`Edge`) conectan estos nodos para entender la jerarquía (ej. "Función A -- depende de --> Librería B").
- **FTS5 (Full Text Search)**: Implementamos búsqueda por texto completo sobre los nodos para que la recuperación de contexto sea instantánea (<10ms), incluso con miles de entradas.
- **DNA de Sesión**: Al finalizar una sesión, Fenrir ejecuta un proceso de **destilación** que resume los cambios (`files_modified`), hallazgos y decisiones en un solo objeto de contexto denso.

### 6.2 Motor de Gobernanza y Verificadores
La arquitectura de gobernanza reside en `internal/graph/arch.go`:
- **Caché de Políticas**: Implementamos un `policyCache` protegido por un `sync.RWMutex`. Esto permite que el agente verifique cada acción contra cientos de decisiones previas sin latencia de disco.
- **Invalidación Inteligente**: El caché se invalida automáticamente solo cuando se detecta un nuevo nodo de tipo `decision` o `policy`, garantizando integridad de datos.

### 6.3 Concurrencia y Eficiencia
Para no ralentizar el ciclo de vida del agente, Fenrir utiliza **Goroutines** para tareas pesadas:
- **Async Drift Calculation**: El cálculo de la deriva arquitectónica se dispara en segundo plano al cerrar una sesión.
- **Async Audit Logging**: Cada llamada a herramienta se registra de forma asíncrona, permitiendo que el agente reciba su respuesta de inmediato mientras Fenrir persiste el log en disco.

### 6.4 Seguridad en el Supply Chain
El módulo `pkg_audit` utiliza heurísticas avanzadas:
- **Levenshtein Distance**: Detecta ataques de **Typosquatting** (ej. si intentas instalar `lodas` en lugar de `lodash`).
- **Age/Downloads Heuristics**: Alerta si un paquete es extremadamente nuevo o tiene muy pocas descargas, indicadores comunes de malware.

### 6.5 Integración MCP (Model Context Protocol)
Fenrir expone sus capacidades como un servidor MCP estandarizado. Esto permite que cualquier IA compatible consuma las herramientas sin necesidad de código personalizado:
- **Server-Side Handlers**: Los handlers en `internal/mcp/server.go` actúan como puente entre la petición JSON-RPC y la lógica del grafo en Go.

## 7. Capacidades Core (v0.6.0)
| Módulo | Función Técnica | Beneficio Dev |
| :--- | :--- | :--- |
| **Memoria** | Almacenamiento semántico y FTS5. | Adiós a la repetición de contexto. |
| **Gobernanza** | Motor de verificación con caché mutex. | Cumplimiento estricto de arquitectura. |
| **Seguridad** | Auditoría heurística de dependencias. | Blindaje contra ataques de cadena de suministro. |
| **Análisis** | Análisis de impacto (BFS) y Drift. | Entender las consecuencias de cada cambio. |
| **TUI** | Interfaz en Bubble Tea (Go). | Monitoreo visual de la salud del proyecto. |

---
*Fenrir es el guardián de la integridad técnica de tu código.*
