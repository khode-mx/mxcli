# Proposal: Agent Document Type Support in MDL

**Status:** Draft
**Date:** 2026-04-13

## Problem Statement

Mendix 11.9 introduces **Agents** as a first-class concept for building agentic AI applications. Agents are defined in Studio Pro through the Agent Editor extension and stored as `CustomBlobDocuments$CustomBlobDocument` with `CustomDocumentType = "agenteditor.agent"`. The ecosystem includes five marketplace modules:

| Module | Role |
|--------|------|
| **GenAICommons** | Core AI types: Request/Response, DeployedModel, Tool, ToolCall, Trace, KnowledgeBase |
| **MxGenAIConnector** | Mendix Cloud AI backend (Bedrock), model config, embeddings, knowledge base collections |
| **AgentCommons** | Agent management: versioned agents, tools, knowledge bases, MCP, test cases |
| **AgentEditorCommons** | Bridge between Studio Pro Agent Editor extension and AgentCommons runtime entities |
| **MCPClient** | Model Context Protocol client: server connections, tool/prompt discovery, execution |
| **ConversationalUI** | Chat widgets, message rendering, tool approval UI, trace monitoring, token dashboards |

Currently, `mxcli` has no visibility into agent documents. `SHOW STRUCTURE` reports the Agents module as empty because it only contains `CustomBlobDocument` units, which are not parsed. An AI coding agent cannot discover, inspect, or create agents via MDL.

## BSON Structure

Agents use a generic extensibility mechanism:

```
CustomBlobDocuments$CustomBlobDocument:
  $ID: bytes
  $Type: "CustomBlobDocuments$CustomBlobDocument"
  Name: string
  Contents: string (JSON payload)
  CustomDocumentType: "agenteditor.agent"
  Documentation: string
  Excluded: true
  ExportLevel: "Hidden"
  Metadata:
    $ID: bytes
    $Type: "CustomBlobDocuments$CustomBlobDocumentMetadata"
    CreatedByExtension: "extension/agent-editor"
    ReadableTypeName: "Agent"
```

The `Contents` field is a JSON string with this schema:

```json
{
  "description": "",
  "systemPrompt": "Extract {{Information}} from text...",
  "userPrompt": "example input text...",
  "usageType": "Task",
  "variables": [
    { "key": "Information", "isAttributeInEntity": true }
  ],
  "tools": [],
  "knowledgebaseTools": [],
  "entity": {
    "documentId": "<uuid>",
    "qualifiedName": "Module.EntityName"
  }
}
```

Key observations:
- `CustomDocumentType` discriminates agents from other CustomBlobDocuments
- `Contents` is a JSON string (not nested BSON) — the agent editor extension owns this schema
- `Excluded: true` and `ExportLevel: Hidden` are always set
- The `entity` field is optional — links the agent to an entity whose attributes become variables
- `variables` contains template placeholders used in `{{varName}}` syntax within prompts
- `tools` and `knowledgebaseTools` are arrays (empty in all current examples — tools are managed through AgentCommons entities at runtime, not in the document)

### Observed Agent Examples (test3 project)

| Agent | System Prompt | Entity | Variables |
|-------|--------------|--------|-----------|
| InformationExtractorAgent | Extract {{Information}} from text | AgentCommons.InformationExtractor_EXAMPLE | Information |
| SummarizationAgent | Summarize in 3-5 sentences | (none) | (none) |
| TranslationAgent | Translate into {{Description}} | System.Language | Description |
| ProductDescription | Sales assistant for product descriptions | AgentCommons.ProductDescriptionGenerator_EXAMPLE | ProductName, Keywords |

## Proposed MDL Syntax

### SHOW AGENTS

```sql
SHOW AGENTS [IN Module]
```

Output:

| Qualified Name | Module | Name | UsageType | Entity | Variables |
|----------------|--------|------|-----------|--------|-----------|
| AgentEditorCommons.InformationExtractorAgent | AgentEditorCommons | InformationExtractorAgent | Task | AgentCommons.InformationExtractor_EXAMPLE | Information |
| AgentEditorCommons.SummarizationAgent | AgentEditorCommons | SummarizationAgent | Task | | |
| AgentEditorCommons.TranslationAgent | AgentEditorCommons | TranslationAgent | Task | System.Language | Description |
| AgentEditorCommons.ProductDescription | AgentEditorCommons | ProductDescription | Task | AgentCommons.ProductDescriptionGenerator_EXAMPLE | ProductName, Keywords |

### DESCRIBE AGENT

```sql
DESCRIBE AGENT AgentEditorCommons.TranslationAgent
```

Output (round-trippable MDL):

```sql
CREATE AGENT AgentEditorCommons."TranslationAgent" (
  UsageType: Task,
  Entity: System.Language,
  SystemPrompt: 'Translate the given text into {{Description}}.',
  UserPrompt: 'What is a multi-agent AI system?...'
)
VARIABLES (
  "Description" ENTITY ATTRIBUTE
);
/
```

### CREATE AGENT

```sql
CREATE AGENT MyModule."SentimentAnalyzer" (
  UsageType: Task,
  Entity: MyModule.FeedbackItem,
  SystemPrompt: 'Analyze the sentiment of the given {{FeedbackText}}. Classify as positive, negative, or neutral.',
  UserPrompt: 'The product quality exceeded my expectations.'
)
VARIABLES (
  "FeedbackText" ENTITY ATTRIBUTE
);
```

With tools and knowledge bases (future, if the Contents schema supports them inline):

```sql
CREATE AGENT MyModule."ResearchAssistant" (
  UsageType: Conversational,
  SystemPrompt: 'You are a research assistant. Use the provided tools and knowledge base to answer questions.',
  UserPrompt: 'What are the latest trends in renewable energy?'
)
VARIABLES (
  "Topic"
);
```

### DROP AGENT

```sql
DROP AGENT MyModule."SentimentAnalyzer"
```

### Syntax Design Rationale

| Decision | Rationale |
|----------|-----------|
| `AGENT` as document type keyword | Matches `Metadata.ReadableTypeName = "Agent"` and Mendix UI terminology |
| `UsageType: Task` in properties | Follows standard `(Key: value)` property pattern used by all MDL commands |
| `VARIABLES (...)` as separate clause | Variables are structurally distinct from properties; a clause keeps the property block clean |
| `ENTITY ATTRIBUTE` modifier on variables | Distinguishes entity-bound variables from free-form template variables |
| No `TOOLS` or `KNOWLEDGEBASE` clauses yet | In all observed agents, tools/knowledge bases are empty in the document — managed at runtime via AgentCommons. Add when we see populated examples |
| Prompts as string literals | Consistent with other MDL string properties; `{{var}}` placeholders are just text |

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | No | — |
| **BSON parser** | No (CustomBlobDocuments not parsed) | — |
| **Reader/Catalog** | No | — |
| **Grammar** | No | — |
| **AST** | No | — |
| **Visitor** | No | — |
| **Executor** | No | — |
| **Generated metamodel** | Partial | `generated/metamodel/types.go` has `CustomBlobDocuments$CustomBlobDocument` |

