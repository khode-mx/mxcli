// SPDX-License-Identifier: Apache-2.0

// tool_templates.go - Templates for multi-tool AI assistant support
package main

import (
	"fmt"
	"path/filepath"
)

// ToolConfig defines configuration for an AI tool
type ToolConfig struct {
	Name        string
	Description string
	Files       []ToolFile
}

// ToolFile defines a configuration file to create
type ToolFile struct {
	Path     string
	Content  func(projectName, mprPath string) string
	Optional bool
}

// SupportedTools defines all AI tools that can be initialized
var SupportedTools = map[string]ToolConfig{
	"claude": {
		Name:        "Claude Code",
		Description: "Claude Code with skills and commands",
		Files: []ToolFile{
			{
				Path:    ".claude/settings.json",
				Content: generateClaudeSettings,
			},
			{
				Path:    "CLAUDE.md",
				Content: generateClaudeMD,
			},
		},
	},
	"cursor": {
		Name:        "Cursor",
		Description: "Cursor AI with MDL rules",
		Files: []ToolFile{
			{
				Path:    ".cursorrules",
				Content: generateCursorRules,
			},
		},
	},
	"continue": {
		Name:        "Continue.dev",
		Description: "Continue.dev with custom commands",
		Files: []ToolFile{
			{
				Path:    ".continue/config.json",
				Content: generateContinueConfig,
			},
		},
	},
	"windsurf": {
		Name:        "Windsurf",
		Description: "Windsurf (Codeium) with MDL rules",
		Files: []ToolFile{
			{
				Path:    ".windsurfrules",
				Content: generateWindsurfRules,
			},
		},
	},
	"aider": {
		Name:        "Aider",
		Description: "Aider with project configuration",
		Files: []ToolFile{
			{
				Path:    ".aider.conf.yml",
				Content: generateAiderConfig,
			},
		},
	},
	"opencode": {
		Name:        "OpenCode",
		Description: "OpenCode AI agent with MDL commands and skills",
		Files: []ToolFile{
			{
				Path:    "opencode.json",
				Content: generateOpenCodeConfig,
			},
		},
	},
	"vibe": {
		Name:        "Mistral Vibe",
		Description: "Mistral Vibe CLI agent with skills",
		Files: []ToolFile{
			{
				Path:    ".vibe/config.toml",
				Content: generateVibeConfig,
			},
			{
				Path:    ".vibe/prompts/mendix-mdl.md",
				Content: generateVibeSystemPrompt,
			},
			{
				Path:    ".vibe/skills/write-microflows/SKILL.md",
				Content: generateVibeSkillWriteMicroflows,
			},
			{
				Path:    ".vibe/skills/create-page/SKILL.md",
				Content: generateVibeSkillCreatePage,
			},
			{
				Path:    ".vibe/skills/check-syntax/SKILL.md",
				Content: generateVibeSkillCheckSyntax,
			},
			{
				Path:    ".vibe/skills/explore-project/SKILL.md",
				Content: generateVibeSkillExploreProject,
			},
		},
	},
}

// Universal files created for all tools
var UniversalFiles = []ToolFile{
	{
		Path:    "AGENTS.md",
		Content: generateProjectAIMD,
	},
}

func generateClaudeSettings(projectName, mprPath string) string {
	return settingsJSON
}

