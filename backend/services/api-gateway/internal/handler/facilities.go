package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

var facilityTypeByName = map[string]identityv1.FacilityType{
	"raw_material_processing": identityv1.FacilityType_FACILITY_TYPE_RAW_MATERIAL_PROCESSING,
	"component_fabrication":   identityv1.FacilityType_FACILITY_TYPE_COMPONENT_FABRICATION,
	"assembly":                identityv1.FacilityType_FACILITY_TYPE_ASSEMBLY,
	"distribution":            identityv1.FacilityType_FACILITY_TYPE_DISTRIBUTION,
}

var facilityTypeName = invertMap(facilityTypeByName)

var gridRegionByName = map[string]identityv1.GridRegion{
	"US-CAISO": identityv1.GridRegion_GRID_REGION_US_CAISO,
	"US-ERCOT": identityv1.GridRegion_GRID_REGION_US_ERCOT,
	"US-PJM":   identityv1.GridRegion_GRID_REGION_US_PJM,
	"US-MISO":  identityv1.GridRegion_GRID_REGION_US_MISO,
	"EU-DE":    identityv1.GridRegion_GRID_REGION_EU_DE,
	"EU-FR":    identityv1.GridRegion_GRID_REGION_EU_FR,
	"EU-PL":    identityv1.GridRegion_GRID_REGION_EU_PL,
	"UK":       identityv1.GridRegion_GRID_REGION_UK,
	"CN-East":  identityv1.GridRegion_GRID_REGION_CN_EAST,
	"CN-South": identityv1.GridRegion_GRID_REGION_CN_SOUTH,
	"IN-North": identityv1.GridRegion_GRID_REGION_IN_NORTH,
	"JP":       identityv1.GridRegion_GRID_REGION_JP,
	"KR":       identityv1.GridRegion_GRID_REGION_KR,
	"TW":       identityv1.GridRegion_GRID_REGION_TW,
	"VN":       identityv1.GridRegion_GRID_REGION_VN,
	"MY":       identityv1.GridRegion_GRID_REGION_MY,
	"SG":       identityv1.GridRegion_GRID_REGION_SG,
	"TH":       identityv1.GridRegion_GRID_REGION_TH,
}

var gridRegionName = invertMap(gridRegionByName)

var facilityVerificationName = map[identityv1.FacilityVerification]string{
	identityv1.FacilityVerification_FACILITY_VERIFICATION_FACILITY_MATCHED:     "facility_matched",
	identityv1.FacilityVerification_FACILITY_VERIFICATION_ORGANIZATION_MATCHED: "organization_matched",
	identityv1.FacilityVerification_FACILITY_VERIFICATION_SELF_DECLARED:        "self_declared",
}

var trustTierName = map[identityv1.TrustTier]string{
	identityv1.TrustTier_TRUST_TIER_NEW:      "new",
	identityv1.TrustTier_TRUST_TIER_VERIFIED: "verified",
	identityv1.TrustTier_TRUST_TIER_TRUSTED:  "trusted",
}

func invertMap[K comparable, V comparable](source map[K]V) map[V]K {
	inverted := make(map[V]K, len(source))
	for key, value := range source {
		inverted[value] = key
	}
	return inverted
}

type createFacilityRequest struct {
	Name              string `json:"name" binding:"required,max=200"`
	Address           string `json:"address" binding:"required,max=400"`
	CountryCode       string `json:"country_code" binding:"required,len=2"`
	GridRegion        string `json:"grid_region" binding:"required"`
	Type              string `json:"type" binding:"required"`
	FacilityReference string `json:"facility_reference" binding:"max=64"`
	DeclaredCapacity  string `json:"declared_capacity" binding:"required,max=32"`
	DeclaredEnergyKwh string `json:"declared_energy_kwh" binding:"required,max=32"`
}

type facilityResponse struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Address               string  `json:"address"`
	CountryCode           string  `json:"country_code"`
	GridRegion            string  `json:"grid_region"`
	Type                  string  `json:"type"`
	FacilityReference     *string `json:"facility_reference"`
	VerificationStatus    string  `json:"verification_status"`
	CeilingDiscountFactor string  `json:"ceiling_discount_factor"`
	TrustTier             string  `json:"trust_tier"`
	DeclaredCapacity      string  `json:"declared_capacity"`
	DeclaredEnergyKwh     string  `json:"declared_energy_kwh"`
	AttestedCapacity      *string `json:"attested_capacity"`
	AttestedEnergyKwh     *string `json:"attested_energy_kwh"`
	CreatedAt             string  `json:"created_at"`
}

func toFacilityResponse(facility *identityv1.Facility) facilityResponse {
	return facilityResponse{
		ID:                    facility.GetId(),
		Name:                  facility.GetName(),
		Address:               facility.GetAddress(),
		CountryCode:           facility.GetCountryCode(),
		GridRegion:            gridRegionName[facility.GetGridRegion()],
		Type:                  facilityTypeName[facility.GetType()],
		FacilityReference:     emptyToNil(facility.GetFacilityReference()),
		VerificationStatus:    facilityVerificationName[facility.GetVerificationStatus()],
		CeilingDiscountFactor: facility.GetCeilingDiscountFactor(),
		TrustTier:             trustTierName[facility.GetTrustTier()],
		DeclaredCapacity:      facility.GetDeclaredCapacity(),
		DeclaredEnergyKwh:     facility.GetDeclaredEnergyKwh(),
		AttestedCapacity:      emptyToNil(facility.GetAttestedCapacity()),
		AttestedEnergyKwh:     emptyToNil(facility.GetAttestedEnergyKwh()),
		CreatedAt:             facility.GetCreatedAt(),
	}
}

func (h *Handlers) ListFacilities(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	listed, err := h.Identity.ListFacilities(c.Request.Context())
	if err != nil {
		h.failFacility(c, err)
		return
	}

	facilities := make([]facilityResponse, 0, len(listed.GetFacilities()))
	for _, facility := range listed.GetFacilities() {
		facilities = append(facilities, toFacilityResponse(facility))
	}

	httpx.Data(c, http.StatusOK, map[string]any{"facilities": facilities})
}

func (h *Handlers) GetFacility(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	found, err := h.Identity.GetFacility(c.Request.Context(), c.Param("facilityId"))
	if err != nil {
		h.failFacility(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, toFacilityResponse(found.GetFacility()))
}

func (h *Handlers) CreateFacility(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body createFacilityRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	facilityType, known := facilityTypeByName[body.Type]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "type", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	gridRegion, known := gridRegionByName[body.GridRegion]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "grid_region", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	created, err := h.Identity.CreateFacility(c.Request.Context(), key,
		&identityv1.CreateFacilityRequest{
			Name:              body.Name,
			Address:           body.Address,
			CountryCode:       body.CountryCode,
			GridRegion:        gridRegion,
			Type:              facilityType,
			FacilityReference: body.FacilityReference,
			DeclaredCapacity:  body.DeclaredCapacity,
			DeclaredEnergyKwh: body.DeclaredEnergyKwh,
		})
	if err != nil {
		h.failFacility(c, err)
		return
	}

	httpx.Data(c, http.StatusCreated, toFacilityResponse(created.GetFacility()))
}

func (h *Handlers) failFacility(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeForbidden)
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeConflict)
	case codes.Aborted:
		httpx.Fail(c, httpx.CodeConflict)
	case codes.FailedPrecondition:
		httpx.Fail(c, httpx.CodeConflict)
	default:
		h.Logger.Error("facility upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
