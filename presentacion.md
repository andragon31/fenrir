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

## 6. Capacidades Core (v0.6.0)
| Módulo | Función Principal |
| :--- | :--- |
| **Memoria** | Guardado de descubrimientos, búsqueda semántica y recuperación de contexto. |
| **Gobernanza** | Verificación de acciones contra decisiones de arquitectura pasadas. |
| **Seguridad** | Auditoría de paquetes, chequeo de licencias y detección de typosquatting. |
| **Análisis** | Cálculo de derivas (drift) y predicción de impacto de cambios. |
| **Visualización** | Generación de gráficas Mermaid y Dashboard TUI interactivo. |

---
*Fenrir no es solo una herramienta, es el guardián de la integridad técnica de tu código.*
