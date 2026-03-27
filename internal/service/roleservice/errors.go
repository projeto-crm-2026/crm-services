package roleservice

import "errors"

var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrRoleNameTaken          = errors.New("a role with this name already exists")
	ErrCannotModifySystemRole = errors.New("system roles cannot be modified")
	ErrCannotDeleteSystemRole = errors.New("system roles cannot be deleted")
)
