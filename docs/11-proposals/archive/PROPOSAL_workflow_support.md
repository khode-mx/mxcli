# Proposal: Workflow Support in mxcli

## Motivation

Workflows are a core Mendix feature (introduced in Mendix 9.0) for orchestrating multi-step business processes involving human tasks, automated actions, decisions, and parallel execution. Across our three test projects:

- **EnquiriesManagement**: 12 workflows (AgentCore: 2, BusinessProcesses: 5, WorkflowCommons: 5)
- **Evora-FactoryManagement**: 1 workflow (AltairIntegration: "WF Schedule technician appointment")
- **LatoProductInventory**: 0 workflows

Currently mxcli only *counts* workflows in `SHOW MODULES` output. It cannot parse, describe, query, or create workflow documents.

## Current State

| Capability | Status |
|-----------|--------|
| Count workflows per module | Supported (`cmd_modules.go` line 293) |
| List workflows | Not supported |
| Describe/show workflow structure | Not supported |
| Catalog table for workflows | Not supported |
| Cross-references (callers/callees) | Not supported |
| MDL syntax for workflows | Not supported |
| Create/modify workflows | Not supported |

## Workflow BSON Structure (from reflection data 11.6.0)

### Top-Level Document: `Workflows$Workflow`

| Field | Storage Name | Type | Notes |
|-------|-------------|------|-------|
| Name | Name | String | Workflow name |
| Title | Title | String | Display title (deprecated, use WorkflowName) |
| Documentation | Documentation | String | Description |
| ExportLevel | ExportLevel | Enum (API/Hidden) | Visibility |
| Parameter | Parameter | PART (`Workflows$Parameter`) | Workflow context parameter |
| Flow | Flow | PART (`Workflows$Flow`) | Main flow container |
| WorkflowName | WorkflowName | PART (`Microflows$StringTemplate`) | Runtime name template |
| WorkflowDescription | WorkflowDescription | PART (`Microflows$StringTemplate`) | Runtime description template |
| AdminPage | AdminPage | PART (`Workflows$PageReference`) | Admin overview page |
| OnWorkflowEvent | OnWorkflowEvent | PART list | Event handlers |
| WorkflowMetaData | WorkflowMetaData | PART | Visual layout data |
| DueDate | DueDate | String | Due date expression |
| Excluded | Excluded | Boolean | Excluded from build |

### Activity Type Hierarchy

```
Workflows$WorkflowActivity (abstract)
├── Workflows$StartWorkflowActivity
├── Workflows$EndWorkflowActivity
├── Workflows$SingleUserTaskActivity
├── Workflows$MultiUserTaskActivity
├── Workflows$CallMicroflowTask
├── Workflows$CallWorkflowActivity
├── Workflows$ExclusiveSplitActivity
├── Workflows$ParallelSplitActivity
├── Workflows$MergeActivity
├── Workflows$JumpToActivity
├── Workflows$WaitForTimerActivity
├── Workflows$WaitForNotificationActivity
├── Workflows$EndOfParallelSplitPathActivity
└── Workflows$EndOfBoundaryEventPathActivity
```

### Key Polymorphic Types

**User Targeting** (`Workflows$UserTargeting`):
- `NoUserTargeting`, `MicroflowUserTargeting`, `MicroflowGroupTargeting`, `XPathUserTargeting`, `XPathGroupTargeting`

**Completion Criteria** (`Workflows$UserTaskCompletionCriteria`):
- `ConsensusCompletionCriteria`, `MajorityCompletionCriteria`, `ThresholdCompletionCriteria`, `VetoCompletionCriteria`, `MicroflowCompletionCriteria`

**Condition Outcomes** (`Workflows$ConditionOutcome`):
- `BooleanConditionOutcome`, `EnumerationValueConditionOutcome`, `VoidConditionOutcome`

**Boundary Events** (`Workflows$BoundaryEvent`):
- `InterruptingTimerBoundaryEvent`, `NonInterruptingTimerBoundaryEvent`

### Real Example: FactoryManagement Workflow

The "WF Schedule technician appointment" workflow contains:
- 1 StartWorkflowActivity
- 5 EndWorkflowActivity nodes
- 9 SingleUserTaskActivity nodes (approval steps)
- 4 CallMicroflowTask nodes
- 4 JumpToActivity nodes
- 25 Flow connections
- 17 UserTaskOutcome definitions
- 9 PageReference objects (user task pages)
- 20 StringTemplate objects (task names/descriptions)