func generateCursorRules(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`# Mendix MDL Project: %s

You are working on a Mendix project with MDL (Mendix Definition Language) support via mxcli.

## Important: mxcli Location

The mxcli tool is in the PROJECT ROOT, not in system PATH. Always use:
- ./mxcli (correct)
- NOT mxcli (will fail)

## Quick Reference

### Project Connection
`+"```bash"+`
./mxcli -p %s -c "SHOW MODULES"
`+"```"+`

### Validate MDL Scripts
`+"```bash"+`
./mxcli check script.mdl                    # Syntax only
./mxcli check script.mdl -p %s --references  # With refs
`+"```"+`

### Execute MDL Scripts
`+"```bash"+`
./mxcli exec script.mdl -p %s
`+"```"+`

### Code Search (requires REFRESH CATALOG FULL)
`+"```bash"+`
./mxcli search -p %s "pattern"
./mxcli callers -p %s Module.Microflow
./mxcli refs -p %s Module.Entity
`+"```"+`

## MDL Syntax Quick Guide

### Microflows
- Variable: `+"`DECLARE $var Type = value;`"+`
- Entity: `+"`DECLARE $entity Module.Entity;`"+` (no AS, no = empty)
- Loop: `+"`LOOP $item IN $list BEGIN ... END LOOP;`"+`
- Change: `+"`CHANGE $obj (Attr = value);`"+`
- If: `+"`IF condition THEN ... END IF;`"+` (not END)
- Log: `+"`LOG WARNING NODE 'Name' 'Message';`"+`

### Pages
- Properties: `+"(Title: 'value', Layout: 'value')"+`
- Widget nesting: curly braces `+"`{ }`"+`
- Widget properties: `+"(Label: 'Name', Attribute: AttrName)"+`

## Documentation

See AGENTS.md for complete documentation and .ai-context/skills/ for patterns.

## Before Writing MDL

1. Read relevant skill file: .ai-context/skills/write-microflows.md or create-page.md
2. Validate: ./mxcli check script.mdl -p %s --references
3. Execute: ./mxcli exec script.mdl -p %s
`, projectName, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile)
}

func generateWindsurfRules(projectName, mprPath string) string {
	// Windsurf uses same format as Cursor
	return generateCursorRules(projectName, mprPath)
}

func generateContinueConfig(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`{
  "name": "%s - Mendix MDL",
  "systemMessage": "You are helping with Mendix development using MDL (Mendix Definition Language). The mxcli tool is located in the project root - always use './mxcli' not 'mxcli'.",
  "docs": [
    "AGENTS.md",
    ".ai-context/skills/"
  ],
  "customCommands": [
    {
      "name": "check-mdl",
      "description": "Check MDL script syntax",
      "prompt": "Run: ./mxcli check {filename}"
    },
    {
      "name": "check-mdl-refs",
      "description": "Check MDL with reference validation",
      "prompt": "Run: ./mxcli check {filename} -p %s --references"
    },
    {
      "name": "execute-mdl",
      "description": "Execute MDL script",
      "prompt": "Run: ./mxcli exec {filename} -p %s"
    },
    {
      "name": "show-entities",
      "description": "Show all entities in project",
      "prompt": "Run: ./mxcli -p %s -c \"SHOW ENTITIES\""
    },
    {
      "name": "search-project",
      "description": "Search project with catalog",
      "prompt": "Run: ./mxcli search -p %s \"{query}\""
    }
  ],
  "slashCommands": [
    {
      "name": "mdl-syntax",
      "description": "Show MDL syntax reference",
      "prompt": "Read and summarize: .ai-context/skills/write-microflows.md"
    },
    {
      "name": "page-syntax",
      "description": "Show page creation syntax",
      "prompt": "Read and summarize: .ai-context/skills/create-page.md"
    }
  ]
}
`, projectName, mprFile, mprFile, mprFile, mprFile)
}

func generateAiderConfig(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`# Mendix MDL Project: %s
# Configuration for Aider AI coding assistant

# Files to read for context
read-files:
  - AGENTS.md
  - .ai-context/skills/*.md

# Project description
description: |
  Mendix project with MDL (Mendix Definition Language) support.
  Use ./mxcli for all project operations.

# Custom commands
commands:
  check: "./mxcli check {file}"
  check-refs: "./mxcli check {file} -p %s --references"
  execute: "./mxcli exec {file} -p %s"
  search: "./mxcli search -p %s {query}"

# Patterns to recognize
recognize:
  - "*.mdl files use MDL syntax (see .ai-context/skills/)"
  - "Always use ./mxcli (local binary) not mxcli"
  - "Microflows: LOOP BEGIN/END LOOP, CHANGE (attr=val)"
  - "Pages: { } blocks, (Prop: value)"
`, projectName, mprFile, mprFile, mprFile)
}

