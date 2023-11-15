package app.rbac

import future.keywords.contains
import future.keywords.if
import future.keywords.in

default allow := false

allow if {
	is_admin
	is_authorized
}

allow if {
	is_owner
	is_authorized
}

allow if {
	is_authorized
	count(authorized_users) == 0
}

is_authorized if {
	some permission in permissions
	input.domain == permission.domain

	some action in permission.actions
	input.action == action
}

is_admin if {
	some user in authorized_users
	user == "admin"

	some role in input.claims.roles
	role == user
}

is_owner if {
	some user in authorized_users
	user == "owner"
	input.object == input.claims.sub
}

permissions contains permission if {
	some role in input.claims.roles
	some permission in data.role_permissions[role]
}

authorized_users contains who if {
	some who, flag in data.who_enum
	bits.and(input.who_flags, flag) > 0
}
