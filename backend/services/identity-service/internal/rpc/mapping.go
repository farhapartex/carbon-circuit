package rpc

import (
	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

var organizationTypes = map[domain.OrganizationType]identityv1.OrganizationType{
	domain.OrganizationManufacturer: identityv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER,
	domain.OrganizationAssembler:    identityv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER,
	domain.OrganizationLogistics:    identityv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS,
	domain.OrganizationCreditBuyer:  identityv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER,
}

var organizationStates = map[domain.OrganizationState]identityv1.OrganizationState{
	domain.OrganizationActive:     identityv1.OrganizationState_ORGANIZATION_STATE_ACTIVE,
	domain.OrganizationRestricted: identityv1.OrganizationState_ORGANIZATION_STATE_RESTRICTED,
	domain.OrganizationReadOnly:   identityv1.OrganizationState_ORGANIZATION_STATE_READ_ONLY,
	domain.OrganizationSuspended:  identityv1.OrganizationState_ORGANIZATION_STATE_SUSPENDED,
}

var verificationStatuses = map[domain.VerificationStatus]identityv1.VerificationStatus{
	domain.VerificationVerified:   identityv1.VerificationStatus_VERIFICATION_STATUS_VERIFIED,
	domain.VerificationUnverified: identityv1.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
	domain.VerificationRejected:   identityv1.VerificationStatus_VERIFICATION_STATUS_REJECTED,
}

var organizationRoles = map[domain.OrganizationRole]identityv1.OrganizationRole{
	domain.RoleOwner:  identityv1.OrganizationRole_ORGANIZATION_ROLE_OWNER,
	domain.RoleAdmin:  identityv1.OrganizationRole_ORGANIZATION_ROLE_ADMIN,
	domain.RoleMember: identityv1.OrganizationRole_ORGANIZATION_ROLE_MEMBER,
}

var platformRoles = map[domain.PlatformRole]identityv1.PlatformRole{
	domain.PlatformVerifier: identityv1.PlatformRole_PLATFORM_ROLE_VERIFIER,
	domain.PlatformAdmin:    identityv1.PlatformRole_PLATFORM_ROLE_PLATFORM_ADMIN,
}

func platformRoleToProto(role *domain.PlatformRole) identityv1.PlatformRole {
	if role == nil {
		return identityv1.PlatformRole_PLATFORM_ROLE_UNSPECIFIED
	}
	return platformRoles[*role]
}

func userToProto(user domain.User) *identityv1.SessionUser {
	return &identityv1.SessionUser{
		Id:           user.ID.String(),
		Email:        user.Email,
		Name:         user.Name,
		PlatformRole: platformRoleToProto(user.PlatformRole),
		MfaEnrolled:  user.MFAEnrolled(),
	}
}

func organizationToProto(organization *domain.Organization, treasuryDesignated bool) *identityv1.SessionOrganization {
	if organization == nil {
		return nil
	}

	return &identityv1.SessionOrganization{
		Id:                 organization.ID.String(),
		Name:               organization.Name,
		Type:               organizationTypes[organization.Type],
		State:              organizationStates[organization.State],
		VerificationStatus: verificationStatuses[organization.VerificationStatus],
		TreasuryDesignated: treasuryDesignated,
	}
}
