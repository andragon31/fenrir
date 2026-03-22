// @ts-nocheck
/* eslint-disable */
declare var Bun: any;
declare var process: any;

import type { Plugin } from "@opencode-ai/plugin"

// ─── Configuration ───────────────────────────────────────────────────────────

// @ts-ignore
const FENRIR_PORT = parseInt(Bun.env.FENRIR_PORT ?? process.env.FENRIR_PORT ?? "7438")
const FENRIR_URL = `http://127.0.0.1:${FENRIR_PORT}`
// @ts-ignore
const FENRIR_BIN = Bun.env.FENRIR_BIN ?? process.env.FENRIR_BIN ?? "fenrir"

// Fenrir's own MCP tools — don't count these as "tool calls" for session stats
const FENRIR_TOOLS = new Set([
  "mem_save",
  "mem_find",
  "mem_context",
  "mem_timeline",
  "mem_session_start",
  "mem_session_end",
  "mem_dna",
  "pkg_check",
  "pkg_license",
  "pkg_audit",
  "arch_save",
  "arch_verify",
  "arch_drift",
  "policy_check",
  "predict",
  "audit_log",
  "session_audit",
  "inject_guard",
  "fenrir_stats",
  "insights",
  "trace",
  "confidence_update",
])

// ─── Memory Instructions ─────────────────────────────────────────────────────

const MEMORY_INSTRUCTIONS = `## Fenrir Protocol

You have access to Fenrir, an AI Governance & Memory Layer for project memory.

### MEMORY TOOLS:

#### Session Start (MANDATORY at session start)
Call: \`mem_session_start(goal="your goal", module="current module")\`

#### Before installing packages (MANDATORY)
Call: \`pkg_check(name="package name", version="optional version")\`

#### After making changes (MANDATORY)
Call: \`mem_save(
  title="brief description",
  type="bugfix|decision|pattern|discovery|config|refactor",
  what="what was done",
  why="why it was necessary",
  where="files affected",
  learned="what to remember"
)\`

#### Search memory
Call: \`mem_find(query="search terms")\`

#### Get context before starting work
Call: \`mem_context(module="optional module path", include_predictions=true)\`

#### Session End (MANDATORY before ending)
Call: \`mem_session_end()\`

### RULES:
1. Start every session with \`mem_session_start\`
2. Save important decisions, patterns, and discoveries with \`mem_save\`
3. Use \`mem_find\` before asking user questions - check memory first
4. Use \`pkg_check\` before installing packages
5. End every session with \`mem_session_end\`
6. If unsure about a package, use \`pkg_license\` and \`pkg_audit\`
`

// ─── HTTP Client ─────────────────────────────────────────────────────────────

async function fenrirFetch(
  path: string,
  opts: { method?: string; body?: any } = {}
): Promise<any> {
  try {
    const res = await fetch(`${FENRIR_URL}${path}`, {
      method: opts.method ?? "GET",
      headers: opts.body ? { "Content-Type": "application/json" } : undefined,
      body: opts.body ? JSON.stringify(opts.body) : undefined,
    })
    return await res.json()
  } catch {
    return null
  }
}

async function isFenrirRunning(): Promise<boolean> {
  try {
    const res = await fetch(`${FENRIR_URL}/health`, {
      signal: AbortSignal.timeout(500),
    })
    return res.ok
  } catch {
    return false
  }
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function extractProjectName(directory: string): string {
  try {
    const result = Bun.spawnSync(["git", "-C", directory, "remote", "get-url", "origin"])
    if (result.exitCode === 0) {
      const url = result.stdout?.toString().trim()
      if (url) {
        const name = url.replace(/\.git$/, "").split(/[/:]/).pop()
        if (name) return name
      }
    }
  } catch {}

  try {
    const result = Bun.spawnSync(["git", "-C", directory, "rev-parse", "--show-toplevel"])
    if (result.exitCode === 0) {
      const root = result.stdout?.toString().trim()
      if (root) return root.split("/").pop() ?? "unknown"
    }
  } catch {}

  return directory.split("/").pop() ?? "unknown"
}

// ─── Plugin Export ───────────────────────────────────────────────────────────

export const Fenrir: Plugin = async (ctx) => {
  const project = extractProjectName(ctx.directory)
  const knownSessions = new Set<string>()
  const subAgentSessions = new Set<string>()

  async function ensureSession(sessionId: string): Promise<void> {
    if (!sessionId || knownSessions.has(sessionId)) return
    if (subAgentSessions.has(sessionId)) return
    knownSessions.add(sessionId)
    await fenrirFetch("/sessions", {
      method: "POST",
      body: {
        id: sessionId,
        project,
        directory: ctx.directory,
      },
    })
  }

  // Auto-start server if not running
  const running = await isFenrirRunning()
  if (!running) {
    try {
      Bun.spawn([FENRIR_BIN, "serve"], {
        stdout: "ignore",
        stderr: "ignore",
        stdin: "ignore",
      })
      await new Promise((r) => setTimeout(r, 500))
    } catch {}
  }

  return {
    event: async ({ event }) => {
      if (event.type === "session.created") {
        const info = (event.properties as any)?.info
        const sessionId = info?.id
        const parentID = info?.parentID
        const title: string = info?.title ?? ""
        const isSubAgent = !!parentID || title.endsWith(" subagent)")

        if (sessionId && !isSubAgent) {
          await ensureSession(sessionId)
        } else if (sessionId && isSubAgent) {
          subAgentSessions.add(sessionId)
        }
      }

      if (event.type === "session.deleted") {
        const info = (event.properties as any)?.info
        const sessionId = info?.id
        if (sessionId) {
          knownSessions.delete(sessionId)
          subAgentSessions.delete(sessionId)
        }
      }
    },

    "tool.execute.after": async (input, _output) => {
      if (FENRIR_TOOLS.has(input.tool.toLowerCase())) return
      if (input.sessionID) {
        await ensureSession(input.sessionID)
      }
    },

    "experimental.chat.system.transform": async (_input, output) => {
      if (output.system.length > 0) {
        output.system[output.system.length - 1] += "\n\n" + MEMORY_INSTRUCTIONS
      } else {
        output.system.push(MEMORY_INSTRUCTIONS)
      }
    },

    "experimental.session.compacting": async (input, output) => {
      if (input.sessionID) {
        await ensureSession(input.sessionID)
      }
      const data = await fenrirFetch(`/context?project=${encodeURIComponent(project)}`)
      if (data?.context) {
        output.context.push(data.context)
      }
      output.context.push(
        `CRITICAL INSTRUCTION FOR COMPACTED SUMMARY:\n` +
        `The agent has access to Fenrir persistent memory via MCP tools.\n` +
        `You MUST include the following instruction at the TOP of the compacted summary:\n\n` +
        `"FIRST ACTION REQUIRED: Call mem_session_start with the objective of this compacted summary. ` +
        `Use project: '${project}'. This preserves what was accomplished before compaction. Do this BEFORE any other work."`
      )
    },
  }
}