## Implementation Plan

### Phase 1: Read Support (SHOW/DESCRIBE)

#### 1.1 Add Model Types

In a new file `sdk/agents/types.go` (or extend an existing domain):

```go
package agents

import "github.com/nicholasgasior/modelsdk-go/model"

type Agent struct {
    ContainerID   model.ID
    ID            model.ID
    Name          string
    Documentation string

    // Parsed from Contents JSON
    Description   string
    SystemPrompt  string
    UserPrompt    string
    UsageType     string       // "Task", "Conversational"
    Variables     []Variable
    Tools         []ToolRef
    KBTools       []KBToolRef
    Entity        *EntityRef   // optional
}

type Variable struct {
    Key                string
    IsAttributeInEntity bool
}

type EntityRef struct {
    DocumentID    string // UUID of the entity's domain model
    QualifiedName string // Module.EntityName
}

type ToolRef struct {
    // TBD — populate when we see non-empty examples
}

type KBToolRef struct {
    // TBD
}
```

#### 1.2 Add BSON Parser

In `sdk/mpr/parser_agent.go`:

- Parse `CustomBlobDocuments$CustomBlobDocument` documents
- Filter by `CustomDocumentType == "agenteditor.agent"`
- Decode `Contents` JSON string into the `Agent` struct
- Store in the reader's document map

The parser should be tolerant: unknown JSON fields in `Contents` are ignored (the agent editor extension may add fields in future versions).

#### 1.3 Add Reader Methods

```go
func (r *Reader) Agents() []*agents.Agent
func (r *Reader) AgentByQualifiedName(name string) *agents.Agent
```

#### 1.4 Add Catalog Table

Add `CATALOG.AGENTS` with columns: `module`, `name`, `qualified_name`, `usage_type`, `entity`, `variables`, `has_tools`, `has_knowledge_base`.

#### 1.5 Add Grammar/AST/Visitor/Executor

- Grammar: `SHOW AGENTS [IN module]`, `DESCRIBE AGENT qualifiedName`
- AST: `ShowAgentsStmt`, `DescribeAgentStmt`
- Executor: format output using standard table/MDL patterns

### Phase 2: Write Support (CREATE/DROP)

#### 2.1 Add BSON Writer

In `sdk/mpr/writer_agent.go`:

- Serialize `Agent` struct to `CustomBlobDocuments$CustomBlobDocument` BSON
- Set `CustomDocumentType = "agenteditor.agent"`
- Set `Metadata` with `CreatedByExtension`, `ReadableTypeName`
- Serialize `Contents` as JSON string
- Set `Excluded = true`, `ExportLevel = "Hidden"`

#### 2.2 Add Grammar/AST/Visitor/Executor for CREATE/DROP

- Grammar: `CREATE AGENT qualifiedName properties variablesClause?`
- AST: `CreateAgentStmt`, `DropAgentStmt`
- Executor: validate, write BSON, register in module

#### 2.3 Validation

- Entity reference must exist (if specified)
- Variables marked `ENTITY ATTRIBUTE` must correspond to attributes on the referenced entity
- `UsageType` must be a known value (`Task` or `Conversational`)
- Variable names used in `{{...}}` in prompts should match declared variables (warning, not error)

### Phase 3: Integration & Catalog

#### 3.1 Catalog Integration

- Include agents in `SHOW STRUCTURE` output
- Add `CATALOG.AGENTS` table for SQL queries
- Include agent references in `SHOW REFERENCES` / `SHOW IMPACT`
- Wire into `REFRESH CATALOG` (both fast and full modes)

#### 3.2 Version Gating

- Agent documents (`CustomBlobDocuments$CustomBlobDocument`) require Mendix 11.x
- Add to `sdk/versions/mendix-11.yaml`:
  ```yaml
  agents:
    agent_document:
      min_version: "11.9.0"
      mdl: "CREATE AGENT Module.Name (...) VARIABLES (...)"
      notes: "Requires AgentEditorCommons marketplace module"
  ```
- Executor pre-check: `checkFeature("agent_document")` before CREATE

#### 3.3 LSP Support

- Hover on agent names shows system prompt summary
- Go-to-definition navigates to agent document
- Completion for `DESCRIBE AGENT` with agent names

### Phase 4: Future Extensions

These are deferred until we have more data on the BSON format:

#### 4.1 ALTER AGENT

```sql
ALTER AGENT MyModule."SentimentAnalyzer"
  SET SystemPrompt = 'New prompt with {{Variable}}.',
  ADD VARIABLE "NewVar" ENTITY ATTRIBUTE,
  DROP VARIABLE "OldVar";
```

#### 4.2 Tools and Knowledge Bases (if stored in document)

```sql
CREATE AGENT MyModule."Assistant" (
  UsageType: Conversational,
  SystemPrompt: '...'
)
TOOLS (
  "SearchTool" MICROFLOW MyModule.SearchAction,
  "Calculator" MICROFLOW MyModule.CalculateAction
)
KNOWLEDGE BASES (
  "ProductDocs" COLLECTION 'product-knowledge' MAX 5 MIN_SIMILARITY 0.7
)
VARIABLES (
  "Query"
);
```

#### 4.3 Consumed MCP Services (if stored in document)

```sql
-- If ConsumedMCPService becomes a document type:
SHOW MCP SERVICES [IN Module]
DESCRIBE MCP SERVICE Module.Name
CREATE MCP SERVICE Module.Name (...)
```

Currently, ConsumedMCPService is a runtime entity managed through MCPClient microflows, not a document type. If it becomes a document type in future Mendix versions, support can be added following the same pattern.

## Building Smart Apps with MDL: End-to-End Examples

This section demonstrates how MDL agent support, combined with the existing agentic marketplace modules (GenAICommons, AgentCommons, MCPClient, ConversationalUI), enables building complete AI-powered applications entirely from MDL scripts.

### Architecture Overview

