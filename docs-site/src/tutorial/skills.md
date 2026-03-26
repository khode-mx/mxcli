# Skills and CLAUDE.md

Skills are markdown files that teach AI assistants how to write correct MDL. Each skill covers a specific topic -- creating pages, writing microflows, managing security -- and contains syntax references, examples, and validation checklists. When an AI assistant needs to generate MDL, it reads the relevant skill first, which dramatically improves output quality.

## Where skills live

Skills are installed in two locations:

| Location | Used by |
|----------|---------|
| `.claude/skills/` | Claude Code (tool-specific copy) |
| `.ai-context/skills/` | All tools (universal copy) |

Both directories contain the same files. The `.claude/skills/` copy exists because Claude Code has a built-in mechanism for reading files from its `.claude/` directory. The `.ai-context/skills/` copy is the universal location that any AI tool can access.

## Available skills

`mxcli init` installs the following skill files:

| Skill File | Topic |
|------------|-------|
| `generate-domain-model.md` | Entity, attribute, and association syntax |
| `write-microflows.md` | Microflow syntax, activities, common mistakes |
| `create-page.md` | Page and widget syntax reference |
| `alter-page.md` | ALTER PAGE/SNIPPET for modifying existing pages |
| `overview-pages.md` | CRUD page patterns (overview + edit) |
| `master-detail-pages.md` | Master-detail page patterns |
| `manage-security.md` | Module roles, user roles, access control, GRANT/REVOKE |
| `manage-navigation.md` | Navigation profiles, home pages, menus |
| `demo-data.md` | Mendix ID system, association storage, demo data insertion |
| `xpath-constraints.md` | XPath syntax in WHERE clauses, nested predicates |
| `database-connections.md` | External database connections from microflows |
| `check-syntax.md` | Pre-flight validation checklist |
| `organize-project.md` | Folders, MOVE command, project structure conventions |
| `test-microflows.md` | Test annotations, file formats, Docker setup |
| `patterns-data-processing.md` | Delta merge, batch processing, list operations |

## What a skill file contains

A typical skill file has four sections:

### 1. Syntax reference

The core MDL syntax for the topic, with all available options and keywords:

```sql
-- From write-microflows.md:
CREATE MICROFLOW Module.Name(
    $param: EntityType
) RETURNS ReturnType AS $result
BEGIN
    -- activities here
END;
```

### 2. Examples

Complete, working MDL examples that the AI can use as templates:

```sql
-- From create-page.md:
CREATE PAGE Sales.Customer_Overview
(
  Title: 'Customers',
  Layout: Atlas_Core.Atlas_Default
)
{
  DATAGRID dgCustomers (
    DataSource: DATABASE FROM Sales.Customer SORT BY Name ASC
  ) {
    COLUMN colName (Attribute: Name, Caption: 'Name')
  }
}
```

### 3. Common mistakes

A list of errors the AI should avoid. For example, `write-microflows.md` warns against creating empty list variables as loop sources, and `create-page.md` documents required widget properties that are easy to forget.

### 4. Validation checklist

Steps the AI should follow after writing MDL to confirm correctness:

```bash
# Syntax check (no project needed)
./mxcli check script.mdl

# Syntax + reference validation
./mxcli check script.mdl -p app.mpr --references
```

## CLAUDE.md

The `CLAUDE.md` file is specific to Claude Code. It sits in the project root and is automatically read by Claude when it starts. It provides:

- **Project overview** -- what the project is, which modules exist, and what mxcli is
- **Available commands** -- a summary of mxcli commands Claude can use
- **Rules** -- instructions like "always read the relevant skill file before writing MDL" and "always validate with `mxcli check` before executing"
- **Conventions** -- project-specific naming conventions, module structure, etc.

Think of `CLAUDE.md` as the "system prompt" for Claude Code in the context of your project. It sets the tone and establishes guardrails.

## AGENTS.md

`AGENTS.md` serves the same purpose as `CLAUDE.md` but in a universal format. It is always created by `mxcli init`, regardless of which tool you selected. AI tools that don't have their own config format (or that read markdown files from the project root) will pick up `AGENTS.md` automatically.

## Adding custom skills

You can create your own skill files to teach the AI about your project's patterns and conventions. Add markdown files to `.ai-context/skills/` (or `.claude/skills/` for Claude Code):

```
.ai-context/skills/
├── write-microflows.md           # Built-in (installed by mxcli init)
├── create-page.md                # Built-in
├── our-naming-conventions.md     # Custom: your team's naming rules
├── order-processing-pattern.md   # Custom: how orders work in your app
└── api-integration-guide.md      # Custom: how to call external APIs
```

A custom skill file is just a markdown document. Write it the same way you would explain something to a new team member:

```markdown
# Order Processing Pattern

When creating microflows that process orders in our application, follow these rules:

1. Always validate the order has at least one OrderLine
2. Use the Sales.OrderStatus enumeration for status tracking
3. Log to the 'OrderProcessing' node at INFO level
4. Send a confirmation email via Sales.SendNotification

## Example

\```sql
CREATE MICROFLOW Sales.ACT_Order_Submit($order: Sales.Order)
RETURNS Boolean AS $success
BEGIN
    -- validation, processing, notification...
END;
\```
```

The AI will read your custom skills alongside the built-in ones, learning your project's specific patterns.

## How the AI uses skills

The workflow is straightforward:

1. You ask for a change: "Create a page that shows all orders"
2. The AI determines which skill is relevant (in this case, `create-page.md` and `overview-pages.md`)
3. The AI reads the skill files
4. The AI writes MDL following the syntax and patterns in the skill
5. The AI validates with `mxcli check`
6. The AI executes the script

Skills act as a guardrail. Without them, AI assistants tend to guess at MDL syntax and get details wrong. With skills, the AI has a precise reference to follow, and the output is correct far more often.

## Next steps

Now that you understand how skills guide AI behavior, see [The MDL + AI Workflow](mdl-ai-workflow.md) for the complete recommended workflow from project initialization to review in Studio Pro.
