package servicetoken

import "time"

const (
	Issuer   = "carboncircuit-gateway"
	Audience = "carboncircuit-internal"
)

type Caller struct {
	Subject            string `json:"sub"`
	UserID             string `json:"uid"`
	OrganizationID     string `json:"org,omitempty"`
	OrganizationType   string `json:"otype,omitempty"`
	Role               string `json:"role,omitempty"`
	PlatformRole       string `json:"prole,omitempty"`
	PlanTier           string `json:"plan,omitempty"`
	VerificationStatus string `json:"vst,omitempty"`
	OrganizationState  string `json:"ost,omitempty"`
}

func (c Caller) HasOrganization() bool { return c.OrganizationID != "" }

type envelope struct {
	Caller
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func (e envelope) valid(now time.Time, skew time.Duration) bool {
	if e.Issuer != Issuer || e.Audience != Audience {
		return false
	}
	if e.Subject == "" || e.UserID == "" {
		return false
	}
	if e.ExpiresAt == 0 || e.IssuedAt == 0 {
		return false
	}
	if now.After(time.Unix(e.ExpiresAt, 0).Add(skew)) {
		return false
	}
	return !now.Add(skew).Before(time.Unix(e.IssuedAt, 0))
}