A "smart app" in Mendix typically has these layers, all expressible in MDL:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Conversational UI                           │
│         Chat widget, tool approval, trace monitoring            │
├─────────────────────────────────────────────────────────────────┤
│                      Agent Layer                                │
│      Agent documents, versions, prompts, variables              │
├──────────────┬────────────────────┬─────────────────────────────┤
│    Tools     │   Knowledge Bases  │      MCP Services           │
│  Microflows  │   RAG retrieval    │  External tool servers      │
├──────────────┴────────────────────┴─────────────────────────────┤
│                    Domain Model                                 │
│           Entities, associations, enumerations                  │
├─────────────────────────────────────────────────────────────────┤
│                     Integrations                                │
│        REST clients, database connectors, OData                 │
└─────────────────────────────────────────────────────────────────┘
```

---

### Example 1: Customer Support Agent with Tools

A conversational agent that helps customer support reps by looking up orders, checking shipment status, and drafting responses. The agent has tools (microflows) it can call autonomously.

#### Step 1: Domain Model

```sql
-- Domain model for the support system

@Position(100, 100)
CREATE PERSISTENT ENTITY Support."Customer" (
  "Name": String(200) NOT NULL ERROR 'Name is required',
  "Email": String(200),
  "Phone": String(50),
  "AccountTier": Enumeration(Support.AccountTier)
);

@Position(350, 100)
CREATE PERSISTENT ENTITY Support."Order" (
  "OrderNumber": String(50) NOT NULL ERROR 'Order number is required',
  "OrderDate": DateTime,
  "TotalAmount": Decimal,
  "Status": Enumeration(Support.OrderStatus)
);

@Position(600, 100)
CREATE PERSISTENT ENTITY Support."SupportTicket" (
  "Subject": String(200),
  "Description": String(unlimited),
  "Priority": Enumeration(Support.TicketPriority),
  "Resolution": String(unlimited),
  "IsResolved": Boolean DEFAULT false
);

CREATE ASSOCIATION Support."Order_Customer"
  FROM Support."Order" TO Support."Customer";

CREATE ASSOCIATION Support."SupportTicket_Customer"
  FROM Support."SupportTicket" TO Support."Customer";

CREATE ASSOCIATION Support."SupportTicket_Order"
  FROM Support."SupportTicket" TO Support."Order";

CREATE ENUMERATION Support."OrderStatus" (
  Pending = 'Pending',
  Shipped = 'Shipped',
  Delivered = 'Delivered',
  Returned = 'Returned'
);

CREATE ENUMERATION Support."TicketPriority" (
  Low = 'Low',
  Medium = 'Medium',
  High = 'High',
  Urgent = 'Urgent'
);

CREATE ENUMERATION Support."AccountTier" (
  Standard = 'Standard',
  Premium = 'Premium',
  Enterprise = 'Enterprise'
);
```

#### Step 2: Tool Microflows

These microflows are what the agent can call as tools. Each tool microflow follows the GenAICommons convention: it receives a `GenAICommons.Request` and `GenAICommons.ToolCall` and returns a `String` result.

```sql
/**
 * Tool: Look up a customer by email address.
 * Returns customer details and account tier.
 *
 * @param $Request The GenAI request context
 * @param $ToolCall Contains the tool input (email address as JSON)
 * @returns String with customer details
 */
CREATE MICROFLOW Support."Tool_LookupCustomer" (
  $Request: GenAICommons.Request,
  $ToolCall: GenAICommons.ToolCall
)
RETURNS String
BEGIN
  DECLARE $Email String = $ToolCall/Input;
  RETRIEVE $Customer FROM DATABASE Support.Customer
    WHERE Email = $Email LIMIT 1;

  IF $Customer != empty THEN
    RETURN 'Customer: ' + $Customer/Name
      + ', Tier: ' + getKey($Customer/AccountTier)
      + ', Email: ' + $Customer/Email
      + ', Phone: ' + $Customer/Phone;
  ELSE
    RETURN 'No customer found with email: ' + $Email;
  END IF;
END;
/

/**
 * Tool: Look up recent orders for a customer.
 * Returns order numbers, dates, amounts, and statuses.
 *
 * @param $Request The GenAI request context
 * @param $ToolCall Contains the tool input (customer name as JSON)
 * @returns String with order summaries
 */
CREATE MICROFLOW Support."Tool_GetOrders" (
  $Request: GenAICommons.Request,
  $ToolCall: GenAICommons.ToolCall
)
RETURNS String
BEGIN
  DECLARE $CustomerName String = $ToolCall/Input;
  RETRIEVE $Customer FROM DATABASE Support.Customer
    WHERE Name = $CustomerName LIMIT 1;

  IF $Customer = empty THEN
    RETURN 'Customer not found: ' + $CustomerName;
  END IF;

  RETRIEVE $OrderList FROM DATABASE Support.Order
    WHERE Support.Order_Customer = $Customer;

  DECLARE $Result String = '';
  LOOP $Order IN $OrderList
  BEGIN
    SET $Result = $Result + 'Order ' + $Order/OrderNumber
      + ' (' + formatDateTime($Order/OrderDate, 'yyyy-MM-dd') + ')'
      + ' - ' + formatDecimal($Order/TotalAmount, 2) + ' EUR'
      + ' - Status: ' + getKey($Order/Status) + '\n';
  END LOOP;

  RETURN if $Result = '' then 'No orders found' else $Result;
END;
/

/**
 * Tool: Create a support ticket for a customer.
 *
 * @param $Request The GenAI request context
 * @param $ToolCall Contains JSON input with subject, description, priority
 * @returns String confirming ticket creation
 */
CREATE MICROFLOW Support."Tool_CreateTicket" (
  $Request: GenAICommons.Request,
  $ToolCall: GenAICommons.ToolCall
)
RETURNS String
BEGIN
  DECLARE $Input String = $ToolCall/Input;
  -- Parse JSON input (simplified — real implementation uses import mapping)

  $Ticket = CREATE Support.SupportTicket (
    Subject = 'Agent-created ticket',
    Description = $Input,
    Priority = Support.TicketPriority.Medium,
    IsResolved = false
  );
  COMMIT $Ticket;

  RETURN 'Ticket created successfully.';
END;
/
```

#### Step 3: Agent Document

```sql
-- The agent definition — stored as a CustomBlobDocument in the project
CREATE AGENT Support."CustomerSupportAgent" (
  UsageType: Conversational,
  SystemPrompt: 'You are a helpful customer support agent for an e-commerce company.

Your capabilities:
- Look up customer information by email
- Check order history and shipment status
- Create support tickets for unresolved issues

Guidelines:
- Always verify the customer identity before sharing order details
- For Premium and Enterprise customers, prioritize their requests
- If you cannot resolve an issue, create a support ticket
- Be empathetic and professional in your responses',
  UserPrompt: 'Hi, I need help with my recent order. My email is jane@example.com.'
);
```

#### Step 4: Runtime Wiring — Action Microflow for Chat

The action microflow connects the ConversationalUI chat widget to the agent. This is the microflow referenced in the ProviderConfig that powers the chat.

```sql
/**
 * Action microflow for the Customer Support chat.
 * Retrieves the agent, builds the request with tools, and calls the LLM.
 *
 * @param $ChatContext The conversation context from the chat widget
 * @returns Boolean indicating success
 */
