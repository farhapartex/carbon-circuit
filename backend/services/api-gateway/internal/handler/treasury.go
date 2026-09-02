package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
)

type treasuryNonceResponse struct {
	Nonce     string `json:"nonce"`
	Domain    string `json:"domain"`
	ChainID   string `json:"chain_id"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
}

type designateTreasuryRequest struct {
	Message   string `json:"message" binding:"required,min=32,max=4096"`
	Signature string `json:"signature" binding:"required,min=64,max=512"`
}

type treasuryResponse struct {
	Address      string `json:"address"`
	DesignatedAt string `json:"designated_at"`
}

func (h *Handlers) IssueTreasuryNonce(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	nonce, err := h.Identity.IssueTreasuryNonce(c.Request.Context())
	if err != nil {
		h.Logger.Error("issue treasury nonce upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
		return
	}

	httpx.Data(c, http.StatusCreated, treasuryNonceResponse{
		Nonce:     nonce.GetNonce(),
		Domain:    nonce.GetDomain(),
		ChainID:   nonce.GetChainId(),
		IssuedAt:  nonce.GetIssuedAt(),
		ExpiresAt: nonce.GetExpiresAt(),
	})
}

func (h *Handlers) DesignateTreasury(c *gin.Context) {
	resolved, present := caller.ContextFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	if !resolved.HasOrganization() {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "organization",
			Code:  "REQUIRED",
		})
		return
	}

	var body designateTreasuryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}

	key, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	designated, err := h.Identity.DesignateTreasury(
		c.Request.Context(), key, body.Message, body.Signature,
	)
	if err != nil {
		h.failDesignate(c, err)
		return
	}

	httpx.Data(c, http.StatusCreated, treasuryResponse{
		Address:      designated.GetAddress(),
		DesignatedAt: designated.GetDesignatedAt(),
	})
}

func (h *Handlers) failDesignate(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeForbidden)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "signature",
			Code:  "PROOF_REJECTED",
		})
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeConflict)
	case codes.Aborted:
		httpx.Fail(c, httpx.CodeRequestInProgress)
	default:
		h.Logger.Error("designate treasury upstream failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
