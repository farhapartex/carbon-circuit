package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/httpx"
)

var errUnsupportedCategory = errors.New("unsupported product category")

var organizationTypeByName = map[string]identityv1.OrganizationType{
	"manufacturer": identityv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER,
	"assembler":    identityv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER,
	"logistics":    identityv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS,
	"credit_buyer": identityv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER,
}

var productCategoryByName = map[string]identityv1.ProductCategory{
	"electronics": identityv1.ProductCategory_PRODUCT_CATEGORY_ELECTRONICS,
	"agriculture": identityv1.ProductCategory_PRODUCT_CATEGORY_AGRICULTURE,
	"pharma":      identityv1.ProductCategory_PRODUCT_CATEGORY_PHARMA,
	"textiles":    identityv1.ProductCategory_PRODUCT_CATEGORY_TEXTILES,
}

var rejectionName = map[identityv1.RegistryRejection]string{
	identityv1.RegistryRejection_REGISTRY_REJECTION_ENTITY_DISSOLVED: "entity_dissolved",
	identityv1.RegistryRejection_REGISTRY_REJECTION_SANCTIONS_FLAG:   "sanctions_flag",
	identityv1.RegistryRejection_REGISTRY_REJECTION_NAME_MISMATCH:    "name_mismatch",
}

type createOrganizationRequest struct {
	Name                       string   `json:"name" binding:"required,min=2,max=200"`
	Type                       string   `json:"type" binding:"required"`
	CountryOfIncorporation     string   `json:"country_of_incorporation" binding:"required,len=2"`
	BusinessRegistrationNumber string   `json:"business_registration_number" binding:"required,min=4,max=64"`
	ProductCategories          []string `json:"product_categories"`
}

type verificationOutcomeResponse struct {
	Status         string  `json:"status"`
	Rejection      *string `json:"rejection"`
	MatchFound     bool    `json:"registry_match_found"`
	NameSimilarity *string `json:"name_similarity"`
}

type createOrganizationResponse struct {
	Organization sessionOrganizationResponse `json:"organization"`
	Outcome      verificationOutcomeResponse `json:"outcome"`
}

func (h *Handlers) CreateOrganization(c *gin.Context) {
	caller, verified := auth.CallerFrom(c.Request.Context())
	if !verified {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	var body createOrganizationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	organizationType, known := organizationTypeByName[body.Type]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "type",
			Code:  "UNSUPPORTED_VALUE",
		})
		return
	}

	categories, err := productCategoriesFrom(body.ProductCategories)
	if err != nil {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "product_categories",
			Code:  "UNSUPPORTED_VALUE",
		})
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	created, err := h.Identity.CreateOrganization(c.Request.Context(), key, caller, &identityv1.CreateOrganizationRequest{
		Auth0Subject:               caller.Subject,
		Name:                       body.Name,
		Type:                       organizationType,
		CountryOfIncorporation:     body.CountryOfIncorporation,
		BusinessRegistrationNumber: body.BusinessRegistrationNumber,
		ProductCategories:          categories,
	})
	if err != nil {
		h.failCreate(c, err)
		return
	}

	httpx.Data(c, http.StatusCreated, toCreateOrganizationResponse(created))
}

func productCategoriesFrom(names []string) ([]identityv1.ProductCategory, error) {
	categories := make([]identityv1.ProductCategory, 0, len(names))
	for _, name := range names {
		category, known := productCategoryByName[name]
		if !known {
			return nil, errUnsupportedCategory
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (h *Handlers) failCreate(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.Aborted:
		httpx.Fail(c, httpx.CodeRequestInProgress)
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeConflict)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation)
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	default:
		h.Logger.Error("create organization upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}

func toCreateOrganizationResponse(
	created *identityv1.CreateOrganizationResponse,
) createOrganizationResponse {
	organization := created.GetOrganization()
	outcome := created.GetOutcome()

	return createOrganizationResponse{
		Organization: sessionOrganizationResponse{
			ID:                 organization.GetId(),
			Name:               organization.GetName(),
			Type:               identityOrganizationTypeName[organization.GetType()],
			State:              organizationStateName[organization.GetState()],
			VerificationStatus: verificationStatusName[organization.GetVerificationStatus()],
			Role:               organizationRoleName[created.GetRole()],
		},
		Outcome: verificationOutcomeResponse{
			Status:         verificationStatusName[outcome.GetStatus()],
			Rejection:      emptyToNil(rejectionName[outcome.GetRejection()]),
			MatchFound:     outcome.GetRegistryMatchFound(),
			NameSimilarity: emptyToNil(outcome.GetNameSimilarity()),
		},
	}
}