CREATE MICROFLOW Support."Chat_CustomerSupport" (
  $ChatContext: ConversationalUI.ChatContext
)
RETURNS Boolean
BEGIN
  -- 1. Preprocess: extract user message, build request with chat history
  $Request = CALL MICROFLOW ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) ON ERROR ROLLBACK;

  IF $Request = empty THEN
    RETURN false;
  END IF;

  -- 2. Retrieve the agent and get the prompt
  RETRIEVE $Agent FROM DATABASE AgentCommons.Agent
    WHERE Title = 'CustomerSupportAgent' LIMIT 1;

  $PromptToUse = CALL JAVA ACTION AgentCommons.PromptToUse_GetAndReplace(
    Agent = $Agent, ContextObject = empty
  ) ON ERROR ROLLBACK;

  -- 3. Add agent capabilities (tools, knowledge bases) to the request
  CALL MICROFLOW AgentCommons.Request_AddAgentCapabilities(
    Request = $Request,
    PromptToUse = $PromptToUse
  ) ON ERROR ROLLBACK;

  -- 4. Get the deployed model and call the LLM
  RETRIEVE $DeployedModel FROM $ChatContext
    /ConversationalUI.ChatContext_ProviderConfig_Active
    /ConversationalUI.ProviderConfig
    /ConversationalUI.ProviderConfig_DeployedModel;

  $Response = CALL MICROFLOW GenAICommons.ChatCompletions_WithHistory(
    Request = $Request,
    DeployedModel = $DeployedModel
  ) ON ERROR ROLLBACK;

  -- 5. Update the chat UI with the response
  IF $Response/GenAICommons.Response_Message != empty THEN
    $Message = CALL MICROFLOW ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      Response = $Response
    ) ON ERROR ROLLBACK;
    RETURN true;
  ELSE
    RETURN false;
  END IF;
END;
/
```

#### Step 5: Chat Page with Conversational UI

```sql
/**
 * Customer support chat page using the ConversationalUI widget
 */
CREATE PAGE Support."SupportChat" (
  Title: 'Customer Support',
  Layout: Atlas_Core.Atlas_Default
) {
  HEADER h1 {
    DYNAMICTEXT title (Caption: 'AI Customer Support')
  }
  CONTAINER chatArea (Class: 'chat-fullscreen') {
    -- The ConversationalUI chat widget is provided by the marketplace module.
    -- It renders the chat interface with message history, tool calls,
    -- and user approval for tool execution.
    SNIPPETCALL chatWidget (Snippet: ConversationalUI.Snippet_Output_WithHistory)
  }
};
/
```

#### Step 6: Security

```sql
-- Module roles
CREATE MODULE ROLE Support."User";
CREATE MODULE ROLE Support."Admin";

-- Entity access
GRANT Support.User ON Support.Customer (READ *);
GRANT Support.User ON Support.Order (READ *);
GRANT Support.User ON Support.SupportTicket (CREATE, READ *, WRITE *);
GRANT Support.Admin ON Support.Customer (CREATE, DELETE, READ *, WRITE *);
GRANT Support.Admin ON Support.Order (CREATE, DELETE, READ *, WRITE *);
GRANT Support.Admin ON Support.SupportTicket (CREATE, DELETE, READ *, WRITE *);

-- Microflow access
GRANT EXECUTE ON MICROFLOW Support.Chat_CustomerSupport TO Support.User;
GRANT EXECUTE ON MICROFLOW Support.Tool_LookupCustomer TO Support.User;
GRANT EXECUTE ON MICROFLOW Support.Tool_GetOrders TO Support.User;
GRANT EXECUTE ON MICROFLOW Support.Tool_CreateTicket TO Support.User;

-- Page access
GRANT VIEW ON PAGE Support.SupportChat TO Support.User;
```

---

### Example 2: MCP-Powered Research Agent

An agent that connects to external MCP servers to access tools like web search, file reading, and database queries. This demonstrates the MCPClient module integration.

#### Step 1: Domain Model

```sql
CREATE PERSISTENT ENTITY Research."ResearchProject" (
  "Title": String(200),
  "Objective": String(unlimited),
  "Status": Enumeration(Research.ProjectStatus),
  "Summary": String(unlimited)
);

CREATE PERSISTENT ENTITY Research."Finding" (
  "Content": String(unlimited),
  "Source": String(500),
  "Relevance": Enumeration(Research.Relevance)
);

CREATE ASSOCIATION Research."Finding_Project"
  FROM Research."Finding" TO Research."ResearchProject";

CREATE ENUMERATION Research."ProjectStatus" (
  InProgress = 'In Progress',
  Completed = 'Completed',
  OnHold = 'On Hold'
);

CREATE ENUMERATION Research."Relevance" (
  High = 'High',
  Medium = 'Medium',
  Low = 'Low'
);
```

#### Step 2: MCP Server Connection

The ConsumedMCPService is a runtime entity, created via microflow at startup or by an admin. This microflow sets up the connection to an external MCP server:

```sql
/**
 * Set up connection to an external MCP server that provides
 * web search and file reading tools.
 * Called from an after-startup microflow or admin page.
 */
CREATE MICROFLOW Research."SetupMCPServer" ()
BEGIN
  -- Create the MCP service configuration
  $MCPService = CALL MICROFLOW MCPClient.ConsumedMCPService_Create() ON ERROR ROLLBACK;

  CHANGE $MCPService (
    Name = 'ResearchTools',
    MCPEndpoint = 'https://mcp.example.com/research',
    ProtocolVersion = MCPClient.ENUM_ProtocolVersion.v2025_03_26,
    ConnectionTimeOutInSeconds = 30
  );

  COMMIT $MCPService;

  -- Sync available tools from the MCP server
  CALL MICROFLOW MCPClient.ConsumedMCPService_Check($MCPService) ON ERROR ROLLBACK;

  LOG INFO NODE 'Research' 'MCP server connected: ResearchTools';
END;
/
```

#### Step 3: Agent with MCP Tools

```sql
-- Research agent that uses MCP server tools for web search
CREATE AGENT Research."ResearchAssistant" (
  UsageType: Conversational,
  Entity: Research.ResearchProject,
  SystemPrompt: 'You are a research assistant helping with project: {{Title}}.

Objective: {{Objective}}

Use the available tools to:
1. Search the web for relevant information
2. Read and analyze documents
3. Summarize findings

Always cite your sources. Present findings in a structured format.',
  UserPrompt: 'Find recent developments in quantum computing error correction.'
)
VARIABLES (
  "Title" ENTITY ATTRIBUTE,
  "Objective" ENTITY ATTRIBUTE
);
```

#### Step 4: Action Microflow with MCP Tool Integration

```sql
/**
 * Action microflow for the research chat.
 * Adds MCP server tools to the request so the LLM can call
 * external tools via the MCP protocol.
 *
 * @param $ChatContext The conversation context
 * @returns Boolean indicating success
 */
