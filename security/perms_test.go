package security

import (
	"testing"
)

func LPEBaseCaseTest(t *testing.T) {
	var OldPerms Overrides = PERMS_NONE
	var NewPerms Overrides = PERMS_CREATE_SELF | PERMS_CREATE_CHILDREN

	// Privs are escalated
	res := IsPrivelledgeEscalated(OldPerms, NewPerms)
	if !res {
		t.Fail()
	}
}

func LPEEqualCaseTest(t *testing.T) {
	var OldPerms Overrides = PERMS_NONE
	var NewPerms Overrides = PERMS_NONE

	// Privs are not escalated
	res := IsPrivelledgeEscalated(OldPerms, NewPerms)
	if res {
		t.Fail()
	}
}

func LPEGeneralCase1Test(t *testing.T) {
	var OldPerms Overrides = PERMS_CREATE_SELF | PERMS_CREATE_CHILDREN
	var NewPerms Overrides = PERMS_NONE

	// Privs are not escalated
	res := IsPrivelledgeEscalated(OldPerms, NewPerms)
	if res {
		t.Fail()
	}
}

func LPEGeneralCase2Test(t *testing.T) {
	var OldPerms Overrides = PERMS_CREATE_SELF
	var NewPerms Overrides = PERMS_CREATE_SELF | PERMS_CREATE_CHILDREN

	// Privs are escalated
	res := IsPrivelledgeEscalated(OldPerms, NewPerms)
	if !res {
		t.Fail()
	}
}

func LPEGeneralCase3Test(t *testing.T) {
	var OldPerms Overrides = PERMS_CREATE_SELF | PERMS_READ_SELF
	var NewPerms Overrides = PERMS_CREATE_SELF | PERMS_CREATE_CHILDREN

	// Privs are escalated
	res := IsPrivelledgeEscalated(OldPerms, NewPerms)
	if !res {
		t.Fail()
	}
}

func LPEAdminCase1Test(t *testing.T) {
	var OldPerms Overrides = PERMS_CREATE_SELF | PERMS_CREATE_CHILDREN
	var NewPerms Overrides = PERMS_CREATE_SELF | PERMS_CREATE_CHILDREN

	// Privs are escalated
	res := IsPrivelledgeEscalated(OldPerms, NewPerms)
	if res {
		t.Fail()
	}
}

func CalcPermsBaseTest(t *testing.T) {
	var currentLayer Overrides = PERMS_NONE
	var prevLayers = []Overrides{PERMS_READ_CHILDREN}

	ret := CalculatePermissions(currentLayer, prevLayers)
	if ret != PERMS_READ_CHILDREN {
		t.Fail()
	}
}

func CalcPermsNoSelfCascadeTest(t *testing.T) {
	var currentLayer Overrides = PERMS_NONE
	var prevLayers = []Overrides{PERMS_CREATE_SELF | PERMS_READ_SELF | PERMS_UPDATE_SELF | PERMS_DELETE_SELF | PERMS_READ_ALL_SELF}

	ret := CalculatePermissions(currentLayer, prevLayers)
	if ret != PERMS_CREATE_SELF|PERMS_READ_SELF|PERMS_UPDATE_SELF|PERMS_DELETE_SELF|PERMS_READ_ALL_SELF {
		t.Fail()
	}
}

func CalcPermsInnerTest(t *testing.T) {
	var currentLayer = []Overrides{1, 2, 4, 8, 16}

	ret := CalculatePermissionsInner(currentLayer)
	if ret != 1|2|4|8|16 {
		t.Fail()
	}
}
