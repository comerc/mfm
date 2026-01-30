# Instructions for AI Agents

## KiloCode Agent Config

### Standards & Rules
Strictly follow standards in `./.kilocode/rules/`. Verify against rules before providing code.

### Role Mapping
Match behavior to KiloCode roles based on GitHub Copilot intent:

| Intent / Mode | KiloCode Role | Skill Path |
| :--- | :--- | :--- |
| /plan / Architecture | Architect | ./.kilocode/skills-architect/**/SKILL.md |
| /edit / Code Gen | Code Specialist | ./.kilocode/skills-code/**/SKILL.md |
| Chat / Error Debug | Debug Specialist | ./.kilocode/skills-debug/**/SKILL.md |
| @workspace | Orchestrator | ./.kilocode/skills-orchestrator/**/SKILL.md |

### Execution Protocol
- Pre-read relevant SKILL.md and rules/*.md via @workspace automatically.
- Do not ask for permission to read rules — just use them.

**Note: Rules & Skill files are optional; use them only if they exist.**
