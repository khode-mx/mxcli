// SPDX-License-Identifier: Apache-2.0

// Package mock provides a configurable MockBackend for testing that
// implements backend.FullBackend. Each method delegates to an optional
// function field; when the field is nil the method returns zero values.
package mock

import (
	"github.com/mendixlabs/mxcli/mdl/backend"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
	"github.com/mendixlabs/mxcli/sdk/javaactions"
	"github.com/mendixlabs/mxcli/sdk/microflows"
	"github.com/mendixlabs/mxcli/sdk/mpr"
	"github.com/mendixlabs/mxcli/sdk/mpr/version"
	"github.com/mendixlabs/mxcli/sdk/pages"
	"github.com/mendixlabs/mxcli/sdk/security"
	"github.com/mendixlabs/mxcli/sdk/workflows"
)

var _ backend.FullBackend = (*MockBackend)(nil)

// MockBackend implements backend.FullBackend. Every interface method is
// backed by a public function field. If the field is nil the method
// returns zero values / nil error (never panics).
type MockBackend struct {
	// ConnectionBackend
	ConnectFunc          func(path string) error
	DisconnectFunc       func() error
	CommitFunc           func() error
	IsConnectedFunc      func() bool
	PathFunc             func() string
	VersionFunc          func() mpr.MPRVersion
	ProjectVersionFunc   func() *version.ProjectVersion
	GetMendixVersionFunc func() (string, error)

	// ModuleBackend
	ListModulesFunc             func() ([]*model.Module, error)
	GetModuleFunc               func(id model.ID) (*model.Module, error)
	GetModuleByNameFunc         func(name string) (*model.Module, error)
	CreateModuleFunc            func(module *model.Module) error
	UpdateModuleFunc            func(module *model.Module) error
	DeleteModuleFunc            func(id model.ID) error
	DeleteModuleWithCleanupFunc func(id model.ID, moduleName string) error

	// FolderBackend
	ListFoldersFunc  func() ([]*mpr.FolderInfo, error)
	CreateFolderFunc func(folder *model.Folder) error
	DeleteFolderFunc func(id model.ID) error
	MoveFolderFunc   func(id model.ID, newContainerID model.ID) error

	// DomainModelBackend
	ListDomainModelsFunc                       func() ([]*domainmodel.DomainModel, error)
	GetDomainModelFunc                         func(moduleID model.ID) (*domainmodel.DomainModel, error)
	GetDomainModelByIDFunc                     func(id model.ID) (*domainmodel.DomainModel, error)
	UpdateDomainModelFunc                      func(dm *domainmodel.DomainModel) error
	CreateEntityFunc                           func(domainModelID model.ID, entity *domainmodel.Entity) error
	UpdateEntityFunc                           func(domainModelID model.ID, entity *domainmodel.Entity) error
	DeleteEntityFunc                           func(domainModelID model.ID, entityID model.ID) error
	MoveEntityFunc                             func(entity *domainmodel.Entity, sourceDMID, targetDMID model.ID, sourceModuleName, targetModuleName string) ([]string, error)
	AddAttributeFunc                           func(domainModelID model.ID, entityID model.ID, attr *domainmodel.Attribute) error
	UpdateAttributeFunc                        func(domainModelID model.ID, entityID model.ID, attr *domainmodel.Attribute) error
	DeleteAttributeFunc                        func(domainModelID model.ID, entityID model.ID, attrID model.ID) error
	CreateAssociationFunc                      func(domainModelID model.ID, assoc *domainmodel.Association) error
	CreateCrossAssociationFunc                 func(domainModelID model.ID, ca *domainmodel.CrossModuleAssociation) error
	DeleteAssociationFunc                      func(domainModelID model.ID, assocID model.ID) error
	DeleteCrossAssociationFunc                 func(domainModelID model.ID, assocID model.ID) error
	CreateViewEntitySourceDocumentFunc         func(moduleID model.ID, moduleName, docName, oqlQuery, documentation string) (model.ID, error)
	DeleteViewEntitySourceDocumentFunc         func(id model.ID) error
	DeleteViewEntitySourceDocumentByNameFunc   func(moduleName, docName string) error
	FindViewEntitySourceDocumentIDFunc         func(moduleName, docName string) (model.ID, error)
	FindAllViewEntitySourceDocumentIDsFunc     func(moduleName, docName string) ([]model.ID, error)
	MoveViewEntitySourceDocumentFunc           func(sourceModuleName string, targetModuleID model.ID, docName string) error
	UpdateOqlQueriesForMovedEntityFunc         func(oldQualifiedName, newQualifiedName string) (int, error)
	UpdateEnumerationRefsInAllDomainModelsFunc func(oldQualifiedName, newQualifiedName string) error

	// MicroflowBackend
	ListMicroflowsFunc  func() ([]*microflows.Microflow, error)
	GetMicroflowFunc    func(id model.ID) (*microflows.Microflow, error)
	CreateMicroflowFunc func(mf *microflows.Microflow) error
	UpdateMicroflowFunc func(mf *microflows.Microflow) error
	DeleteMicroflowFunc func(id model.ID) error
	MoveMicroflowFunc   func(mf *microflows.Microflow) error
	ListNanoflowsFunc   func() ([]*microflows.Nanoflow, error)
	GetNanoflowFunc     func(id model.ID) (*microflows.Nanoflow, error)
	CreateNanoflowFunc  func(nf *microflows.Nanoflow) error
	UpdateNanoflowFunc  func(nf *microflows.Nanoflow) error
	DeleteNanoflowFunc  func(id model.ID) error
	MoveNanoflowFunc    func(nf *microflows.Nanoflow) error

	// PageBackend
	ListPagesFunc          func() ([]*pages.Page, error)
	GetPageFunc            func(id model.ID) (*pages.Page, error)
	CreatePageFunc         func(page *pages.Page) error
	UpdatePageFunc         func(page *pages.Page) error
	DeletePageFunc         func(id model.ID) error
	MovePageFunc           func(page *pages.Page) error
	ListLayoutsFunc        func() ([]*pages.Layout, error)
	GetLayoutFunc          func(id model.ID) (*pages.Layout, error)
	CreateLayoutFunc       func(layout *pages.Layout) error
	UpdateLayoutFunc       func(layout *pages.Layout) error
	DeleteLayoutFunc       func(id model.ID) error
	ListSnippetsFunc       func() ([]*pages.Snippet, error)
	CreateSnippetFunc      func(snippet *pages.Snippet) error
	UpdateSnippetFunc      func(snippet *pages.Snippet) error
	DeleteSnippetFunc      func(id model.ID) error
	MoveSnippetFunc        func(snippet *pages.Snippet) error
	ListBuildingBlocksFunc func() ([]*pages.BuildingBlock, error)
	ListPageTemplatesFunc  func() ([]*pages.PageTemplate, error)

	// EnumerationBackend
	ListEnumerationsFunc  func() ([]*model.Enumeration, error)
	GetEnumerationFunc    func(id model.ID) (*model.Enumeration, error)
	CreateEnumerationFunc func(enum *model.Enumeration) error
	UpdateEnumerationFunc func(enum *model.Enumeration) error
	MoveEnumerationFunc   func(enum *model.Enumeration) error
	DeleteEnumerationFunc func(id model.ID) error

	// ConstantBackend
	ListConstantsFunc  func() ([]*model.Constant, error)
	GetConstantFunc    func(id model.ID) (*model.Constant, error)
	CreateConstantFunc func(constant *model.Constant) error
	UpdateConstantFunc func(constant *model.Constant) error
	MoveConstantFunc   func(constant *model.Constant) error
	DeleteConstantFunc func(id model.ID) error

	// SecurityBackend
	GetProjectSecurityFunc               func() (*security.ProjectSecurity, error)
	SetProjectSecurityLevelFunc          func(unitID model.ID, level string) error
	SetProjectDemoUsersEnabledFunc       func(unitID model.ID, enabled bool) error
	AddUserRoleFunc                      func(unitID model.ID, name string, moduleRoles []string, manageAllRoles bool) error
	AlterUserRoleModuleRolesFunc         func(unitID model.ID, userRoleName string, add bool, moduleRoles []string) error
	RemoveUserRoleFunc                   func(unitID model.ID, name string) error
	AddDemoUserFunc                      func(unitID model.ID, userName, password, entity string, userRoles []string) error
	RemoveDemoUserFunc                   func(unitID model.ID, userName string) error
	ListModuleSecurityFunc               func() ([]*security.ModuleSecurity, error)
	GetModuleSecurityFunc                func(moduleID model.ID) (*security.ModuleSecurity, error)
	AddModuleRoleFunc                    func(unitID model.ID, roleName, description string) error
	RemoveModuleRoleFunc                 func(unitID model.ID, roleName string) error
	RemoveModuleRoleFromAllUserRolesFunc func(unitID model.ID, qualifiedRole string) (int, error)
	UpdateAllowedRolesFunc               func(unitID model.ID, roles []string) error
	UpdatePublishedRestServiceRolesFunc  func(unitID model.ID, roles []string) error
	RemoveFromAllowedRolesFunc           func(unitID model.ID, roleName string) (bool, error)
	AddEntityAccessRuleFunc              func(unitID model.ID, entityName string, roleNames []string, allowCreate, allowDelete bool, defaultMemberAccess string, xpathConstraint string, memberAccesses []mpr.EntityMemberAccess) error
	RemoveEntityAccessRuleFunc           func(unitID model.ID, entityName string, roleNames []string) (int, error)
	RevokeEntityMemberAccessFunc         func(unitID model.ID, entityName string, roleNames []string, revocation mpr.EntityAccessRevocation) (int, error)
	RemoveRoleFromAllEntitiesFunc        func(unitID model.ID, roleName string) (int, error)
	ReconcileMemberAccessesFunc          func(unitID model.ID, moduleName string) (int, error)

	// NavigationBackend
	ListNavigationDocumentsFunc func() ([]*mpr.NavigationDocument, error)
	GetNavigationFunc           func() (*mpr.NavigationDocument, error)
	UpdateNavigationProfileFunc func(navDocID model.ID, profileName string, spec mpr.NavigationProfileSpec) error

	// ServiceBackend
	ListConsumedODataServicesFunc   func() ([]*model.ConsumedODataService, error)
	ListPublishedODataServicesFunc  func() ([]*model.PublishedODataService, error)
	CreateConsumedODataServiceFunc  func(svc *model.ConsumedODataService) error
	UpdateConsumedODataServiceFunc  func(svc *model.ConsumedODataService) error
	DeleteConsumedODataServiceFunc  func(id model.ID) error
	CreatePublishedODataServiceFunc func(svc *model.PublishedODataService) error
	UpdatePublishedODataServiceFunc func(svc *model.PublishedODataService) error
	DeletePublishedODataServiceFunc func(id model.ID) error
	ListConsumedRestServicesFunc    func() ([]*model.ConsumedRestService, error)
	ListPublishedRestServicesFunc   func() ([]*model.PublishedRestService, error)
	CreateConsumedRestServiceFunc   func(svc *model.ConsumedRestService) error
	UpdateConsumedRestServiceFunc   func(svc *model.ConsumedRestService) error
	DeleteConsumedRestServiceFunc   func(id model.ID) error
	CreatePublishedRestServiceFunc  func(svc *model.PublishedRestService) error
	UpdatePublishedRestServiceFunc  func(svc *model.PublishedRestService) error
	DeletePublishedRestServiceFunc  func(id model.ID) error
	ListBusinessEventServicesFunc   func() ([]*model.BusinessEventService, error)
	CreateBusinessEventServiceFunc  func(svc *model.BusinessEventService) error
	UpdateBusinessEventServiceFunc  func(svc *model.BusinessEventService) error
	DeleteBusinessEventServiceFunc  func(id model.ID) error
	ListDatabaseConnectionsFunc     func() ([]*model.DatabaseConnection, error)
	CreateDatabaseConnectionFunc    func(conn *model.DatabaseConnection) error
	UpdateDatabaseConnectionFunc    func(conn *model.DatabaseConnection) error
	MoveDatabaseConnectionFunc      func(conn *model.DatabaseConnection) error
	DeleteDatabaseConnectionFunc    func(id model.ID) error
	ListDataTransformersFunc        func() ([]*model.DataTransformer, error)
	CreateDataTransformerFunc       func(dt *model.DataTransformer) error
	DeleteDataTransformerFunc       func(id model.ID) error

	// MappingBackend
	ListImportMappingsFunc              func() ([]*model.ImportMapping, error)
	GetImportMappingByQualifiedNameFunc func(moduleName, name string) (*model.ImportMapping, error)
	CreateImportMappingFunc             func(im *model.ImportMapping) error
	UpdateImportMappingFunc             func(im *model.ImportMapping) error
	DeleteImportMappingFunc             func(id model.ID) error
	MoveImportMappingFunc               func(im *model.ImportMapping) error
	ListExportMappingsFunc              func() ([]*model.ExportMapping, error)
	GetExportMappingByQualifiedNameFunc func(moduleName, name string) (*model.ExportMapping, error)
	CreateExportMappingFunc             func(em *model.ExportMapping) error
	UpdateExportMappingFunc             func(em *model.ExportMapping) error
	DeleteExportMappingFunc             func(id model.ID) error
	MoveExportMappingFunc               func(em *model.ExportMapping) error
	ListJsonStructuresFunc              func() ([]*mpr.JsonStructure, error)
	GetJsonStructureByQualifiedNameFunc func(moduleName, name string) (*mpr.JsonStructure, error)
	CreateJsonStructureFunc             func(js *mpr.JsonStructure) error
	DeleteJsonStructureFunc             func(id string) error

	// JavaBackend
	ListJavaActionsFunc            func() ([]*mpr.JavaAction, error)
	ListJavaScriptActionsFunc      func() ([]*mpr.JavaScriptAction, error)
	ReadJavaActionByNameFunc       func(qualifiedName string) (*javaactions.JavaAction, error)
	ReadJavaScriptActionByNameFunc func(qualifiedName string) (*mpr.JavaScriptAction, error)
	CreateJavaActionFunc           func(ja *javaactions.JavaAction) error
	UpdateJavaActionFunc           func(ja *javaactions.JavaAction) error
	DeleteJavaActionFunc           func(id model.ID) error
	WriteJavaSourceFileFunc        func(moduleName, actionName string, javaCode string, params []*javaactions.JavaActionParameter, returnType javaactions.CodeActionReturnType) error
	ReadJavaSourceFileFunc         func(moduleName, actionName string) (string, error)

	// WorkflowBackend
	ListWorkflowsFunc  func() ([]*workflows.Workflow, error)
	GetWorkflowFunc    func(id model.ID) (*workflows.Workflow, error)
	CreateWorkflowFunc func(wf *workflows.Workflow) error
	DeleteWorkflowFunc func(id model.ID) error

	// SettingsBackend
	GetProjectSettingsFunc    func() (*model.ProjectSettings, error)
	UpdateProjectSettingsFunc func(ps *model.ProjectSettings) error

	// ImageBackend
	ListImageCollectionsFunc  func() ([]*mpr.ImageCollection, error)
	CreateImageCollectionFunc func(ic *mpr.ImageCollection) error
	DeleteImageCollectionFunc func(id string) error

	// ScheduledEventBackend
	ListScheduledEventsFunc func() ([]*model.ScheduledEvent, error)
	GetScheduledEventFunc   func(id model.ID) (*model.ScheduledEvent, error)

	// RenameBackend
	UpdateQualifiedNameInAllUnitsFunc func(oldName, newName string) (int, error)
	RenameReferencesFunc              func(oldName, newName string, dryRun bool) ([]mpr.RenameHit, error)
	RenameDocumentByNameFunc          func(moduleName, oldName, newName string) error

	// RawUnitBackend
	GetRawUnitFunc            func(id model.ID) (map[string]any, error)
	GetRawUnitBytesFunc       func(id string) ([]byte, error)
	ListRawUnitsByTypeFunc    func(typePrefix string) ([]*mpr.RawUnit, error)
	ListRawUnitsFunc          func(objectType string) ([]*mpr.RawUnitInfo, error)
	GetRawUnitByNameFunc      func(objectType, qualifiedName string) (*mpr.RawUnitInfo, error)
	GetRawMicroflowByNameFunc func(qualifiedName string) ([]byte, error)
	UpdateRawUnitFunc         func(unitID string, contents []byte) error

	// MetadataBackend
	ListAllUnitIDsFunc   func() ([]string, error)
	ListUnitsFunc        func() ([]*mpr.UnitInfo, error)
	GetUnitTypesFunc     func() (map[string]int, error)
	GetProjectRootIDFunc func() (string, error)
	ContentsDirFunc      func() string
	ExportJSONFunc       func() ([]byte, error)
	InvalidateCacheFunc  func()

	// WidgetBackend
	FindCustomWidgetTypeFunc     func(widgetID string) (*mpr.RawCustomWidgetType, error)
	FindAllCustomWidgetTypesFunc func(widgetID string) ([]*mpr.RawCustomWidgetType, error)

	// AgentEditorBackend
	ListAgentEditorModelsFunc              func() ([]*agenteditor.Model, error)
	ListAgentEditorKnowledgeBasesFunc      func() ([]*agenteditor.KnowledgeBase, error)
	ListAgentEditorConsumedMCPServicesFunc func() ([]*agenteditor.ConsumedMCPService, error)
	ListAgentEditorAgentsFunc              func() ([]*agenteditor.Agent, error)
}