## Implementation Plan

### Phase 1: Read-Only Support (SHOW/DESCRIBE/Catalog)

**Goal**: Parse workflows, list them, describe their structure, and add to the catalog.

#### 1a. SDK Types (`sdk/workflows/`)

Create a new package with Go types for workflow documents:

```go
type Workflow struct {
    model.BaseElement
    Name          string
    Documentation string
    ExportLevel   string
    Parameter     *Parameter
    Flow          *Flow
    // ... other fields
}

type Flow struct {
    model.BaseElement
    Activities []WorkflowActivity // polymorphic
}

type WorkflowActivity interface {
    ActivityType() string
    ActivityName() string
    ActivityCaption() string
}
```

Activity types: `StartWorkflowActivity`, `EndWorkflowActivity`, `SingleUserTaskActivity`, `MultiUserTaskActivity`, `CallMicroflowTask`, `CallWorkflowActivity`, `ExclusiveSplitActivity`, `ParallelSplitActivity`, `MergeActivity`, `JumpToActivity`, `WaitForTimerActivity`, `WaitForNotificationActivity`

#### 1b. BSON Parser (`sdk/mpr/parser_workflow.go`)

Add workflow document parsing following the pattern of `parser_microflow.go`:
- Parse top-level Workflow fields
- Parse Flow and Activities array with polymorphic type dispatch
- Parse parameter mappings, outcomes, user targeting

#### 1c. Reader Methods (`sdk/mpr/reader_documents.go`)

```go
func (r *Reader) ListWorkflows() ([]*workflows.Workflow, error)
func (r *Reader) GetWorkflow(id string) (*workflows.Workflow, error)
```

#### 1d. SHOW/DESCRIBE Commands

- `SHOW WORKFLOWS [IN Module]` - List all workflows with activity counts
- `DESCRIBE WORKFLOW Module.WorkflowName` - Full MDL-style output

Example describe output:
```sql
WORKFLOW AgentCore.WF_EnquiryProcessing
  PARAMETER $WorkflowContext: AgentCore.Enquiry

  START start1;

  CALL MICROFLOW AgentCore.ACT_ClassifyEnquiry
    ($Enquiry = $WorkflowContext)
    ON Boolean DO
      TRUE -> userTask1;
      FALSE -> end1;
    END;

  USER TASK userTask1 'Review Classification'
    PAGE AgentCore.WF_ReviewClassification
    TARGETING MICROFLOW AgentCore.DS_GetReviewers
    OUTCOMES
      'Approve' -> callMicroflow2;
      'Reject' -> end2;
    END;

  END end1;
  END end2;
END WORKFLOW;
```

#### 1e. Catalog Table (`mdl/catalog/`)

Add `CATALOG.WORKFLOWS` table:

| Column | Type | Description |
|--------|------|-------------|
| Id | TEXT | Workflow UUID |
| Name | TEXT | Workflow name |
| QualifiedName | TEXT | Module.WorkflowName |
| ModuleName | TEXT | Module name |
| Folder | TEXT | Folder path |
| Documentation | TEXT | Description |
| ExportLevel | TEXT | API or Hidden |
| ActivityCount | INTEGER | Total activities |
| UserTaskCount | INTEGER | User task count |
| MicroflowCallCount | INTEGER | Microflow call count |
| DecisionCount | INTEGER | Decision/split count |
| ParameterEntity | TEXT | Context entity QN |

Add to `objects` view UNION with `ObjectType = 'WORKFLOW'`.

#### 1f. Cross-References (refs table)

Add workflow-related reference types:
- `call_microflow` - CallMicroflowTask references a microflow
- `call_workflow` - CallWorkflowActivity references a sub-workflow
- `show_page` - UserTask references a task page
- `parameter` - Workflow parameter references an entity
- `user_targeting` - MicroflowUserTargeting references a microflow
- `admin_page` - AdminPage references a page

#### 1g. Source Generation

Add workflows to `buildSource()` in `builder_source.go` so they appear in full-text search results.

### Phase 2: Context & Navigation

