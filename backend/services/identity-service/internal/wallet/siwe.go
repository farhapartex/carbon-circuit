package wallet

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spruceid/siwe-go"
)

var (
	ErrMessageUnreadable = errors.New("sign-in message could not be parsed")
	ErrDomainMismatch    = errors.New("sign-in message was issued for another domain")
	ErrChainMismatch     = errors.New("sign-in message was signed for another chain")
	ErrSignatureInvalid  = errors.New("signature does not match the message")
)

type Expectation struct {
	Domain  string
	ChainID int
}

type Proof struct {
	Address string
	Nonce   string
}

func Verify(expectation Expectation, rawMessage, signature string) (Proof, error) {
	message, err := siwe.ParseMessage(rawMessage)
	if err != nil {
		return Proof{}, fmt.Errorf("%w: %v", ErrMessageUnreadable, err)
	}

	if !strings.EqualFold(message.GetDomain(), expectation.Domain) {
		return Proof{}, ErrDomainMismatch
	}

	if message.GetChainID() != expectation.ChainID {
		return Proof{}, ErrChainMismatch
	}

	now := time.Now()
	nonce := message.GetNonce()

	if _, err := message.Verify(signature, &expectation.Domain, &nonce, &now); err != nil {
		return Proof{}, fmt.Errorf("%w: %v", ErrSignatureInvalid, err)
	}

	return Proof{
		Address: strings.ToLower(message.GetAddress().Hex()),
		Nonce:   nonce,
	}, nil
}

func (p Proof) lowered() string {
	return strings.ToLower(p.Address)
}
