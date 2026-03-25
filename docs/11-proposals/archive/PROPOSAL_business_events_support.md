# Proposal: Business Events Support in mxcli

## Motivation

Business Events is a Mendix feature for asynchronous event-driven integration, allowing applications to publish and subscribe to business events via channels and messages. The feature was introduced in Mendix 9.8 and underwent a major redesign in 9.24.

**Important**: None of our three test projects (EnquiriesManagement, Evora-FactoryManagement, LatoProductInventory) contain any Business Event documents. This proposal is based entirely on the Mendix metamodel reflection data and generated types.

## Current State

- mxcli counts Business Events per module in `SHOW MODULES` (checks `BusinessEvents$` prefix)
- No BSON parser, reader methods, catalog table, or MDL syntax exists
- Generated metamodel types exist in `generated/metamodel/types.go` (auto-generated from reflection data)

## Business Events Metamodel (from reflection data 11.6.0)

### Document Hierarchy

```
BusinessEvents$BusinessEventService (MODEL_UNIT - top-level document)
├── Name: String
├── Documentation: String
├── ExportLevel: Enum (API | Hidden)
├── Excluded: Boolean
├── Document: String
├── Definition: BusinessEvents$BusinessEventDefinition?
│   ├── ServiceName: String
│   ├── EventNamePrefix: String
│   ├── Description: String
│   ├── Summary: String
│   └── Channels: BusinessEvents$Channel[]
│       ├── ChannelName: String
│       ├── Description: String
│       └── Messages: BusinessEvents$Message[]
│           ├── MessageName: String
│           ├── Description: String
│           ├── CanPublish: Boolean
│           ├── CanSubscribe: Boolean
│           └── Attributes: BusinessEvents$MessageAttribute[]
│               ├── AttributeName: String
│               ├── AttributeType: DomainModels$AttributeType (required)
│               ├── Description: String
│               └── EnumerationDefinition: BusinessEvents$AttributeEnumeration?
│                   └── Items: BusinessEvents$AttributeEnumerationItem[]
│                       └── Value: String
└── OperationImplementations: BusinessEvents$ServiceOperation[]
    ├── MessageName: String
    ├── Operation: String
    ├── Entity: DomainModels$Entity (BY_NAME, required)
    └── Microflow: Microflows$Microflow (BY_NAME, optional)
```

### Cross-References to Other Domains

| Reference | Target | Type |
|-----------|--------|------|
| ServiceOperation.Entity | DomainModels$Entity | BY_NAME (required) |
| ServiceOperation.Microflow | Microflows$Microflow | BY_NAME (optional) |
| MessageAttribute.AttributeType | DomainModels$AttributeType | PART |

### Version History

| Version | Changes |
|---------|---------|
| 9.8.0 | Introduced `ConsumedBusinessEventService` and `PublishedBusinessEventService` |
| 9.11.0 | Added `ConsumedBusinessEvent`, `PublishedMessage`, `PublishedMessageAttribute` |
| 9.24.0 | **Major redesign**: Replaced published/consumed model with unified `BusinessEventService`, `BusinessEventDefinition`, `Channel`, `Message` |
| 10.0.0 | Added `AttributeEnumeration` and `AttributeEnumerationItem` |
| 10.21.0 | Added `SourceApi` property |

**Note**: All Business Events types are marked as **experimental** in the TypeScript SDK.

### BSON Storage Names

All storage names match their qualified names (no aliasing):
- `BusinessEvents$BusinessEventService`
- `BusinessEvents$BusinessEventDefinition`
- `BusinessEvents$Channel`
- `BusinessEvents$Message`
- `BusinessEvents$MessageAttribute`
- `BusinessEvents$AttributeEnumeration`
- `BusinessEvents$AttributeEnumerationItem`
- `BusinessEvents$ServiceOperation`

## Implementation Plan

### Phase 1: Read-Only Support

Given that we have no real test data, this should be a lightweight implementation.

#### 1a. SDK Types (`sdk/businessevents/`)

Simple Go types mirroring the metamodel:

```go
type BusinessEventService struct {
    model.BaseElement
    Name                     string
    Documentation            string
    ExportLevel              string
    Excluded                 bool
    Definition               *BusinessEventDefinition
    OperationImplementations []*ServiceOperation
}

type BusinessEventDefinition struct {
    model.BaseElement
    ServiceName     string
    EventNamePrefix string
    Description     string
    Summary         string
    Channels        []*Channel
}

type Channel struct {
    model.BaseElement
    ChannelName string
    Description string
    Messages    []*Message
}

type Message struct {
    model.BaseElement
    MessageName  string
    Description  string
    CanPublish   bool
    CanSubscribe bool
    Attributes   []*MessageAttribute
}

type MessageAttribute struct {
    model.BaseElement
    AttributeName string
    AttributeType string // simplified
    Description   string
}

type ServiceOperation struct {
    model.BaseElement
    MessageName string
    Operation   string
    Entity      string // qualified name
    Microflow   string // qualified name
}
```

#### 1b. BSON Parser (`sdk/mpr/parser_businessevents.go`)

Parse the nested structure following reflection data field names.

#### 1c. Reader Methods

```go
func (r *Reader) ListBusinessEventServices() ([]*businessevents.BusinessEventService, error)
func (r *Reader) GetBusinessEventService(id string) (*businessevents.BusinessEventService, error)
```

#### 1d. SHOW/DESCRIBE Commands

- `SHOW BUSINESS EVENTS [IN Module]`
- `DESCRIBE BUSINESS EVENT Module.ServiceName`

Example output:
```sql
BUSINESS EVENT Module.OrderEvents
  EXPORT LEVEL: API
  DEFINITION
    SERVICE NAME: 'OrderService'
    EVENT PREFIX: 'com.example.orders'
    CHANNELS
      CHANNEL 'orders'
        MESSAGE 'OrderCreated' (Publish, Subscribe)
          ATTRIBUTES
            OrderId: Integer
            CustomerName: String
            Amount: Decimal
  OPERATIONS
    'OrderCreated' -> Module.Order (MICROFLOW Module.ACT_HandleOrderCreated)
END;
```

#### 1e. Catalog Table

Add `CATALOG.BUSINESS_EVENTS` table:

| Column | Type |
|--------|------|
| Id | TEXT |
| Name | TEXT |
| QualifiedName | TEXT |
| ModuleName | TEXT |
| Documentation | TEXT |
| ExportLevel | TEXT |
| ChannelCount | INTEGER |
| MessageCount | INTEGER |
| OperationCount | INTEGER |

#### 1f. Cross-References

- `entity` - ServiceOperation references an entity
- `call` - ServiceOperation references a microflow

## Priority Assessment

**Low priority.** Business Events:
- Are not present in any of our test projects
- Are marked as experimental in the Mendix SDK
- Have a relatively simple structure (no complex flow graphs like workflows)
- Are less commonly used than workflows in typical Mendix applications

Recommend implementing after workflow support is complete and only if user demand warrants it.

## Files to Create/Modify

### New Files
| File | Description |
|------|-------------|
| `sdk/businessevents/businessevents.go` | Type definitions |
| `sdk/mpr/parser_businessevents.go` | BSON parser |
| `mdl/catalog/builder_businessevents.go` | Catalog builder |

### Modified Files
| File | Changes |
|------|---------|
| `sdk/mpr/reader_documents.go` | Add List/Get methods |
| `mdl/catalog/tables.go` | Add table schema |
| `mdl/catalog/builder.go` | Add to build pipeline |
| `mdl/catalog/catalog.go` | Add to Tables() |
| `mdl/executor/cmd_show.go` | Add SHOW command |
| `mdl/executor/cmd_describe.go` | Add DESCRIBE command |

## Verification

Since we have no test projects with Business Events, verification would require:
1. Creating a test project with Business Events in Studio Pro
2. Or finding a Marketplace app that uses them

```bash
# When a test project with Business Events is available:
./bin/mxcli -p /path/to/app.mpr -c "SHOW BUSINESS EVENTS"
./bin/mxcli -p /path/to/app.mpr -c "DESCRIBE BUSINESS EVENT Module.ServiceName"
./bin/mxcli -p /path/to/app.mpr -c "REFRESH CATALOG FULL FORCE; SELECT * FROM CATALOG.BUSINESS_EVENTS"
```