CREATE MICROFLOW Research."Chat_Research" (
  $ChatContext: ConversationalUI.ChatContext
)
RETURNS Boolean
BEGIN
  $Request = CALL MICROFLOW ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) ON ERROR ROLLBACK;

  IF $Request = empty THEN
    RETURN false;
  END IF;

  -- Add MCP tools from the ResearchTools server
  RETRIEVE $MCPService FROM DATABASE MCPClient.ConsumedMCPService
    WHERE Name = 'ResearchTools' LIMIT 1;

  IF $MCPService != empty THEN
    CALL MICROFLOW MCPClient.Request_AddAllMCPToolsFromServer(
      Request = $Request,
      ConsumedMCPService = $MCPService,
      ENUM_UserAccessApproval = GenAICommons.ENUM_UserAccessApproval.VisibleForUser
    ) ON ERROR ROLLBACK;
  END IF;

  -- Call the LLM with MCP tools available
  RETRIEVE $DeployedModel FROM $ChatContext
    /ConversationalUI.ChatContext_ProviderConfig_Active
    /ConversationalUI.ProviderConfig
    /ConversationalUI.ProviderConfig_DeployedModel;

  $Response = CALL MICROFLOW GenAICommons.ChatCompletions_WithHistory(
    Request = $Request,
    DeployedModel = $DeployedModel
  ) ON ERROR ROLLBACK;

  IF $Response/GenAICommons.Response_Message != empty THEN
    $Message = CALL MICROFLOW ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      Response = $Response
    ) ON ERROR ROLLBACK;
    RETURN true;
  ELSE
    RETURN false;
  END IF;
END;
/
```

---

### Example 3: Single-Call Agent for Data Processing

Not all agents need a chat interface. A "Task" agent processes a single request and returns a result — useful for batch operations, background processing, and microflow-embedded AI.

#### Step 1: Domain Model

```sql
@Position(100, 100)
CREATE PERSISTENT ENTITY Reviews."ProductReview" (
  "ProductName": String(200),
  "ReviewText": String(unlimited),
  "Sentiment": String(50),
  "KeyThemes": String(unlimited),
  "IsProcessed": Boolean DEFAULT false
);
```

#### Step 2: Agent Document

```sql
CREATE AGENT Reviews."SentimentAnalyzer" (
  UsageType: Task,
  Entity: Reviews.ProductReview,
  SystemPrompt: 'Analyze the following product review for {{ProductName}}.

Extract:
1. Overall sentiment (Positive, Negative, Neutral, Mixed)
2. Key themes mentioned (comma-separated)

Respond in this exact format:
Sentiment: <sentiment>
Themes: <theme1>, <theme2>, <theme3>',
  UserPrompt: '{{ReviewText}}'
)
VARIABLES (
  "ProductName" ENTITY ATTRIBUTE,
  "ReviewText" ENTITY ATTRIBUTE
);
```

#### Step 3: Processing Microflow (No Chat UI)

```sql
/**
 * Process a single product review using the SentimentAnalyzer agent.
 * Called from a batch processing microflow or a button action.
 *
 * @param $Review The review to analyze
 */
CREATE MICROFLOW Reviews."ProcessReview" (
  $Review: Reviews.ProductReview
)
BEGIN
  -- Retrieve the agent
  RETRIEVE $Agent FROM DATABASE AgentCommons.Agent
    WHERE Title = 'SentimentAnalyzer' LIMIT 1;

  -- Build a request (single-call, no chat history)
  $Request = CALL MICROFLOW GenAICommons.Request_Create(
    SystemPrompt = empty,
    Temperature = empty,
    MaxTokens = 200,
    TopP = empty,
    ToolChoice = empty
  ) ON ERROR ROLLBACK;

  -- Get prompt with variables replaced by entity attribute values
  $PromptToUse = CALL JAVA ACTION AgentCommons.PromptToUse_GetAndReplace(
    Agent = $Agent,
    ContextObject = $Review
  ) ON ERROR ROLLBACK;

  -- Add the resolved prompts to the request
  CALL MICROFLOW AgentCommons.Request_AddAgentCapabilities(
    Request = $Request,
    PromptToUse = $PromptToUse
  ) ON ERROR ROLLBACK;

  -- Call the LLM (without chat history)
  RETRIEVE $Model FROM DATABASE GenAICommons.DeployedModel
    WHERE IsActive = true LIMIT 1;

  $Response = CALL JAVA ACTION AgentCommons.Agent_Call_WithoutHistory(
    Agent = $Agent,
    ContextObject = $Review,
    Request = $Request,
    FileCollection = empty
  ) ON ERROR ROLLBACK;

  -- Parse the response and update the review
  DECLARE $ResponseText String = $Response/ResponseText;

  CHANGE $Review (
    Sentiment = $ResponseText,
    IsProcessed = true
  );
  COMMIT $Review;
END;
/

/**
 * Batch process all unprocessed reviews.
 *
 * @param $ReviewList The list of reviews to process
 */
CREATE MICROFLOW Reviews."ProcessAllReviews" (
  $ReviewList: List of Reviews.ProductReview
)
BEGIN
  LOOP $Review IN $ReviewList
  BEGIN
    IF $Review/IsProcessed = false THEN
      CALL MICROFLOW Reviews.ProcessReview(Review = $Review) ON ERROR CONTINUE;
    END IF;
  END LOOP;
END;
/
```

---

### Example 4: Multi-Agent System with Tool Approval

A more advanced pattern where the agent's tool calls require user approval before execution. This uses the ConversationalUI tool approval workflow.

#### Step 1: Agent with Sensitive Tools

```sql
CREATE AGENT Finance."ExpenseApprovalAgent" (
  UsageType: Conversational,
  SystemPrompt: 'You are a financial assistant that helps managers review and approve expense reports.

You have access to tools that can:
- Look up expense report details
- Check budget remaining for a department
- Approve or reject expense reports (requires user confirmation)

IMPORTANT: Always show the expense details and budget impact before recommending approval.
Never approve expenses that exceed the remaining department budget.',
  UserPrompt: 'Please review the pending expense reports for the Engineering department.'
);
```

#### Step 2: Tool with User Confirmation

The key is `UserAccessApproval` — when set to `UserConfirmationRequired`, the ConversationalUI shows an approval dialog before the tool executes.

```sql
/**
 * Action microflow for the finance chat.
 * Adds tools with user confirmation requirement for sensitive operations.
 */