// ---------------------------------------------------------------------------
// ConnectionBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) Connect(path string) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(path)
	}
	return nil
}

func (m *MockBackend) Disconnect() error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc()
	}
	return nil
}

func (m *MockBackend) Commit() error {
	if m.CommitFunc != nil {
		return m.CommitFunc()
	}
	return nil
}

func (m *MockBackend) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	return false
}

func (m *MockBackend) Path() string {
	if m.PathFunc != nil {
		return m.PathFunc()
	}
	return ""
}

func (m *MockBackend) Version() mpr.MPRVersion {
	if m.VersionFunc != nil {
		return m.VersionFunc()
	}
	var zero mpr.MPRVersion
	return zero
}

func (m *MockBackend) ProjectVersion() *version.ProjectVersion {
	if m.ProjectVersionFunc != nil {
		return m.ProjectVersionFunc()
	}
	return nil
}

func (m *MockBackend) GetMendixVersion() (string, error) {
	if m.GetMendixVersionFunc != nil {
		return m.GetMendixVersionFunc()
	}
	return "", nil
}

// ---------------------------------------------------------------------------
// ModuleBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListModules() ([]*model.Module, error) {
	if m.ListModulesFunc != nil {
		return m.ListModulesFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetModule(id model.ID) (*model.Module, error) {
	if m.GetModuleFunc != nil {
		return m.GetModuleFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) GetModuleByName(name string) (*model.Module, error) {
	if m.GetModuleByNameFunc != nil {
		return m.GetModuleByNameFunc(name)
	}
	return nil, nil
}

func (m *MockBackend) CreateModule(module *model.Module) error {
	if m.CreateModuleFunc != nil {
		return m.CreateModuleFunc(module)
	}
	return nil
}

func (m *MockBackend) UpdateModule(module *model.Module) error {
	if m.UpdateModuleFunc != nil {
		return m.UpdateModuleFunc(module)
	}
	return nil
}

func (m *MockBackend) DeleteModule(id model.ID) error {
	if m.DeleteModuleFunc != nil {
		return m.DeleteModuleFunc(id)
	}
	return nil
}

func (m *MockBackend) DeleteModuleWithCleanup(id model.ID, moduleName string) error {
	if m.DeleteModuleWithCleanupFunc != nil {
		return m.DeleteModuleWithCleanupFunc(id, moduleName)
	}
	return nil
}

// ---------------------------------------------------------------------------
// FolderBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListFolders() ([]*mpr.FolderInfo, error) {
	if m.ListFoldersFunc != nil {
		return m.ListFoldersFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateFolder(folder *model.Folder) error {
	if m.CreateFolderFunc != nil {
		return m.CreateFolderFunc(folder)
	}
	return nil
}

func (m *MockBackend) DeleteFolder(id model.ID) error {
	if m.DeleteFolderFunc != nil {
		return m.DeleteFolderFunc(id)
	}
	return nil
}

func (m *MockBackend) MoveFolder(id model.ID, newContainerID model.ID) error {
	if m.MoveFolderFunc != nil {
		return m.MoveFolderFunc(id, newContainerID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// DomainModelBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListDomainModels() ([]*domainmodel.DomainModel, error) {
	if m.ListDomainModelsFunc != nil {
		return m.ListDomainModelsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetDomainModel(moduleID model.ID) (*domainmodel.DomainModel, error) {
	if m.GetDomainModelFunc != nil {
		return m.GetDomainModelFunc(moduleID)
	}
	return nil, nil
}

func (m *MockBackend) GetDomainModelByID(id model.ID) (*domainmodel.DomainModel, error) {
	if m.GetDomainModelByIDFunc != nil {
		return m.GetDomainModelByIDFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) UpdateDomainModel(dm *domainmodel.DomainModel) error {
	if m.UpdateDomainModelFunc != nil {
		return m.UpdateDomainModelFunc(dm)
	}
	return nil
}

func (m *MockBackend) CreateEntity(domainModelID model.ID, entity *domainmodel.Entity) error {
	if m.CreateEntityFunc != nil {
		return m.CreateEntityFunc(domainModelID, entity)
	}
	return nil
}

func (m *MockBackend) UpdateEntity(domainModelID model.ID, entity *domainmodel.Entity) error {
	if m.UpdateEntityFunc != nil {
		return m.UpdateEntityFunc(domainModelID, entity)
	}
	return nil
}

func (m *MockBackend) DeleteEntity(domainModelID model.ID, entityID model.ID) error {
	if m.DeleteEntityFunc != nil {
		return m.DeleteEntityFunc(domainModelID, entityID)
	}
	return nil
}

func (m *MockBackend) MoveEntity(entity *domainmodel.Entity, sourceDMID, targetDMID model.ID, sourceModuleName, targetModuleName string) ([]string, error) {
	if m.MoveEntityFunc != nil {
		return m.MoveEntityFunc(entity, sourceDMID, targetDMID, sourceModuleName, targetModuleName)
	}
	return nil, nil
}

func (m *MockBackend) AddAttribute(domainModelID model.ID, entityID model.ID, attr *domainmodel.Attribute) error {
	if m.AddAttributeFunc != nil {
		return m.AddAttributeFunc(domainModelID, entityID, attr)
	}
	return nil
}

func (m *MockBackend) UpdateAttribute(domainModelID model.ID, entityID model.ID, attr *domainmodel.Attribute) error {
	if m.UpdateAttributeFunc != nil {
		return m.UpdateAttributeFunc(domainModelID, entityID, attr)
	}
	return nil
}

func (m *MockBackend) DeleteAttribute(domainModelID model.ID, entityID model.ID, attrID model.ID) error {
	if m.DeleteAttributeFunc != nil {
		return m.DeleteAttributeFunc(domainModelID, entityID, attrID)
	}
	return nil
}

func (m *MockBackend) CreateAssociation(domainModelID model.ID, assoc *domainmodel.Association) error {
	if m.CreateAssociationFunc != nil {
		return m.CreateAssociationFunc(domainModelID, assoc)
	}
	return nil
}

func (m *MockBackend) CreateCrossAssociation(domainModelID model.ID, ca *domainmodel.CrossModuleAssociation) error {
	if m.CreateCrossAssociationFunc != nil {
		return m.CreateCrossAssociationFunc(domainModelID, ca)
	}
	return nil
}

func (m *MockBackend) DeleteAssociation(domainModelID model.ID, assocID model.ID) error {
	if m.DeleteAssociationFunc != nil {
		return m.DeleteAssociationFunc(domainModelID, assocID)
	}
	return nil
}

func (m *MockBackend) DeleteCrossAssociation(domainModelID model.ID, assocID model.ID) error {
	if m.DeleteCrossAssociationFunc != nil {
		return m.DeleteCrossAssociationFunc(domainModelID, assocID)
	}
	return nil
}

func (m *MockBackend) CreateViewEntitySourceDocument(moduleID model.ID, moduleName, docName, oqlQuery, documentation string) (model.ID, error) {
	if m.CreateViewEntitySourceDocumentFunc != nil {
		return m.CreateViewEntitySourceDocumentFunc(moduleID, moduleName, docName, oqlQuery, documentation)
	}
	return "", nil
}

func (m *MockBackend) DeleteViewEntitySourceDocument(id model.ID) error {
	if m.DeleteViewEntitySourceDocumentFunc != nil {
		return m.DeleteViewEntitySourceDocumentFunc(id)
	}
	return nil
}

func (m *MockBackend) DeleteViewEntitySourceDocumentByName(moduleName, docName string) error {
	if m.DeleteViewEntitySourceDocumentByNameFunc != nil {
		return m.DeleteViewEntitySourceDocumentByNameFunc(moduleName, docName)
	}
	return nil
}

func (m *MockBackend) FindViewEntitySourceDocumentID(moduleName, docName string) (model.ID, error) {
	if m.FindViewEntitySourceDocumentIDFunc != nil {
		return m.FindViewEntitySourceDocumentIDFunc(moduleName, docName)
	}
	return "", nil
}

func (m *MockBackend) FindAllViewEntitySourceDocumentIDs(moduleName, docName string) ([]model.ID, error) {
	if m.FindAllViewEntitySourceDocumentIDsFunc != nil {
		return m.FindAllViewEntitySourceDocumentIDsFunc(moduleName, docName)
	}
	return nil, nil
}

func (m *MockBackend) MoveViewEntitySourceDocument(sourceModuleName string, targetModuleID model.ID, docName string) error {
	if m.MoveViewEntitySourceDocumentFunc != nil {
		return m.MoveViewEntitySourceDocumentFunc(sourceModuleName, targetModuleID, docName)
	}
	return nil
}

func (m *MockBackend) UpdateOqlQueriesForMovedEntity(oldQualifiedName, newQualifiedName string) (int, error) {
	if m.UpdateOqlQueriesForMovedEntityFunc != nil {
		return m.UpdateOqlQueriesForMovedEntityFunc(oldQualifiedName, newQualifiedName)
	}
	return 0, nil
}

func (m *MockBackend) UpdateEnumerationRefsInAllDomainModels(oldQualifiedName, newQualifiedName string) error {
	if m.UpdateEnumerationRefsInAllDomainModelsFunc != nil {
		return m.UpdateEnumerationRefsInAllDomainModelsFunc(oldQualifiedName, newQualifiedName)
	}
	return nil
}

// ---------------------------------------------------------------------------
// MicroflowBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListMicroflows() ([]*microflows.Microflow, error) {
	if m.ListMicroflowsFunc != nil {
		return m.ListMicroflowsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetMicroflow(id model.ID) (*microflows.Microflow, error) {
	if m.GetMicroflowFunc != nil {
		return m.GetMicroflowFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) CreateMicroflow(mf *microflows.Microflow) error {
	if m.CreateMicroflowFunc != nil {
		return m.CreateMicroflowFunc(mf)
	}
	return nil
}

func (m *MockBackend) UpdateMicroflow(mf *microflows.Microflow) error {
	if m.UpdateMicroflowFunc != nil {
		return m.UpdateMicroflowFunc(mf)
	}
	return nil
}

func (m *MockBackend) DeleteMicroflow(id model.ID) error {
	if m.DeleteMicroflowFunc != nil {
		return m.DeleteMicroflowFunc(id)
	}
	return nil
}

func (m *MockBackend) MoveMicroflow(mf *microflows.Microflow) error {
	if m.MoveMicroflowFunc != nil {
		return m.MoveMicroflowFunc(mf)
	}
	return nil
}

func (m *MockBackend) ListNanoflows() ([]*microflows.Nanoflow, error) {
	if m.ListNanoflowsFunc != nil {
		return m.ListNanoflowsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetNanoflow(id model.ID) (*microflows.Nanoflow, error) {
	if m.GetNanoflowFunc != nil {
		return m.GetNanoflowFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) CreateNanoflow(nf *microflows.Nanoflow) error {
	if m.CreateNanoflowFunc != nil {
		return m.CreateNanoflowFunc(nf)
	}
	return nil
}

func (m *MockBackend) UpdateNanoflow(nf *microflows.Nanoflow) error {
	if m.UpdateNanoflowFunc != nil {
		return m.UpdateNanoflowFunc(nf)
	}
	return nil
}

func (m *MockBackend) DeleteNanoflow(id model.ID) error {
	if m.DeleteNanoflowFunc != nil {
		return m.DeleteNanoflowFunc(id)
	}
	return nil
}

func (m *MockBackend) MoveNanoflow(nf *microflows.Nanoflow) error {
	if m.MoveNanoflowFunc != nil {
		return m.MoveNanoflowFunc(nf)
	}
	return nil
}

// ---------------------------------------------------------------------------
// PageBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListPages() ([]*pages.Page, error) {
	if m.ListPagesFunc != nil {
		return m.ListPagesFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetPage(id model.ID) (*pages.Page, error) {
	if m.GetPageFunc != nil {
		return m.GetPageFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) CreatePage(page *pages.Page) error {
	if m.CreatePageFunc != nil {
		return m.CreatePageFunc(page)
	}
	return nil
}

func (m *MockBackend) UpdatePage(page *pages.Page) error {
	if m.UpdatePageFunc != nil {
		return m.UpdatePageFunc(page)
	}
	return nil
}

func (m *MockBackend) DeletePage(id model.ID) error {
	if m.DeletePageFunc != nil {
		return m.DeletePageFunc(id)
	}
	return nil
}

func (m *MockBackend) MovePage(page *pages.Page) error {
	if m.MovePageFunc != nil {
		return m.MovePageFunc(page)
	}
	return nil
}

func (m *MockBackend) ListLayouts() ([]*pages.Layout, error) {
	if m.ListLayoutsFunc != nil {
		return m.ListLayoutsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetLayout(id model.ID) (*pages.Layout, error) {
	if m.GetLayoutFunc != nil {
		return m.GetLayoutFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) CreateLayout(layout *pages.Layout) error {
	if m.CreateLayoutFunc != nil {
		return m.CreateLayoutFunc(layout)
	}
	return nil
}

func (m *MockBackend) UpdateLayout(layout *pages.Layout) error {
	if m.UpdateLayoutFunc != nil {
		return m.UpdateLayoutFunc(layout)
	}
	return nil
}

func (m *MockBackend) DeleteLayout(id model.ID) error {
	if m.DeleteLayoutFunc != nil {
		return m.DeleteLayoutFunc(id)
	}
	return nil
}

func (m *MockBackend) ListSnippets() ([]*pages.Snippet, error) {
	if m.ListSnippetsFunc != nil {
		return m.ListSnippetsFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateSnippet(snippet *pages.Snippet) error {
	if m.CreateSnippetFunc != nil {
		return m.CreateSnippetFunc(snippet)
	}
	return nil
}

func (m *MockBackend) UpdateSnippet(snippet *pages.Snippet) error {
	if m.UpdateSnippetFunc != nil {
		return m.UpdateSnippetFunc(snippet)
	}
	return nil
}

func (m *MockBackend) DeleteSnippet(id model.ID) error {
	if m.DeleteSnippetFunc != nil {
		return m.DeleteSnippetFunc(id)
	}
	return nil
}

func (m *MockBackend) MoveSnippet(snippet *pages.Snippet) error {
	if m.MoveSnippetFunc != nil {
		return m.MoveSnippetFunc(snippet)
	}
	return nil
}

func (m *MockBackend) ListBuildingBlocks() ([]*pages.BuildingBlock, error) {
	if m.ListBuildingBlocksFunc != nil {
		return m.ListBuildingBlocksFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListPageTemplates() ([]*pages.PageTemplate, error) {
	if m.ListPageTemplatesFunc != nil {
		return m.ListPageTemplatesFunc()
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// EnumerationBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListEnumerations() ([]*model.Enumeration, error) {
	if m.ListEnumerationsFunc != nil {
		return m.ListEnumerationsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetEnumeration(id model.ID) (*model.Enumeration, error) {
	if m.GetEnumerationFunc != nil {
		return m.GetEnumerationFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) CreateEnumeration(enum *model.Enumeration) error {
	if m.CreateEnumerationFunc != nil {
		return m.CreateEnumerationFunc(enum)
	}
	return nil
}

func (m *MockBackend) UpdateEnumeration(enum *model.Enumeration) error {
	if m.UpdateEnumerationFunc != nil {
		return m.UpdateEnumerationFunc(enum)
	}
	return nil
}

func (m *MockBackend) MoveEnumeration(enum *model.Enumeration) error {
	if m.MoveEnumerationFunc != nil {
		return m.MoveEnumerationFunc(enum)
	}
	return nil
}

func (m *MockBackend) DeleteEnumeration(id model.ID) error {
	if m.DeleteEnumerationFunc != nil {
		return m.DeleteEnumerationFunc(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ConstantBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListConstants() ([]*model.Constant, error) {
	if m.ListConstantsFunc != nil {
		return m.ListConstantsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetConstant(id model.ID) (*model.Constant, error) {
	if m.GetConstantFunc != nil {
		return m.GetConstantFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) CreateConstant(constant *model.Constant) error {
	if m.CreateConstantFunc != nil {
		return m.CreateConstantFunc(constant)
	}
	return nil
}

func (m *MockBackend) UpdateConstant(constant *model.Constant) error {
	if m.UpdateConstantFunc != nil {
		return m.UpdateConstantFunc(constant)
	}
	return nil
}

func (m *MockBackend) MoveConstant(constant *model.Constant) error {
	if m.MoveConstantFunc != nil {
		return m.MoveConstantFunc(constant)
	}
	return nil
}

func (m *MockBackend) DeleteConstant(id model.ID) error {
	if m.DeleteConstantFunc != nil {
		return m.DeleteConstantFunc(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// SecurityBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) GetProjectSecurity() (*security.ProjectSecurity, error) {
	if m.GetProjectSecurityFunc != nil {
		return m.GetProjectSecurityFunc()
	}
	return nil, nil
}

func (m *MockBackend) SetProjectSecurityLevel(unitID model.ID, level string) error {
	if m.SetProjectSecurityLevelFunc != nil {
		return m.SetProjectSecurityLevelFunc(unitID, level)
	}
	return nil
}

func (m *MockBackend) SetProjectDemoUsersEnabled(unitID model.ID, enabled bool) error {
	if m.SetProjectDemoUsersEnabledFunc != nil {
		return m.SetProjectDemoUsersEnabledFunc(unitID, enabled)
	}
	return nil
}

func (m *MockBackend) AddUserRole(unitID model.ID, name string, moduleRoles []string, manageAllRoles bool) error {
	if m.AddUserRoleFunc != nil {
		return m.AddUserRoleFunc(unitID, name, moduleRoles, manageAllRoles)
	}
	return nil
}

func (m *MockBackend) AlterUserRoleModuleRoles(unitID model.ID, userRoleName string, add bool, moduleRoles []string) error {
	if m.AlterUserRoleModuleRolesFunc != nil {
		return m.AlterUserRoleModuleRolesFunc(unitID, userRoleName, add, moduleRoles)
	}
	return nil
}

func (m *MockBackend) RemoveUserRole(unitID model.ID, name string) error {
	if m.RemoveUserRoleFunc != nil {
		return m.RemoveUserRoleFunc(unitID, name)
	}
	return nil
}

func (m *MockBackend) AddDemoUser(unitID model.ID, userName, password, entity string, userRoles []string) error {
	if m.AddDemoUserFunc != nil {
		return m.AddDemoUserFunc(unitID, userName, password, entity, userRoles)
	}
	return nil
}

func (m *MockBackend) RemoveDemoUser(unitID model.ID, userName string) error {
	if m.RemoveDemoUserFunc != nil {
		return m.RemoveDemoUserFunc(unitID, userName)
	}
	return nil
}

func (m *MockBackend) ListModuleSecurity() ([]*security.ModuleSecurity, error) {
	if m.ListModuleSecurityFunc != nil {
		return m.ListModuleSecurityFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetModuleSecurity(moduleID model.ID) (*security.ModuleSecurity, error) {
	if m.GetModuleSecurityFunc != nil {
		return m.GetModuleSecurityFunc(moduleID)
	}
	return nil, nil
}

func (m *MockBackend) AddModuleRole(unitID model.ID, roleName, description string) error {
	if m.AddModuleRoleFunc != nil {
		return m.AddModuleRoleFunc(unitID, roleName, description)
	}
	return nil
}

func (m *MockBackend) RemoveModuleRole(unitID model.ID, roleName string) error {
	if m.RemoveModuleRoleFunc != nil {
		return m.RemoveModuleRoleFunc(unitID, roleName)
	}
	return nil
}

func (m *MockBackend) RemoveModuleRoleFromAllUserRoles(unitID model.ID, qualifiedRole string) (int, error) {
	if m.RemoveModuleRoleFromAllUserRolesFunc != nil {
		return m.RemoveModuleRoleFromAllUserRolesFunc(unitID, qualifiedRole)
	}
	return 0, nil
}

func (m *MockBackend) UpdateAllowedRoles(unitID model.ID, roles []string) error {
	if m.UpdateAllowedRolesFunc != nil {
		return m.UpdateAllowedRolesFunc(unitID, roles)
	}
	return nil
}

func (m *MockBackend) UpdatePublishedRestServiceRoles(unitID model.ID, roles []string) error {
	if m.UpdatePublishedRestServiceRolesFunc != nil {
		return m.UpdatePublishedRestServiceRolesFunc(unitID, roles)
	}
	return nil
}

func (m *MockBackend) RemoveFromAllowedRoles(unitID model.ID, roleName string) (bool, error) {
	if m.RemoveFromAllowedRolesFunc != nil {
		return m.RemoveFromAllowedRolesFunc(unitID, roleName)
	}
	return false, nil
}

func (m *MockBackend) AddEntityAccessRule(unitID model.ID, entityName string, roleNames []string, allowCreate, allowDelete bool, defaultMemberAccess string, xpathConstraint string, memberAccesses []mpr.EntityMemberAccess) error {
	if m.AddEntityAccessRuleFunc != nil {
		return m.AddEntityAccessRuleFunc(unitID, entityName, roleNames, allowCreate, allowDelete, defaultMemberAccess, xpathConstraint, memberAccesses)
	}
	return nil
}

func (m *MockBackend) RemoveEntityAccessRule(unitID model.ID, entityName string, roleNames []string) (int, error) {
	if m.RemoveEntityAccessRuleFunc != nil {
		return m.RemoveEntityAccessRuleFunc(unitID, entityName, roleNames)
	}
	return 0, nil
}

func (m *MockBackend) RevokeEntityMemberAccess(unitID model.ID, entityName string, roleNames []string, revocation mpr.EntityAccessRevocation) (int, error) {
	if m.RevokeEntityMemberAccessFunc != nil {
		return m.RevokeEntityMemberAccessFunc(unitID, entityName, roleNames, revocation)
	}
	return 0, nil
}

func (m *MockBackend) RemoveRoleFromAllEntities(unitID model.ID, roleName string) (int, error) {
	if m.RemoveRoleFromAllEntitiesFunc != nil {
		return m.RemoveRoleFromAllEntitiesFunc(unitID, roleName)
	}
	return 0, nil
}

func (m *MockBackend) ReconcileMemberAccesses(unitID model.ID, moduleName string) (int, error) {
	if m.ReconcileMemberAccessesFunc != nil {
		return m.ReconcileMemberAccessesFunc(unitID, moduleName)
	}
	return 0, nil
}

// ---------------------------------------------------------------------------
// NavigationBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListNavigationDocuments() ([]*mpr.NavigationDocument, error) {
	if m.ListNavigationDocumentsFunc != nil {
		return m.ListNavigationDocumentsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetNavigation() (*mpr.NavigationDocument, error) {
	if m.GetNavigationFunc != nil {
		return m.GetNavigationFunc()
	}
	return nil, nil
}

func (m *MockBackend) UpdateNavigationProfile(navDocID model.ID, profileName string, spec mpr.NavigationProfileSpec) error {
	if m.UpdateNavigationProfileFunc != nil {
		return m.UpdateNavigationProfileFunc(navDocID, profileName, spec)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ServiceBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListConsumedODataServices() ([]*model.ConsumedODataService, error) {
	if m.ListConsumedODataServicesFunc != nil {
		return m.ListConsumedODataServicesFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListPublishedODataServices() ([]*model.PublishedODataService, error) {
	if m.ListPublishedODataServicesFunc != nil {
		return m.ListPublishedODataServicesFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateConsumedODataService(svc *model.ConsumedODataService) error {
	if m.CreateConsumedODataServiceFunc != nil {
		return m.CreateConsumedODataServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) UpdateConsumedODataService(svc *model.ConsumedODataService) error {
	if m.UpdateConsumedODataServiceFunc != nil {
		return m.UpdateConsumedODataServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) DeleteConsumedODataService(id model.ID) error {
	if m.DeleteConsumedODataServiceFunc != nil {
		return m.DeleteConsumedODataServiceFunc(id)
	}
	return nil
}

func (m *MockBackend) CreatePublishedODataService(svc *model.PublishedODataService) error {
	if m.CreatePublishedODataServiceFunc != nil {
		return m.CreatePublishedODataServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) UpdatePublishedODataService(svc *model.PublishedODataService) error {
	if m.UpdatePublishedODataServiceFunc != nil {
		return m.UpdatePublishedODataServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) DeletePublishedODataService(id model.ID) error {
	if m.DeletePublishedODataServiceFunc != nil {
		return m.DeletePublishedODataServiceFunc(id)
	}
	return nil
}

func (m *MockBackend) ListConsumedRestServices() ([]*model.ConsumedRestService, error) {
	if m.ListConsumedRestServicesFunc != nil {
		return m.ListConsumedRestServicesFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListPublishedRestServices() ([]*model.PublishedRestService, error) {
	if m.ListPublishedRestServicesFunc != nil {
		return m.ListPublishedRestServicesFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateConsumedRestService(svc *model.ConsumedRestService) error {
	if m.CreateConsumedRestServiceFunc != nil {
		return m.CreateConsumedRestServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) UpdateConsumedRestService(svc *model.ConsumedRestService) error {
	if m.UpdateConsumedRestServiceFunc != nil {
		return m.UpdateConsumedRestServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) DeleteConsumedRestService(id model.ID) error {
	if m.DeleteConsumedRestServiceFunc != nil {
		return m.DeleteConsumedRestServiceFunc(id)
	}
	return nil
}

func (m *MockBackend) CreatePublishedRestService(svc *model.PublishedRestService) error {
	if m.CreatePublishedRestServiceFunc != nil {
		return m.CreatePublishedRestServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) UpdatePublishedRestService(svc *model.PublishedRestService) error {
	if m.UpdatePublishedRestServiceFunc != nil {
		return m.UpdatePublishedRestServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) DeletePublishedRestService(id model.ID) error {
	if m.DeletePublishedRestServiceFunc != nil {
		return m.DeletePublishedRestServiceFunc(id)
	}
	return nil
}

func (m *MockBackend) ListBusinessEventServices() ([]*model.BusinessEventService, error) {
	if m.ListBusinessEventServicesFunc != nil {
		return m.ListBusinessEventServicesFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateBusinessEventService(svc *model.BusinessEventService) error {
	if m.CreateBusinessEventServiceFunc != nil {
		return m.CreateBusinessEventServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) UpdateBusinessEventService(svc *model.BusinessEventService) error {
	if m.UpdateBusinessEventServiceFunc != nil {
		return m.UpdateBusinessEventServiceFunc(svc)
	}
	return nil
}

func (m *MockBackend) DeleteBusinessEventService(id model.ID) error {
	if m.DeleteBusinessEventServiceFunc != nil {
		return m.DeleteBusinessEventServiceFunc(id)
	}
	return nil
}

func (m *MockBackend) ListDatabaseConnections() ([]*model.DatabaseConnection, error) {
	if m.ListDatabaseConnectionsFunc != nil {
		return m.ListDatabaseConnectionsFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateDatabaseConnection(conn *model.DatabaseConnection) error {
	if m.CreateDatabaseConnectionFunc != nil {
		return m.CreateDatabaseConnectionFunc(conn)
	}
	return nil
}

func (m *MockBackend) UpdateDatabaseConnection(conn *model.DatabaseConnection) error {
	if m.UpdateDatabaseConnectionFunc != nil {
		return m.UpdateDatabaseConnectionFunc(conn)
	}
	return nil
}

func (m *MockBackend) MoveDatabaseConnection(conn *model.DatabaseConnection) error {
	if m.MoveDatabaseConnectionFunc != nil {
		return m.MoveDatabaseConnectionFunc(conn)
	}
	return nil
}

func (m *MockBackend) DeleteDatabaseConnection(id model.ID) error {
	if m.DeleteDatabaseConnectionFunc != nil {
		return m.DeleteDatabaseConnectionFunc(id)
	}
	return nil
}

func (m *MockBackend) ListDataTransformers() ([]*model.DataTransformer, error) {
	if m.ListDataTransformersFunc != nil {
		return m.ListDataTransformersFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateDataTransformer(dt *model.DataTransformer) error {
	if m.CreateDataTransformerFunc != nil {
		return m.CreateDataTransformerFunc(dt)
	}
	return nil
}

func (m *MockBackend) DeleteDataTransformer(id model.ID) error {
	if m.DeleteDataTransformerFunc != nil {
		return m.DeleteDataTransformerFunc(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// MappingBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListImportMappings() ([]*model.ImportMapping, error) {
	if m.ListImportMappingsFunc != nil {
		return m.ListImportMappingsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetImportMappingByQualifiedName(moduleName, name string) (*model.ImportMapping, error) {
	if m.GetImportMappingByQualifiedNameFunc != nil {
		return m.GetImportMappingByQualifiedNameFunc(moduleName, name)
	}
	return nil, nil
}

func (m *MockBackend) CreateImportMapping(im *model.ImportMapping) error {
	if m.CreateImportMappingFunc != nil {
		return m.CreateImportMappingFunc(im)
	}
	return nil
}

func (m *MockBackend) UpdateImportMapping(im *model.ImportMapping) error {
	if m.UpdateImportMappingFunc != nil {
		return m.UpdateImportMappingFunc(im)
	}
	return nil
}

func (m *MockBackend) DeleteImportMapping(id model.ID) error {
	if m.DeleteImportMappingFunc != nil {
		return m.DeleteImportMappingFunc(id)
	}
	return nil
}

func (m *MockBackend) MoveImportMapping(im *model.ImportMapping) error {
	if m.MoveImportMappingFunc != nil {
		return m.MoveImportMappingFunc(im)
	}
	return nil
}

func (m *MockBackend) ListExportMappings() ([]*model.ExportMapping, error) {
	if m.ListExportMappingsFunc != nil {
		return m.ListExportMappingsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetExportMappingByQualifiedName(moduleName, name string) (*model.ExportMapping, error) {
	if m.GetExportMappingByQualifiedNameFunc != nil {
		return m.GetExportMappingByQualifiedNameFunc(moduleName, name)
	}
	return nil, nil
}

func (m *MockBackend) CreateExportMapping(em *model.ExportMapping) error {
	if m.CreateExportMappingFunc != nil {
		return m.CreateExportMappingFunc(em)
	}
	return nil
}

func (m *MockBackend) UpdateExportMapping(em *model.ExportMapping) error {
	if m.UpdateExportMappingFunc != nil {
		return m.UpdateExportMappingFunc(em)
	}
	return nil
}

func (m *MockBackend) DeleteExportMapping(id model.ID) error {
	if m.DeleteExportMappingFunc != nil {
		return m.DeleteExportMappingFunc(id)
	}
	return nil
}

func (m *MockBackend) MoveExportMapping(em *model.ExportMapping) error {
	if m.MoveExportMappingFunc != nil {
		return m.MoveExportMappingFunc(em)
	}
	return nil
}

func (m *MockBackend) ListJsonStructures() ([]*mpr.JsonStructure, error) {
	if m.ListJsonStructuresFunc != nil {
		return m.ListJsonStructuresFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetJsonStructureByQualifiedName(moduleName, name string) (*mpr.JsonStructure, error) {
	if m.GetJsonStructureByQualifiedNameFunc != nil {
		return m.GetJsonStructureByQualifiedNameFunc(moduleName, name)
	}
	return nil, nil
}

func (m *MockBackend) CreateJsonStructure(js *mpr.JsonStructure) error {
	if m.CreateJsonStructureFunc != nil {
		return m.CreateJsonStructureFunc(js)
	}
	return nil
}

func (m *MockBackend) DeleteJsonStructure(id string) error {
	if m.DeleteJsonStructureFunc != nil {
		return m.DeleteJsonStructureFunc(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// JavaBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListJavaActions() ([]*mpr.JavaAction, error) {
	if m.ListJavaActionsFunc != nil {
		return m.ListJavaActionsFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListJavaScriptActions() ([]*mpr.JavaScriptAction, error) {
	if m.ListJavaScriptActionsFunc != nil {
		return m.ListJavaScriptActionsFunc()
	}
	return nil, nil
}

func (m *MockBackend) ReadJavaActionByName(qualifiedName string) (*javaactions.JavaAction, error) {
	if m.ReadJavaActionByNameFunc != nil {
		return m.ReadJavaActionByNameFunc(qualifiedName)
	}
	return nil, nil
}

func (m *MockBackend) ReadJavaScriptActionByName(qualifiedName string) (*mpr.JavaScriptAction, error) {
	if m.ReadJavaScriptActionByNameFunc != nil {
		return m.ReadJavaScriptActionByNameFunc(qualifiedName)
	}
	return nil, nil
}

func (m *MockBackend) CreateJavaAction(ja *javaactions.JavaAction) error {
	if m.CreateJavaActionFunc != nil {
		return m.CreateJavaActionFunc(ja)
	}
	return nil
}

func (m *MockBackend) UpdateJavaAction(ja *javaactions.JavaAction) error {
	if m.UpdateJavaActionFunc != nil {
		return m.UpdateJavaActionFunc(ja)
	}
	return nil
}

func (m *MockBackend) DeleteJavaAction(id model.ID) error {
	if m.DeleteJavaActionFunc != nil {
		return m.DeleteJavaActionFunc(id)
	}
	return nil
}

func (m *MockBackend) WriteJavaSourceFile(moduleName, actionName string, javaCode string, params []*javaactions.JavaActionParameter, returnType javaactions.CodeActionReturnType) error {
	if m.WriteJavaSourceFileFunc != nil {
		return m.WriteJavaSourceFileFunc(moduleName, actionName, javaCode, params, returnType)
	}
	return nil
}

func (m *MockBackend) ReadJavaSourceFile(moduleName, actionName string) (string, error) {
	if m.ReadJavaSourceFileFunc != nil {
		return m.ReadJavaSourceFileFunc(moduleName, actionName)
	}
	return "", nil
}

// ---------------------------------------------------------------------------
// WorkflowBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListWorkflows() ([]*workflows.Workflow, error) {
	if m.ListWorkflowsFunc != nil {
		return m.ListWorkflowsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetWorkflow(id model.ID) (*workflows.Workflow, error) {
	if m.GetWorkflowFunc != nil {
		return m.GetWorkflowFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) CreateWorkflow(wf *workflows.Workflow) error {
	if m.CreateWorkflowFunc != nil {
		return m.CreateWorkflowFunc(wf)
	}
	return nil
}

func (m *MockBackend) DeleteWorkflow(id model.ID) error {
	if m.DeleteWorkflowFunc != nil {
		return m.DeleteWorkflowFunc(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// SettingsBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) GetProjectSettings() (*model.ProjectSettings, error) {
	if m.GetProjectSettingsFunc != nil {
		return m.GetProjectSettingsFunc()
	}
	return nil, nil
}

func (m *MockBackend) UpdateProjectSettings(ps *model.ProjectSettings) error {
	if m.UpdateProjectSettingsFunc != nil {
		return m.UpdateProjectSettingsFunc(ps)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ImageBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListImageCollections() ([]*mpr.ImageCollection, error) {
	if m.ListImageCollectionsFunc != nil {
		return m.ListImageCollectionsFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateImageCollection(ic *mpr.ImageCollection) error {
	if m.CreateImageCollectionFunc != nil {
		return m.CreateImageCollectionFunc(ic)
	}
	return nil
}

func (m *MockBackend) DeleteImageCollection(id string) error {
	if m.DeleteImageCollectionFunc != nil {
		return m.DeleteImageCollectionFunc(id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ScheduledEventBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListScheduledEvents() ([]*model.ScheduledEvent, error) {
	if m.ListScheduledEventsFunc != nil {
		return m.ListScheduledEventsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetScheduledEvent(id model.ID) (*model.ScheduledEvent, error) {
	if m.GetScheduledEventFunc != nil {
		return m.GetScheduledEventFunc(id)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// RenameBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) UpdateQualifiedNameInAllUnits(oldName, newName string) (int, error) {
	if m.UpdateQualifiedNameInAllUnitsFunc != nil {
		return m.UpdateQualifiedNameInAllUnitsFunc(oldName, newName)
	}
	return 0, nil
}

func (m *MockBackend) RenameReferences(oldName, newName string, dryRun bool) ([]mpr.RenameHit, error) {
	if m.RenameReferencesFunc != nil {
		return m.RenameReferencesFunc(oldName, newName, dryRun)
	}
	return nil, nil
}

func (m *MockBackend) RenameDocumentByName(moduleName, oldName, newName string) error {
	if m.RenameDocumentByNameFunc != nil {
		return m.RenameDocumentByNameFunc(moduleName, oldName, newName)
	}
	return nil
}

// ---------------------------------------------------------------------------
// RawUnitBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) GetRawUnit(id model.ID) (map[string]any, error) {
	if m.GetRawUnitFunc != nil {
		return m.GetRawUnitFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) GetRawUnitBytes(id string) ([]byte, error) {
	if m.GetRawUnitBytesFunc != nil {
		return m.GetRawUnitBytesFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) ListRawUnitsByType(typePrefix string) ([]*mpr.RawUnit, error) {
	if m.ListRawUnitsByTypeFunc != nil {
		return m.ListRawUnitsByTypeFunc(typePrefix)
	}
	return nil, nil
}

func (m *MockBackend) ListRawUnits(objectType string) ([]*mpr.RawUnitInfo, error) {
	if m.ListRawUnitsFunc != nil {
		return m.ListRawUnitsFunc(objectType)
	}
	return nil, nil
}

func (m *MockBackend) GetRawUnitByName(objectType, qualifiedName string) (*mpr.RawUnitInfo, error) {
	if m.GetRawUnitByNameFunc != nil {
		return m.GetRawUnitByNameFunc(objectType, qualifiedName)
	}
	return nil, nil
}

func (m *MockBackend) GetRawMicroflowByName(qualifiedName string) ([]byte, error) {
	if m.GetRawMicroflowByNameFunc != nil {
		return m.GetRawMicroflowByNameFunc(qualifiedName)
	}
	return nil, nil
}

func (m *MockBackend) UpdateRawUnit(unitID string, contents []byte) error {
	if m.UpdateRawUnitFunc != nil {
		return m.UpdateRawUnitFunc(unitID, contents)
	}
	return nil
}

// ---------------------------------------------------------------------------
// MetadataBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListAllUnitIDs() ([]string, error) {
	if m.ListAllUnitIDsFunc != nil {
		return m.ListAllUnitIDsFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListUnits() ([]*mpr.UnitInfo, error) {
	if m.ListUnitsFunc != nil {
		return m.ListUnitsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetUnitTypes() (map[string]int, error) {
	if m.GetUnitTypesFunc != nil {
		return m.GetUnitTypesFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetProjectRootID() (string, error) {
	if m.GetProjectRootIDFunc != nil {
		return m.GetProjectRootIDFunc()
	}
	return "", nil
}

func (m *MockBackend) ContentsDir() string {
	if m.ContentsDirFunc != nil {
		return m.ContentsDirFunc()
	}
	return ""
}

func (m *MockBackend) ExportJSON() ([]byte, error) {
	if m.ExportJSONFunc != nil {
		return m.ExportJSONFunc()
	}
	return nil, nil
}

func (m *MockBackend) InvalidateCache() {
	if m.InvalidateCacheFunc != nil {
		m.InvalidateCacheFunc()
	}
}

// ---------------------------------------------------------------------------
// WidgetBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) FindCustomWidgetType(widgetID string) (*mpr.RawCustomWidgetType, error) {
	if m.FindCustomWidgetTypeFunc != nil {
		return m.FindCustomWidgetTypeFunc(widgetID)
	}
	return nil, nil
}

func (m *MockBackend) FindAllCustomWidgetTypes(widgetID string) ([]*mpr.RawCustomWidgetType, error) {
	if m.FindAllCustomWidgetTypesFunc != nil {
		return m.FindAllCustomWidgetTypesFunc(widgetID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// AgentEditorBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListAgentEditorModels() ([]*agenteditor.Model, error) {
	if m.ListAgentEditorModelsFunc != nil {
		return m.ListAgentEditorModelsFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListAgentEditorKnowledgeBases() ([]*agenteditor.KnowledgeBase, error) {
	if m.ListAgentEditorKnowledgeBasesFunc != nil {
		return m.ListAgentEditorKnowledgeBasesFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListAgentEditorConsumedMCPServices() ([]*agenteditor.ConsumedMCPService, error) {
	if m.ListAgentEditorConsumedMCPServicesFunc != nil {
		return m.ListAgentEditorConsumedMCPServicesFunc()
	}
	return nil, nil
}

func (m *MockBackend) ListAgentEditorAgents() ([]*agenteditor.Agent, error) {
	if m.ListAgentEditorAgentsFunc != nil {
		return m.ListAgentEditorAgentsFunc()
	}
	return nil, nil
}
