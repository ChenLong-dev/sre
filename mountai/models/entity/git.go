package entity

// GitMemberAccess 用户权限
type GitMemberAccess int

const (
	// GitMemberAccessNoAccess ...
	GitMemberAccessNoAccess GitMemberAccess = 0
	// GitMemberAccessGuest ...
	GitMemberAccessGuest GitMemberAccess = 10
	// GitMemberAccessReporter ...
	GitMemberAccessReporter GitMemberAccess = 20
	// GitMemberAccessDeveloper ...
	GitMemberAccessDeveloper GitMemberAccess = 30
	// GitMemberAccessMaintainer ...
	GitMemberAccessMaintainer GitMemberAccess = 40
	// GitMemberAccessOwner ...
	GitMemberAccessOwner GitMemberAccess = 50
)

var GitMemberAccessRoleMap = map[GitMemberAccess]string{
	GitMemberAccessNoAccess:   "NoAccess",
	GitMemberAccessGuest:      "Guest",
	GitMemberAccessReporter:   "Reporter",
	GitMemberAccessDeveloper:  "Developer",
	GitMemberAccessMaintainer: "Maintainer",
	GitMemberAccessOwner:      "Owner",
}

// GitMemberState git 用户的状态
type GitMemberState string

const (
	// GitMemberStateActive ...
	GitMemberStateActive GitMemberState = "active"
)