CREATE MICROFLOW Finance."Chat_ExpenseApproval" (
  $ChatContext: ConversationalUI.ChatContext
)
RETURNS Boolean
BEGIN
  $Request = CALL MICROFLOW ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) ON ERROR ROLLBACK;

  IF $Request = empty THEN
    RETURN false;
  END IF;

  -- Add a read-only tool (visible but no confirmation needed)
  $ToolCollection = CALL MICROFLOW GenAICommons.Request_GetCreateToolCollection(
    Request = $Request
  ) ON ERROR ROLLBACK;

  CALL JAVA ACTION GenAICommons.Request_AddFunction(
    Request = $Request,
    Name = 'lookup_expenses',
    Description = 'Look up pending expense reports for a department',
    Microflow = 'Finance.Tool_LookupExpenses',
    InputSchema = empty,
    DisplayTitle = 'Look Up Expenses',
    DisplayDescription = 'Search for pending expense reports'
  ) ON ERROR ROLLBACK;

  -- Add a write tool with user confirmation required
  CALL JAVA ACTION GenAICommons.Request_AddFunction(
    Request = $Request,
    Name = 'approve_expense',
    Description = 'Approve an expense report by ID',
    Microflow = 'Finance.Tool_ApproveExpense',
    InputSchema = empty,
    DisplayTitle = 'Approve Expense',
    DisplayDescription = 'Approve the specified expense report'
  ) ON ERROR ROLLBACK;

  -- The ConversationalUI will show an approval dialog for tool calls
  -- marked with UserConfirmationRequired. The user sees the tool name,
  -- description, and input before deciding to approve or reject.

  RETRIEVE $DeployedModel FROM $ChatContext
    /ConversationalUI.ChatContext_ProviderConfig_Active
    /ConversationalUI.ProviderConfig
    /ConversationalUI.ProviderConfig_DeployedModel;

  $Response = CALL MICROFLOW GenAICommons.ChatCompletions_WithHistory(
    Request = $Request,
    DeployedModel = $DeployedModel
  ) ON ERROR ROLLBACK;

  IF $Response/GenAICommons.Response_Message != empty THEN
    $Message = CALL MICROFLOW ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      Response = $Response
    ) ON ERROR ROLLBACK;
    RETURN true;
  ELSE
    RETURN false;
  END IF;
END;
/
```

---

### Example 5: Knowledge Base RAG Agent

An agent that uses a knowledge base (vector store) for Retrieval-Augmented Generation. The agent searches product documentation before answering.

```sql
-- Domain model for the help desk
CREATE PERSISTENT ENTITY HelpDesk."HelpRequest" (
  "Question": String(unlimited),
  "Answer": String(unlimited),
  "Confidence": Decimal
);

-- Agent with knowledge base context
CREATE AGENT HelpDesk."ProductExpert" (
  UsageType: Conversational,
  SystemPrompt: 'You are a product expert for our software platform.

Answer questions using ONLY the information from the provided knowledge base context.
If the knowledge base does not contain relevant information, say so clearly.
Always include the source document reference in your answer.

Do not make up information that is not in the context.',
  UserPrompt: 'How do I configure single sign-on?'
);

/**
 * Action microflow that adds knowledge base retrieval to the request.
 * The GenAICommons framework automatically retrieves relevant chunks
 * from the vector store and includes them in the LLM context.
 */
CREATE MICROFLOW HelpDesk."Chat_ProductExpert" (
  $ChatContext: ConversationalUI.ChatContext
)
RETURNS Boolean
BEGIN
  $Request = CALL MICROFLOW ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) ON ERROR ROLLBACK;

  IF $Request = empty THEN
    RETURN false;
  END IF;

  -- Retrieve the knowledge base configuration
  RETRIEVE $KnowledgeBase FROM DATABASE GenAICommons.DeployedKnowledgeBase
    WHERE Name = 'product-docs' LIMIT 1;

  IF $KnowledgeBase != empty THEN
    -- Add RAG: retrieve top 5 chunks with >= 0.7 similarity
    CALL MICROFLOW GenAICommons.Request_AddKnowledgeBaseRetrieval(
      Request = $Request,
      MaxNumberOfResults = 5,
      MinimumSimilarity = 0.7,
      MetadataCollection = empty,
      Query = empty,
      EmbeddingMicroflow = empty,
      DeployedKnowledgeBase = $KnowledgeBase,
      CollectionIdentifier = 'product-documentation',
      IsConversationAware = true
    ) ON ERROR ROLLBACK;
  END IF;

  RETRIEVE $DeployedModel FROM $ChatContext
    /ConversationalUI.ChatContext_ProviderConfig_Active
    /ConversationalUI.ProviderConfig
    /ConversationalUI.ProviderConfig_DeployedModel;

  $Response = CALL MICROFLOW GenAICommons.ChatCompletions_WithHistory(
    Request = $Request,
    DeployedModel = $DeployedModel
  ) ON ERROR ROLLBACK;

  IF $Response/GenAICommons.Response_Message != empty THEN
    -- Add source references from RAG to the message
    $Message = CALL MICROFLOW ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      Response = $Response
    ) ON ERROR ROLLBACK;

    -- Extract and store sources from the response
    RETRIEVE $References FROM $Response/GenAICommons.Response_Message
      /GenAICommons.Message_Reference;

    LOOP $Ref IN $References
    BEGIN
      CALL MICROFLOW ConversationalUI.Source_Create(
        Message = $Message,
        Title = $Ref/Title,
        Content = $Ref/Content,
        Source = $Ref/Source
      ) ON ERROR CONTINUE;
    END LOOP;

    RETURN true;
  ELSE
    RETURN false;
  END IF;
END;
/
```

---

### Example 6: Building an MCP Server in Mendix

Mendix apps can also act as MCP servers, exposing their capabilities as tools that external AI systems can call. This is done via a Published REST Service that implements the MCP protocol.

```sql
-- Expose Mendix microflows as MCP tools via a Published REST Service.
-- External AI agents (Claude, ChatGPT, custom agents) can discover
-- and call these tools via the MCP protocol.

CREATE PUBLISHED REST SERVICE Support."MCPToolServer" (
  Path: '/mcp/v1',
  Version: '1.0'
)
RESOURCES (
  RESOURCE 'tools' (
    -- List available tools
    OPERATION GET 'list' (
      Microflow: Support.MCP_ListTools,
      Returns: JSON
    ),
    -- Execute a tool call
    OPERATION POST 'call' (
      Microflow: Support.MCP_ExecuteTool,
      Returns: JSON
    )
  )
);

/**
 * MCP endpoint: List available tools.
 * Returns a JSON array of tool definitions that external agents can call.
 */
