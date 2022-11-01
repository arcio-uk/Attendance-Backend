package security

import (
	"math/bits"
)

// Layers
type Layer int

const (
	Global      Layer = iota
	Module      Layer = iota
	ModuleGroup Layer = iota
	Attendance  Layer = iota
)

func (l Layer) IsValid() bool {
	switch l {
	case Global, Module, ModuleGroup, Attendance:
		return true
	}
	return false
}

type Overrides uint32

// See https://docs.arcio.uk/attendance-system/perms for more detail
// Permission consts
const PERMS_NONE = 0 // No permissions at all

// CRUD Self (current level)
const PERMS_CREATE_SELF = 1 << 0
const PERMS_READ_SELF = 1 << 1 // read entries in a layer that involve you
const PERMS_UPDATE_SELF = 1 << 2
const PERMS_DELETE_SELF = 1 << 3

// CRUD Children (lower levels)
// Applies to Global and, Module
const PERMS_CREATE_CHILDREN = 1 << 4
const PERMS_READ_CHILDREN = 1 << 5 // read entries in lower levels that involve you
const PERMS_UPDATE_CHILDREN = 1 << 6
const PERMS_DELETE_CHILDREN = 1 << 7

// Entries for reading all entries in the layer, you can only update and, delete things that you can see
const PERMS_READ_ALL_SELF = 1 << 8
const PERMS_READ_ALL_CHILDREN = 1 << 9

// Roles management
const PERMS_MANAGE_ROLES = 1 << 10

// Attendance permissions
const PERMS_ATTENDANCE_ALLOW_PAST_MARK = 1 << 11 // whether a user can change historical records

const PERMS_VALID_MASK = 0b11111111111 // 11 ones I hope lmao

// This is a bit mask to ignore the self mask of previous layers when calculating permissions
const PERMS_CHILDREN_MASK = 0xFFFFFFFF ^ PERMS_CREATE_SELF ^ PERMS_READ_SELF ^ PERMS_UPDATE_SELF ^ PERMS_DELETE_SELF ^ PERMS_READ_ALL_SELF

// Permission utils
func CalculatePermissions(currentLayer Overrides, prevLayers []Overrides) Overrides {
	var ret Overrides
	for i := 0; i < len(prevLayers); i++ {
		ret |= prevLayers[i]
	}

	return (ret & PERMS_CHILDREN_MASK) | currentLayer
}

func CalculatePermissionsInner(ovr []Overrides) Overrides {
	var ret Overrides
	for i := 0; i < len(ovr); i++ {
		ret |= ovr[i]
	}

	return ret
}

func IsPrivelledgeEscalated(OldPerms Overrides, NewPerms Overrides) bool {
	tmp := OldPerms | NewPerms

	return bits.OnesCount64(uint64(tmp)) > bits.OnesCount64(uint64(OldPerms))
}

// Permissions checkers
// Composite permissions strings for each CRUD to check for self and, children perms at once
const PERMS_CAN_CREATE = PERMS_CREATE_SELF | PERMS_CREATE_CHILDREN
const PERMS_CAN_READ_ALL = PERMS_READ_ALL_SELF | PERMS_READ_ALL_CHILDREN
const PERMS_CAN_READ = PERMS_READ_SELF | PERMS_READ_CHILDREN | PERMS_CAN_READ_ALL
const PERMS_CAN_UPDATE = PERMS_UPDATE_SELF | PERMS_UPDATE_CHILDREN

/*
 * Checks that user's permissions match the permission strings provided
 * userPerms: the permissions of the user
 * ...requiredPerms: is a list of bit strings that must be matched. A bit string can have multple flags
 * if at least one of them must be matched
 * the logic is return userPerms && requiredPerms[0] && requiredPerms[1] && ...
 */
func checkPerms(userPerms Overrides, requiredPerms []Overrides) bool {
	ret := true

	for i := 0; i < len(requiredPerms); i++ {
		ret = ret && userPerms&requiredPerms[i] != 0
	}

	return ret
}

/*
 * Checks that user's permissions match the permission strings provided and, the read permission
 * userPerms: the permissions of the user
 * ...requiredPerms: is a list of bit strings that must be matched. A bit string can have multple flags
 * if at least one of them must be matched
 * the logic is return userPerms && requiredPerms[0] && requiredPerms[1] && ...
 */
func CheckPerms(userPerms Overrides, requiredPerms ...Overrides) bool {
	return checkPerms(userPerms, append(requiredPerms, PERMS_CAN_READ))
}