- Add workflows to `SHOW CONTEXT OF` command
- Add workflows to `SHOW CALLERS OF` / `SHOW CALLEES OF`
- Add workflows to `SHOW IMPACT OF`
- Add `detectElementType` support for workflows in `cmd_context.go`

### Phase 3: MDL Syntax (Future)

Define MDL grammar for creating/modifying workflows:

```sql
CREATE WORKFLOW Module.WF_ProcessOrder
  PARAMETER $Context: Module.Order
BEGIN
  START;

  CALL MICROFLOW Module.ACT_ValidateOrder ($Order = $Context)
    ON Boolean
      TRUE -> reviewTask;
      FALSE -> endReject;

  USER TASK reviewTask 'Review Order'
    PAGE Module.WF_ReviewOrder
    TARGETING MICROFLOW Module.DS_GetReviewers
    OUTCOMES
      'Approve' -> endApprove;
      'Reject' -> endReject;

  END endApprove;
  END endReject;
END;
```

This is complex and should be deferred until read-only support is stable.

## Files to Create/Modify

### New Files
| File | Description |
|------|-------------|
| `sdk/workflows/workflow.go` | Workflow, Flow, Parameter types |
| `sdk/workflows/activities.go` | All activity type definitions |
| `sdk/workflows/targeting.go` | UserTargeting types |
| `sdk/workflows/outcomes.go` | Outcome and criteria types |
| `sdk/mpr/parser_workflow.go` | BSON parsing for workflow documents |
| `mdl/catalog/builder_workflows.go` | Catalog builder for workflows |

### Modified Files
| File | Changes |
|------|---------|
| `sdk/mpr/reader_documents.go` | Add ListWorkflows(), GetWorkflow() |
| `mdl/catalog/tables.go` | Add workflows table schema |
| `mdl/catalog/builder.go` | Add buildWorkflows() to pipeline |
| `mdl/catalog/builder_references.go` | Add workflow cross-references |
| `mdl/catalog/builder_source.go` | Add workflows to source generation |
| `mdl/catalog/builder_strings.go` | Extract strings from workflows |
| `mdl/catalog/catalog.go` | Add to Tables() list |
| `mdl/executor/cmd_show.go` | Add SHOW WORKFLOWS command |
| `mdl/executor/cmd_describe.go` | Add DESCRIBE WORKFLOW command |
| `mdl/executor/cmd_context.go` | Add workflow detection |

## Effort Estimate

| Phase | Scope | Complexity |
|-------|-------|-----------|
| Phase 1a-1b | SDK types + BSON parser | High (14 activity types, many polymorphic types) |
| Phase 1c | Reader methods | Low (follows existing pattern) |
| Phase 1d | SHOW/DESCRIBE | Medium (workflow-specific formatting) |
| Phase 1e | Catalog table | Low (follows java_actions pattern) |
| Phase 1f | Cross-references | Medium (multiple reference types) |
| Phase 2 | Context/navigation | Low (extend existing commands) |
| Phase 3 | MDL create syntax | High (grammar + executor) |

## Verification

```bash
make build && make test

# Test with FactoryManagement project (has 1 workflow)
./bin/mxcli -p mx-test-projects/Evora-FactoryManagement/Evora-FactoryManagement.mpr \
  -c "SHOW WORKFLOWS"

./bin/mxcli -p mx-test-projects/Evora-FactoryManagement/Evora-FactoryManagement.mpr \
  -c "DESCRIBE WORKFLOW AltairIntegration.WF_Schedule_technician_appointment"

# Test with EnquiriesManagement project (has 12 workflows)
./bin/mxcli -p mx-test-projects/EnquiriesManagement/EnquiriesManagement.mpr \
  -c "SHOW WORKFLOWS"

# Catalog queries
./bin/mxcli -p mx-test-projects/EnquiriesManagement/EnquiriesManagement.mpr \
  -c "REFRESH CATALOG FULL FORCE; SELECT * FROM CATALOG.WORKFLOWS"

# Cross-references
./bin/mxcli -p mx-test-projects/EnquiriesManagement/EnquiriesManagement.mpr \
  -c "REFRESH CATALOG FULL FORCE; SHOW CALLERS OF AgentCore.ACT_ClassifyEnquiry"
```
