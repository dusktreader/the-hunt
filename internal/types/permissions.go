package types

import (
	"github.com/hashicorp/go-set/v3"
)

type PermCode string

const (
	CompanyRead		PermCode = "companies:read"
	CompanyWrite	PermCode = "companies:write"
	UserRead		PermCode = "users:read"
	UserWrite		PermCode = "users:write"
)

type PermissionSet = set.Set[PermCode]

func NewPermissionSet(perms ...PermCode) *PermissionSet {
	return set.From(perms)
}

func HasPerms(ps *PermissionSet, strategy PermissionStrategy, perms ...PermCode) bool {
	in := set.From(perms)
	if strategy == All {
		return ps.Subset(in)
	} else {
		return ps.Intersect(in).Size() > 0
	}
}

type PermissionStrategy string
const All PermissionStrategy = "all"
const Some PermissionStrategy = "some"
