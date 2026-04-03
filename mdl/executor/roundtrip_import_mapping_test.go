// SPDX-License-Identifier: Apache-2.0

//go:build integration

package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestRoundtripImportMapping_NoSchema(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	if err := env.executeMDL(`CREATE ENTITY ` + testModule + `.IMPet (
  PetId: Integer,
  Name: String(200)
);`); err != nil {
		t.Fatalf("CREATE ENTITY failed: %v", err)
	}

	createMDL := `CREATE IMPORT MAPPING ` + testModule + `.ImportPetBasic {
  CREATE ` + testModule + `.IMPet {
    PetId = id KEY,
    Name = name
  }
};`

	env.assertContains(createMDL, []string{
		"IMPORT MAPPING",
		"ImportPetBasic",
		"IMPet",
		"CREATE",
	})
}

func TestRoundtripImportMapping_WithJsonStructureRef(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	if err := env.executeMDL(`CREATE ENTITY ` + testModule + `.IMOrder (
  OrderId: Integer,
  Total: Decimal
);`); err != nil {
		t.Fatalf("CREATE ENTITY failed: %v", err)
	}

	if err := env.executeMDL(`CREATE JSON STRUCTURE ` + testModule + `.OrderJS
SNIPPET '{"orderId": 1, "total": 99.99}';`); err != nil {
		t.Fatalf("CREATE JSON STRUCTURE failed: %v", err)
	}

	createMDL := `CREATE IMPORT MAPPING ` + testModule + `.ImportOrder
  WITH JSON STRUCTURE ` + testModule + `.OrderJS
{
  CREATE ` + testModule + `.IMOrder {
    OrderId = orderId KEY,
    Total = total
  }
};`

	env.assertContains(createMDL, []string{
		"IMPORT MAPPING",
		"ImportOrder",
		"WITH JSON STRUCTURE",
		"IMOrder",
		"orderId",
		"total",
	})
}

func TestRoundtripImportMapping_ValueTypes(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	if err := env.executeMDL(`CREATE ENTITY ` + testModule + `.IMAllTypes (
  IntVal: Integer,
  DecVal: Decimal,
  BoolVal: Boolean DEFAULT false,
  DateVal: DateTime
);`); err != nil {
		t.Fatalf("CREATE ENTITY failed: %v", err)
	}

	createMDL := `CREATE IMPORT MAPPING ` + testModule + `.ImportAllTypes {
  CREATE ` + testModule + `.IMAllTypes {
    IntVal = intVal KEY,
    DecVal = decVal,
    BoolVal = boolVal,
    DateVal = dateVal
  }
};`

	env.assertContains(createMDL, []string{
		"IntVal",
		"DecVal",
		"BoolVal",
		"DateVal",
	})
}

func TestRoundtripImportMapping_Drop(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	if err := env.executeMDL(`CREATE ENTITY ` + testModule + `.IMDropPet (
  PetId: Integer
);`); err != nil {
		t.Fatalf("CREATE ENTITY failed: %v", err)
	}

	if err := env.executeMDL(`CREATE IMPORT MAPPING ` + testModule + `.ToDropIM {
  CREATE ` + testModule + `.IMDropPet {
    PetId = id KEY
  }
};`); err != nil {
		t.Fatalf("CREATE IMPORT MAPPING failed: %v", err)
	}

	if _, err := env.describeMDL(`DESCRIBE IMPORT MAPPING ` + testModule + `.ToDropIM;`); err != nil {
		t.Fatalf("import mapping should exist before DROP: %v", err)
	}

	if err := env.executeMDL(`DROP IMPORT MAPPING ` + testModule + `.ToDropIM;`); err != nil {
		t.Fatalf("DROP IMPORT MAPPING failed: %v", err)
	}

	if _, err := env.describeMDL(`DESCRIBE IMPORT MAPPING ` + testModule + `.ToDropIM;`); err == nil {
		t.Error("import mapping should not exist after DROP")
	}
}

func TestRoundtripImportMapping_ShowAppearsInList(t *testing.T) {
	env := setupTestEnv(t)
	defer env.teardown()

	if err := env.executeMDL(`CREATE ENTITY ` + testModule + `.IMListPet (
  PetId: Integer
);`); err != nil {
		t.Fatalf("CREATE ENTITY failed: %v", err)
	}

	if err := env.executeMDL(`CREATE IMPORT MAPPING ` + testModule + `.ListableIM {
  CREATE ` + testModule + `.IMListPet {
    PetId = id KEY
  }
};`); err != nil {
		t.Fatalf("CREATE IMPORT MAPPING failed: %v", err)
	}

	env.output.Reset()
	if err := env.executeMDL(`SHOW IMPORT MAPPINGS IN ` + testModule + `;`); err != nil {
		t.Fatalf("SHOW failed: %v", err)
	}

	if !strings.Contains(env.output.String(), "ListableIM") {
		t.Errorf("expected 'ListableIM' in SHOW output:\n%s", env.output.String())
	}
}

// --- MX Check ---

func TestMxCheck_ImportMapping_Basic(t *testing.T) {
	if !mxCheckAvailable() {
		t.Skip("mx command not available")
	}

	env := setupTestEnv(t)
	defer env.teardown()

	if err := env.executeMDL(`CREATE ENTITY ` + testModule + `.MxCheckIMPet (
  PetId: Integer,
  Name: String(200)
);`); err != nil {
		t.Fatalf("CREATE ENTITY failed: %v", err)
	}

	if err := env.executeMDL(`CREATE JSON STRUCTURE ` + testModule + `.MxCheckIMJS
SNIPPET '{"id": 1, "name": "Fido"}';`); err != nil {
		t.Fatalf("CREATE JSON STRUCTURE failed: %v", err)
	}

	if err := env.executeMDL(`CREATE IMPORT MAPPING ` + testModule + `.MxCheckImportPet
  WITH JSON STRUCTURE ` + testModule + `.MxCheckIMJS
{
  CREATE ` + testModule + `.MxCheckIMPet {
    PetId = id,
    Name = name
  }
};`); err != nil {
		t.Fatalf("CREATE IMPORT MAPPING failed: %v", err)
	}

	env.executor.Execute(&ast.DisconnectStmt{})

	output, err := runMxCheck(t, env.projectPath)
	assertMxCheckPassed(t, output, err)
}