CREATE MICROFLOW Support."MCP_ListTools" (
  $HttpRequest: System.HttpRequest
)
RETURNS String
BEGIN
  RETURN '{
    "tools": [
      {
        "name": "lookup_customer",
        "description": "Look up customer information by email address",
        "inputSchema": {
          "type": "object",
          "properties": {
            "email": { "type": "string", "description": "Customer email" }
          },
          "required": ["email"]
        }
      },
      {
        "name": "get_order_status",
        "description": "Get the status of an order by order number",
        "inputSchema": {
          "type": "object",
          "properties": {
            "orderNumber": { "type": "string", "description": "Order number" }
          },
          "required": ["orderNumber"]
        }
      }
    ]
  }';
END;
/

/**
 * MCP endpoint: Execute a tool call.
 * Dispatches to the appropriate microflow based on the tool name.
 */
CREATE MICROFLOW Support."MCP_ExecuteTool" (
  $HttpRequest: System.HttpRequest
)
RETURNS String
BEGIN
  DECLARE $Body String = $HttpRequest/Content;
  -- Parse tool name and arguments from the MCP request body
  -- (simplified — real implementation uses JSON structure + import mapping)

  DECLARE $ToolName String = '';
  -- Dispatch based on tool name to the actual implementation microflow
  -- Each tool microflow performs the business logic and returns a result string

  LOG INFO NODE 'MCP' 'Tool call received: ' + $Body;
  RETURN '{"result": "Tool executed successfully"}';
END;
/
```

---

### Example 7: Full Smart App Script — IT Help Desk

This script creates a complete AI-powered IT help desk application in a single MDL file. It demonstrates how all the pieces fit together.

```sql
-- =============================================================
-- IT Help Desk Smart App
-- A complete AI-powered help desk built with MDL
-- =============================================================

-- 1. Module
CREATE MODULE ITHelp;

-- 2. Domain Model
@Position(100, 100)
CREATE PERSISTENT ENTITY ITHelp."Ticket" (
  "Subject": String(200) NOT NULL ERROR 'Subject is required',
  "Description": String(unlimited),
  "Category": Enumeration(ITHelp.Category),
  "Status": Enumeration(ITHelp.TicketStatus),
  "AssignedTo": String(200),
  "Resolution": String(unlimited)
);

@Position(100, 300)
CREATE PERSISTENT ENTITY ITHelp."KBArticle" (
  "Title": String(200),
  "Content": String(unlimited),
  "Category": Enumeration(ITHelp.Category),
  "ViewCount": Integer DEFAULT 0
);

CREATE ENUMERATION ITHelp."Category" (
  Network = 'Network',
  Hardware = 'Hardware',
  Software = 'Software',
  Access = 'Access & Permissions',
  Other = 'Other'
);

CREATE ENUMERATION ITHelp."TicketStatus" (
  New = 'New',
  InProgress = 'In Progress',
  WaitingOnUser = 'Waiting on User',
  Resolved = 'Resolved',
  Closed = 'Closed'
);

-- 3. Agent Definition
CREATE AGENT ITHelp."ITSupportAgent" (
  UsageType: Conversational,
  SystemPrompt: 'You are an IT support agent for a corporate help desk.

You can:
1. Search the knowledge base for solutions to common problems
2. Create support tickets when issues need escalation
3. Check the status of existing tickets
4. Suggest troubleshooting steps

Always try the knowledge base first before creating a ticket.
Be patient and ask clarifying questions when the issue is unclear.
For password resets and access requests, always create a ticket.',
  UserPrompt: 'My laptop screen is flickering and I cannot connect to the VPN.'
);

-- 4. Tool Microflows
/**
 * Tool: Search the knowledge base for relevant articles.
 */
CREATE MICROFLOW ITHelp."Tool_SearchKB" (
  $Request: GenAICommons.Request,
  $ToolCall: GenAICommons.ToolCall
)
RETURNS String
BEGIN
  DECLARE $Query String = $ToolCall/Input;
  RETRIEVE $Articles FROM DATABASE ITHelp.KBArticle
    WHERE contains(Title, $Query) OR contains(Content, $Query);

  DECLARE $Result String = '';
  LOOP $Article IN $Articles
  BEGIN
    SET $Result = $Result + '## ' + $Article/Title + '\n'
      + $Article/Content + '\n\n';
  END LOOP;

  RETURN if $Result = '' then 'No articles found for: ' + $Query else $Result;
END;
/

/**
 * Tool: Create a new support ticket.
 */
CREATE MICROFLOW ITHelp."Tool_CreateTicket" (
  $Request: GenAICommons.Request,
  $ToolCall: GenAICommons.ToolCall
)
RETURNS String
BEGIN
  $Ticket = CREATE ITHelp.Ticket (
    Subject = 'Auto-created: ' + $ToolCall/Input,
    Description = $ToolCall/Input,
    Status = ITHelp.TicketStatus.New,
    Category = ITHelp.Category.Other
  );
  COMMIT $Ticket;
  RETURN 'Ticket created. A technician will be assigned shortly.';
END;
/

-- 5. Chat Action Microflow
/**
 * Main action microflow powering the IT help desk chat.
 */
CREATE MICROFLOW ITHelp."Chat_ITSupport" (
  $ChatContext: ConversationalUI.ChatContext
)
RETURNS Boolean
BEGIN
  $Request = CALL MICROFLOW ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) ON ERROR ROLLBACK;

  IF $Request = empty THEN
    RETURN false;
  END IF;

  -- Retrieve agent and add its tools/prompts
  RETRIEVE $Agent FROM DATABASE AgentCommons.Agent
    WHERE Title = 'ITSupportAgent' LIMIT 1;

  $PromptToUse = CALL JAVA ACTION AgentCommons.PromptToUse_GetAndReplace(
    Agent = $Agent, ContextObject = empty
  ) ON ERROR ROLLBACK;

  CALL MICROFLOW AgentCommons.Request_AddAgentCapabilities(
    Request = $Request, PromptToUse = $PromptToUse
  ) ON ERROR ROLLBACK;

  RETRIEVE $DeployedModel FROM $ChatContext
    /ConversationalUI.ChatContext_ProviderConfig_Active
    /ConversationalUI.ProviderConfig
    /ConversationalUI.ProviderConfig_DeployedModel;

  $Response = CALL MICROFLOW GenAICommons.ChatCompletions_WithHistory(
    Request = $Request, DeployedModel = $DeployedModel
  ) ON ERROR ROLLBACK;

  IF $Response/GenAICommons.Response_Message != empty THEN
    $Message = CALL MICROFLOW ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      Response = $Response
    ) ON ERROR ROLLBACK;
    RETURN true;
  ELSE
    RETURN false;
  END IF;
