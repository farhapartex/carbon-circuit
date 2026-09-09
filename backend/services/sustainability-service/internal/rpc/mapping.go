package rpc

import (
	"encoding/json"
	"time"

	sustainabilityv1 "github.com/carboncircuit/backend/gen/carboncircuit/sustainability/v1"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

var activityNames = map[sustainabilityv1.ActivityType]domain.ActivityType{
	sustainabilityv1.ActivityType_ACTIVITY_TYPE_RENEWABLE_ENERGY:           domain.RenewableEnergy,
	sustainabilityv1.ActivityType_ACTIVITY_TYPE_REDUCED_EMISSION_LOGISTICS: domain.ReducedEmissionLogistics,
	sustainabilityv1.ActivityType_ACTIVITY_TYPE_RESPONSIBLE_SOURCING:       domain.ResponsibleSourcing,
}

var activityCodes = map[domain.ActivityType]sustainabilityv1.ActivityType{
	domain.RenewableEnergy:          sustainabilityv1.ActivityType_ACTIVITY_TYPE_RENEWABLE_ENERGY,
	domain.ReducedEmissionLogistics: sustainabilityv1.ActivityType_ACTIVITY_TYPE_REDUCED_EMISSION_LOGISTICS,
	domain.ResponsibleSourcing:      sustainabilityv1.ActivityType_ACTIVITY_TYPE_RESPONSIBLE_SOURCING,
}

var statusNames = map[sustainabilityv1.ClaimStatus]domain.ClaimStatus{
	sustainabilityv1.ClaimStatus_CLAIM_STATUS_SUBMITTED:                  domain.Submitted,
	sustainabilityv1.ClaimStatus_CLAIM_STATUS_AI_REVIEW:                  domain.AIReview,
	sustainabilityv1.ClaimStatus_CLAIM_STATUS_HUMAN_REVIEW:               domain.HumanReview,
	sustainabilityv1.ClaimStatus_CLAIM_STATUS_APPROVED:                   domain.Approved,
	sustainabilityv1.ClaimStatus_CLAIM_STATUS_REJECTED:                   domain.Rejected,
	sustainabilityv1.ClaimStatus_CLAIM_STATUS_MORE_INFORMATION_REQUESTED: domain.MoreInformationRequested,
}

var statusCodes = map[domain.ClaimStatus]sustainabilityv1.ClaimStatus{
	domain.Submitted:                sustainabilityv1.ClaimStatus_CLAIM_STATUS_SUBMITTED,
	domain.AIReview:                 sustainabilityv1.ClaimStatus_CLAIM_STATUS_AI_REVIEW,
	domain.HumanReview:              sustainabilityv1.ClaimStatus_CLAIM_STATUS_HUMAN_REVIEW,
	domain.Approved:                 sustainabilityv1.ClaimStatus_CLAIM_STATUS_APPROVED,
	domain.Rejected:                 sustainabilityv1.ClaimStatus_CLAIM_STATUS_REJECTED,
	domain.MoreInformationRequested: sustainabilityv1.ClaimStatus_CLAIM_STATUS_MORE_INFORMATION_REQUESTED,
}

var priorityCodes = map[domain.QueuePriority]sustainabilityv1.QueuePriority{
	domain.Normal:   sustainabilityv1.QueuePriority_QUEUE_PRIORITY_NORMAL,
	domain.High:     sustainabilityv1.QueuePriority_QUEUE_PRIORITY_HIGH,
	domain.Critical: sustainabilityv1.QueuePriority_QUEUE_PRIORITY_CRITICAL,
}

func activityFrom(code sustainabilityv1.ActivityType) (domain.ActivityType, bool) {
	activity, known := activityNames[code]
	return activity, known
}

func statusFrom(code sustainabilityv1.ClaimStatus) domain.ClaimStatus {
	return statusNames[code]
}

func claimMessage(view service.ClaimView) *sustainabilityv1.Claim {
	claim := view.Claim

	figures := map[string]string{}
	if len(claim.DeclaredFigures) > 0 {
		json.Unmarshal(claim.DeclaredFigures, &figures)
	}

	evidence := make([]*sustainabilityv1.ClaimEvidence, 0, len(view.Evidence))
	for _, attachment := range view.Evidence {
		pages := 0
		if attachment.PageCount != nil {
			pages = *attachment.PageCount
		}
		evidence = append(evidence, &sustainabilityv1.ClaimEvidence{
			EvidenceId:  attachment.EvidenceID.String(),
			FileName:    attachment.FileName,
			MediaType:   attachment.MediaType,
			ContentHash: attachment.ContentHash,
			PageCount:   int32(pages),
			ByteSize:    attachment.ByteSize,
		})
	}

	issued := ""
	if claim.IssuedAmount != nil {
		issued = *claim.IssuedAmount
	}

	return &sustainabilityv1.Claim{
		Id:                    claim.ID.String(),
		OrganizationId:        claim.OrganizationID.String(),
		FacilityId:            claim.FacilityID.String(),
		FacilityName:          claim.FacilityName,
		ActivityType:          activityCodes[claim.ActivityType],
		VintageYear:           int32(claim.VintageYear),
		PeriodStart:           claim.PeriodStart.Format(time.DateOnly),
		PeriodEnd:             claim.PeriodEnd.Format(time.DateOnly),
		DeclaredFigures:       figures,
		RequestedAmount:       claim.RequestedAmount,
		ComputedCeiling:       claim.ComputedCeiling,
		CapacityBasis:         claim.CapacityBasis,
		CapacitySource:        claim.CapacitySource,
		DiscountFactor:        claim.DiscountFactor,
		ReferenceFactorValue:  claim.ReferenceFactorValue,
		Status:                statusCodes[claim.Status],
		Priority:              priorityCodes[claim.Priority],
		RequiresDualApproval:  claim.RequiresDualApproval,
		ExclusivityAttestedAt: claim.ExclusivityAttestedAt.UTC().Format(time.RFC3339),
		IssuedAmount:          issued,
		CreatedAt:             claim.CreatedAt.UTC().Format(time.RFC3339),
		Evidence:              evidence,
		ReferenceFactorId:     claim.ReferenceFactorID.String(),
		ReferenceLookupKey:    claim.ReferenceLookupKey,
		VintageCeiling:        claim.VintageCeiling,
		ConsumedAtSubmission:  claim.ConsumedAtSubmission,
	}
}
