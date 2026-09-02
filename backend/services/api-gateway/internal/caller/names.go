package caller

import identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"

var organizationStateName = map[identityv1.OrganizationState]string{
	identityv1.OrganizationState_ORGANIZATION_STATE_ACTIVE:     "active",
	identityv1.OrganizationState_ORGANIZATION_STATE_RESTRICTED: "restricted",
	identityv1.OrganizationState_ORGANIZATION_STATE_READ_ONLY:  "read_only",
	identityv1.OrganizationState_ORGANIZATION_STATE_SUSPENDED:  "suspended",
}

var verificationStatusName = map[identityv1.VerificationStatus]string{
	identityv1.VerificationStatus_VERIFICATION_STATUS_VERIFIED:   "verified",
	identityv1.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED: "unverified",
	identityv1.VerificationStatus_VERIFICATION_STATUS_REJECTED:   "rejected",
}

var organizationRoleName = map[identityv1.OrganizationRole]string{
	identityv1.OrganizationRole_ORGANIZATION_ROLE_OWNER:  "owner",
	identityv1.OrganizationRole_ORGANIZATION_ROLE_ADMIN:  "admin",
	identityv1.OrganizationRole_ORGANIZATION_ROLE_MEMBER: "member",
}

var platformRoleName = map[identityv1.PlatformRole]string{
	identityv1.PlatformRole_PLATFORM_ROLE_VERIFIER:       "verifier",
	identityv1.PlatformRole_PLATFORM_ROLE_PLATFORM_ADMIN: "platform_admin",
}
