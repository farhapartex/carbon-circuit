package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	CodeValidation              ErrorCode = "VALIDATION_ERROR"
	CodeUnauthenticated         ErrorCode = "UNAUTHENTICATED"
	CodeTokenRevoked            ErrorCode = "TOKEN_REVOKED"
	CodeForbidden               ErrorCode = "FORBIDDEN"
	CodeOrganizationNotVerified ErrorCode = "ORGANIZATION_NOT_VERIFIED"
	CodeOrganizationRestricted  ErrorCode = "ORGANIZATION_RESTRICTED"
	CodeOrganizationReadOnly    ErrorCode = "ORGANIZATION_READ_ONLY"
	CodePlanLimitExceeded       ErrorCode = "PLAN_LIMIT_EXCEEDED"
	CodeMFARequired             ErrorCode = "MFA_REQUIRED"
	CodeResourceNotFound        ErrorCode = "RESOURCE_NOT_FOUND"
	CodeConflict                ErrorCode = "CONFLICT"
	CodeIdempotencyKeyReused    ErrorCode = "IDEMPOTENCY_KEY_REUSED"
	CodeRequestInProgress       ErrorCode = "REQUEST_IN_PROGRESS"
	CodeEvidenceNotReady        ErrorCode = "EVIDENCE_NOT_READY"
	CodePayloadTooLarge         ErrorCode = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMediaType    ErrorCode = "UNSUPPORTED_MEDIA_TYPE"
	CodeIdempotencyKeyRequired  ErrorCode = "IDEMPOTENCY_KEY_REQUIRED"
	CodeRateLimited             ErrorCode = "RATE_LIMITED"
	CodeInternal                ErrorCode = "INTERNAL_ERROR"
	CodeDependencyUnavailable   ErrorCode = "DEPENDENCY_UNAVAILABLE"
	CodeCapacityShed            ErrorCode = "CAPACITY_SHED"
	CodeGatewayTimeout          ErrorCode = "GATEWAY_TIMEOUT"
)

var codeCatalogue = map[ErrorCode]struct {
	status  int
	message string
}{
	CodeValidation:              {http.StatusBadRequest, "One or more fields are invalid."},
	CodeUnauthenticated:         {http.StatusUnauthorized, "Authentication is required."},
	CodeTokenRevoked:            {http.StatusUnauthorized, "This credential has been revoked."},
	CodeForbidden:               {http.StatusForbidden, "You are not permitted to perform this action."},
	CodeOrganizationNotVerified: {http.StatusForbidden, "This action requires a verified organization."},
	CodeOrganizationRestricted:  {http.StatusForbidden, "This organization is currently restricted."},
	CodeOrganizationReadOnly:    {http.StatusForbidden, "This organization is read-only."},
	CodePlanLimitExceeded:       {http.StatusForbidden, "A plan limit has been reached."},
	CodeMFARequired:             {http.StatusForbidden, "Multi-factor re-authentication is required."},
	CodeResourceNotFound:        {http.StatusNotFound, "The requested resource could not be found."},
	CodeConflict:                {http.StatusConflict, "This request conflicts with the current state."},
	CodeIdempotencyKeyReused:    {http.StatusConflict, "This idempotency key was used with a different request."},
	CodeRequestInProgress:       {http.StatusConflict, "An identical request is already being processed."},
	CodeEvidenceNotReady:        {http.StatusConflict, "The referenced evidence has not finished scanning."},
	CodePayloadTooLarge:         {http.StatusRequestEntityTooLarge, "The request body exceeds its limit."},
	CodeUnsupportedMediaType:    {http.StatusUnsupportedMediaType, "That file type is not accepted."},
	CodeIdempotencyKeyRequired:  {http.StatusUnprocessableEntity, "An Idempotency-Key header is required."},
	CodeRateLimited:             {http.StatusTooManyRequests, "Too many requests."},
	CodeInternal:                {http.StatusInternalServerError, "Something went wrong on our side."},
	CodeDependencyUnavailable:   {http.StatusServiceUnavailable, "A required dependency is unavailable. Try again shortly."},
	CodeCapacityShed:            {http.StatusServiceUnavailable, "The service is shedding load. Try again shortly."},
	CodeGatewayTimeout:          {http.StatusGatewayTimeout, "The request took too long."},
}

type FieldError struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

type errorBody struct {
	Code      ErrorCode    `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id"`
	Details   []FieldError `json:"details,omitempty"`
}

type PageMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func Data(c *gin.Context, status int, payload any) {
	c.JSON(status, gin.H{"data": payload})
}

func Paginated(c *gin.Context, payload any, meta PageMeta) {
	c.JSON(http.StatusOK, gin.H{"data": payload, "meta": meta})
}

func Fail(c *gin.Context, code ErrorCode, details ...FieldError) {
	entry, known := codeCatalogue[code]
	if !known {
		entry = codeCatalogue[CodeInternal]
		code = CodeInternal
	}

	c.AbortWithStatusJSON(entry.status, gin.H{
		"error": errorBody{
			Code:      code,
			Message:   entry.message,
			RequestID: CorrelationID(c),
			Details:   details,
		},
	})
}
