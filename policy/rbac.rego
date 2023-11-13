package app.rbac

import future.keywords.contains
import future.keywords.if
import future.keywords.in

default allow := false

allow if {
	some permission in permissions
	input.domain == permission.domain

	some action in permission.actions
	input.action == action
}

permissions contains permission if {
	some role in input.roles
	some permission in data.role_permissions[role]
}