func generateDevcontainerJSON(projectName, mprPath string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "build": {
    "dockerfile": "Dockerfile"
  },
  "features": {
    "ghcr.io/devcontainers/features/docker-in-docker:2": {}
  },
  "forwardPorts": [8080, 8090, 5432],
  "portsAttributes": {
    "8080-8099": { "onAutoForward": "silent" },
    "5432-5499": { "onAutoForward": "silent" }
  },
  "containerEnv": {
    "PLAYWRIGHT_CLI_SESSION": "mendix-app"
  },
  "postCreateCommand": "curl -fsSL https://claude.ai/install.sh | bash && if [ -f ./mxcli ] && ! file ./mxcli | grep -q Linux; then echo '⚠ ./mxcli is not a Linux binary. Replace it with the linux-amd64 or linux-arm64 build.'; fi",
  "customizations": {
    "vscode": {
      "extensions": [
        "anthropic.claude-code"
      ],
      "settings": {
        "mdl.mxcliPath": "./mxcli"
      }
    }
  },
  "remoteUser": "vscode"
}
`, projectName)
}

func generateDockerfile(projectName, mprPath string) string {
	return `FROM mcr.microsoft.com/devcontainers/base:bookworm

# Install Adoptium JDK 21 (required by MxBuild), Node.js 22, and utility tools
RUN apt-get update && apt-get install -y --no-install-recommends wget apt-transport-https gpg ca-certificates curl && \
    wget -qO - https://packages.adoptium.net/artifactory/api/gpg/key/public | gpg --dearmor -o /etc/apt/keyrings/adoptium.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/adoptium.gpg] https://packages.adoptium.net/artifactory/deb bookworm main" > /etc/apt/sources.list.d/adoptium.list && \
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
       temurin-21-jdk \
       nodejs \
       postgresql-client \
       kafkacat \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install playwright-cli and Chromium with all system dependencies (must run as root)
RUN npm install -g @playwright/cli@latest && \
    npx playwright install --with-deps chromium
`
}

func generatePlaywrightConfig() string {
	return `{
  "browser": {
    "browserName": "chromium",
    "isolated": true,
    "launchOptions": {
      "headless": true
    }
  },
  "timeouts": {
    "action": 10000,
    "navigation": 30000
  },
  "network": {
    "allowedOrigins": [
      "http://localhost:8079",
      "http://localhost:8080",
      "http://localhost:8081",
      "http://localhost:8082",
      "http://localhost:8083",
      "http://localhost:8084",
      "http://localhost:8085"
    ]
  }
}
`
}

func generateProjectAIMD(projectName, mprPath string) string {
	return generateClaudeMD(projectName, mprPath)
}

func generateVibeConfig(projectName, mprPath string) string {
	return fmt.Sprintf(`# Mistral Vibe configuration for Mendix project: %s
# See: https://docs.mistral.ai/mistral-vibe/introduction/configuration

# Use the project AGENTS.md as system prompt context
system_prompt_id = "mendix-mdl"

# Skills from .vibe/skills/ are auto-discovered
# Additional context files
# skill_paths = [".ai-context/skills"]

# Tool permissions for MDL workflow
[tools.bash]
permission = "ask"

[tools.read_file]
permission = "always"

[tools.write_file]
permission = "ask"

[tools.search_replace]
permission = "ask"
`, projectName)
}

func generateVibeSystemPrompt(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`You are helping with a Mendix project using MDL (Mendix Definition Language) via mxcli.

## Project: %s

MPR file: %s

## Key Rules

- The mxcli tool is in the project root. Always use ./mxcli, not mxcli.
- Read AGENTS.md for full project documentation.
- Read .ai-context/skills/ for MDL syntax patterns before writing scripts.
- Always validate MDL scripts: ./mxcli check script.mdl -p %s --references
- Microflow variables start with $. Entity declarations have no AS keyword.
- Page widgets nest with curly braces { }. Properties use (Key: value).
- Single quotes in expressions are escaped by doubling: 'it''s here'

## Quick Commands

- Explore: ./mxcli -p %s -c "SHOW STRUCTURE"
- Check: ./mxcli check script.mdl -p %s --references
- Execute: ./mxcli exec script.mdl -p %s
- Search: ./mxcli search -p %s "keyword"
`, projectName, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile)
}

func generateVibeSkillWriteMicroflows(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`---
name: write-microflows
description: Write MDL microflows for Mendix projects
user-invocable: true
allowed-tools:
  - read_file
  - write_file
  - bash
  - grep
---

# Write Microflows

Write MDL microflow scripts for the Mendix project.

## Important

- The mxcli tool is in the project root: `+"`./mxcli`"+`
- MPR file: `+"`%s`"+`
- Always validate before presenting: `+"`./mxcli check script.mdl -p %s --references`"+`

## MDL Microflow Syntax