END;
/

-- 6. Pages
CREATE PAGE ITHelp."HelpDesk" (
  Title: 'IT Help Desk',
  Layout: Atlas_Core.Atlas_Default
) {
  HEADER h1 {
    DYNAMICTEXT title (Caption: 'IT Help Desk')
  }
  LAYOUTGRID grid {
    ROW row1 {
      COLUMN col1 (Weight: 12) {
        CONTAINER chatContainer {
          SNIPPETCALL chat (Snippet: ConversationalUI.Snippet_Output_WithHistory)
        }
      }
    }
  }
};
/

CREATE PAGE ITHelp."TicketOverview" (
  Title: 'Support Tickets',
  Layout: Atlas_Core.Atlas_Default
) {
  HEADER h1 {
    DYNAMICTEXT title (Caption: 'Support Tickets')
  }
  DATAGRID ticketGrid (DataSource: DATABASE ITHelp.Ticket) {
    COLUMN col1 (Attribute: Subject, Caption: 'Subject')
    COLUMN col2 (Attribute: Category, Caption: 'Category')
    COLUMN col3 (Attribute: Status, Caption: 'Status')
    COLUMN col4 (Attribute: AssignedTo, Caption: 'Assigned To')
  }
};
/

-- 7. Security
CREATE MODULE ROLE ITHelp."User";
CREATE MODULE ROLE ITHelp."Admin";

GRANT ITHelp.User ON ITHelp.Ticket (CREATE, READ *, WRITE (ITHelp.Ticket.Description));
GRANT ITHelp.User ON ITHelp.KBArticle (READ *);
GRANT ITHelp.Admin ON ITHelp.Ticket (CREATE, DELETE, READ *, WRITE *);
GRANT ITHelp.Admin ON ITHelp.KBArticle (CREATE, DELETE, READ *, WRITE *);

GRANT EXECUTE ON MICROFLOW ITHelp.Chat_ITSupport TO ITHelp.User;
GRANT EXECUTE ON MICROFLOW ITHelp.Tool_SearchKB TO ITHelp.User;
GRANT EXECUTE ON MICROFLOW ITHelp.Tool_CreateTicket TO ITHelp.User;
GRANT VIEW ON PAGE ITHelp.HelpDesk TO ITHelp.User;
GRANT VIEW ON PAGE ITHelp.TicketOverview TO ITHelp.User, ITHelp.Admin;

-- 8. Navigation
CREATE OR REPLACE NAVIGATION Responsive_web
  HOME PAGE ITHelp.HelpDesk FOR ITHelp.User
  MENU (
    ITEM 'Help Desk' PAGE ITHelp.HelpDesk,
    ITEM 'Tickets' PAGE ITHelp.TicketOverview,
    ITEM 'Agent Admin' PAGE AgentCommons.Agent_Overview
  );
```

---

### Summary: What MDL Agent Support Enables

| Capability | Without MDL Agent Support | With MDL Agent Support |
|------------|---------------------------|------------------------|
| **Discover agents** | Open Studio Pro, navigate to Agent Editor | `SHOW AGENTS` in CLI or script |
| **Inspect agent prompts** | Click through Agent Editor UI | `DESCRIBE AGENT Module.Name` |
| **Create agents** | Only via Studio Pro Agent Editor | `CREATE AGENT` in MDL scripts |
| **Version control** | Binary CustomBlobDocument diffs | Human-readable MDL diffs |
| **AI-assisted development** | AI cannot see or create agents | AI generates complete smart apps |
| **Batch operations** | Manual, one agent at a time | Script creates multiple agents |
| **Code review** | Cannot review agent changes in PR | MDL changes are reviewable text |
| **Migration** | Manual recreation in new project | Copy/paste MDL scripts |
| **Documentation** | Screenshots of Agent Editor | `DESCRIBE AGENT` produces docs |
| **Testing** | Manual testing in Studio Pro | Scriptable test cases with mxcli |

The combination of `CREATE AGENT` (document definition), tool microflows (business logic), MCP connections (external tools), knowledge bases (RAG), and ConversationalUI (chat interface) means an AI coding agent can scaffold an entire smart app from a natural-language description — creating all layers from domain model to navigation in a single MDL session.

## Open Questions

1. **CustomBlobDocument extensibility**: Will Mendix add more `CustomDocumentType` values beyond `agenteditor.agent`? If so, should we build a generic `CustomBlobDocument` parser that dispatches by `CustomDocumentType`, or keep agent-specific parsing?

2. **Contents schema stability**: The `Contents` JSON schema is owned by the agent editor extension. How stable is it across Mendix versions? Should the parser be strictly typed or loosely parse into a map?

3. **Tools in Contents vs. Runtime**: In all observed agents, `tools` and `knowledgebaseTools` are empty arrays in the document. Tools are managed via AgentCommons entities at runtime. Will future versions store tool definitions inline in the document?

4. **AgentEditorCommons sync**: The `AgentEditorCommons.Agent_CreateUpdate` microflow syncs agent documents to AgentCommons runtime entities. Should `CREATE AGENT` also trigger this sync (by calling the microflow), or only write the document and let the runtime sync on next startup?

5. **Module placement**: Agent documents currently live in AgentEditorCommons (a marketplace module). Can users create agents in their own modules? The BSON format supports it (any module can have CustomBlobDocuments), but does the Agent Editor extension expect them in a specific location?

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Contents JSON schema changes in future Mendix versions | Medium | Medium | Parse tolerantly (ignore unknown fields), version-gate new fields |
| CustomBlobDocument format changes | Low | High | Monitor Mendix release notes, BSON schema comparison |
| Studio Pro fails to open MDL-created agents | Medium | High | Test with `mx check` and Studio Pro after creation; compare BSON byte-for-byte with editor-created agents |
| Tools/KB move from runtime to document | Medium | Low | Phase 4 syntax already designed; add when format is observed |

## References

- Test project: `mx-test-projects/test3-app/test3.mpr` (Mendix 11.9.0)
- Agent documents: `mprcontents/50/15/5015e35c-...`, `f5/80/f5802216-...`, `e8/d9/e8d9c5c8-...`, `8e/bc/8ebc7f85-...`
- Agent Editor extension manifest: `.mendix-cache/modules/agenteditor.mxmodule/extensions/agent-editor/manifest.json`
- AgentCommons module: Marketplace v3.1.0 (31 entities, 226 microflows)
- MCPClient module: Marketplace v3.0.1 (20 entities, 35 microflows)
- GenAICommons module: Marketplace v6.1.0 (34 entities, 112 microflows)
- ConversationalUI module: Marketplace v6.1.0 (17 entities, 152 microflows)