`+"```sql"+`
CREATE MICROFLOW Module.Name($Param1 Type, $Param2 Module.Entity) RETURNS Boolean
BEGIN
  DECLARE $Var Type = value;
  DECLARE $Entity Module.Entity;

  RETRIEVE $Entity FROM Database WHERE Attr = $Param1 LIMIT 1;

  IF $Entity != empty THEN
    CHANGE $Entity (Attr = 'value');
    COMMIT $Entity;
  END IF;

  RETURN true;
END;
/
`+"```"+`

## Key Rules

1. Variables start with $
2. Entity declarations: no AS keyword, no = empty
3. LOOP $item IN $list BEGIN ... END LOOP
4. IF ... THEN ... END IF (not just END)
5. CHANGE $obj (Attr = value) — parentheses required
6. String escaping: double single quotes 'it''s here'
7. Never create empty list variables as loop sources

## Read .ai-context/skills/write-microflows.md for full reference.
`, mprFile, mprFile)
}

func generateVibeSkillCreatePage(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`---
name: create-page
description: Create MDL pages and widgets for Mendix projects
user-invocable: true
allowed-tools:
  - read_file
  - write_file
  - bash
  - grep
---

# Create Pages

Create MDL page scripts for the Mendix project.

## Important

- mxcli location: `+"`./mxcli`"+`
- MPR file: `+"`%s`"+`
- Validate: `+"`./mxcli check script.mdl -p %s --references`"+`

## MDL Page Syntax

`+"```sql"+`
CREATE PAGE Module.PageName (Title: 'Page Title', Layout: 'Atlas_Default')
{
  DATAVIEW (Entity: Module.Entity, Source: Context) {
    TABLE {
      ROW { TEXTBOX (Label: 'Name', Attribute: Name) }
      ROW { TEXTBOX (Label: 'Email', Attribute: Email) }
    }
    ACTIONBUTTON (Caption: 'Save', Action: SaveChanges)
  }
};
/
`+"```"+`

## Key Rules

1. Widget nesting uses curly braces { }
2. Properties in parentheses (Key: value)
3. String values in single quotes
4. Attribute references without quotes

## Read .ai-context/skills/create-page.md for full reference.
`, mprFile, mprFile)
}

func generateVibeSkillCheckSyntax(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`---
name: check-syntax
description: Validate MDL script syntax and references
user-invocable: true
allowed-tools:
  - bash
  - read_file
---

# Check MDL Syntax

Validate MDL scripts before execution.

## Commands

`+"```bash"+`
# Syntax check only (no project needed)
./mxcli check script.mdl

# Syntax + reference validation
./mxcli check script.mdl -p %s --references

# Execute a script
./mxcli exec script.mdl -p %s
`+"```"+`

## Checklist

1. Run syntax check first
2. Fix any parse errors
3. Run with --references to validate entity/microflow names
4. Execute against the project
`, mprFile, mprFile)
}

func generateVibeSkillExploreProject(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`---
name: explore-project
description: Explore and query Mendix project structure
user-invocable: true
allowed-tools:
  - bash
  - read_file
---

# Explore Project

Query the Mendix project structure using mxcli.

## Common Commands

`+"```bash"+`
# Show project structure
./mxcli -p %s -c "SHOW STRUCTURE"

# List modules, entities, microflows, pages
./mxcli -p %s -c "SHOW MODULES"
./mxcli -p %s -c "SHOW ENTITIES"
./mxcli -p %s -c "SHOW MICROFLOWS"
./mxcli -p %s -c "SHOW PAGES"

# Describe specific elements
./mxcli -p %s -c "DESCRIBE ENTITY Module.EntityName"
./mxcli -p %s -c "DESCRIBE MICROFLOW Module.MicroflowName"

# Show constants and their values per configuration
./mxcli -p %s -c "SHOW CONSTANTS"
./mxcli -p %s -c "SHOW CONSTANT VALUES"

# Search
./mxcli search -p %s "keyword"
`+"```"+`
`, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile)
}

func generateOpenCodeConfig(projectName, mprPath string) string {
	return `{
  "$schema": "https://opencode.ai/config.json",
  "instructions": [
    "AGENTS.md",
    ".opencode/skills/**/SKILL.md",
    ".ai-context/skills/*.md"
  ]
}
`
}
