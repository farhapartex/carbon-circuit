package creditclass

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

var ErrActivityUnknown = errors.New("activity type has no on-chain code")

const domainSeed = "carboncircuit.credit.class.v1"

const (
	RenewableEnergy          = "renewable_energy"
	ReducedEmissionLogistics = "reduced_emission_logistics"
	ResponsibleSourcing      = "responsible_sourcing"
)

var activityCodes = map[string]byte{
	RenewableEnergy:          1,
	ReducedEmissionLogistics: 2,
	ResponsibleSourcing:      3,
}

func Domain() [32]byte {
	return [32]byte(crypto.Keccak256([]byte(domainSeed)))
}

func facilityWord(facilityID uuid.UUID) [32]byte {
	var word [32]byte
	copy(word[16:], facilityID[:])
	return word
}

func numberWord(value uint64) [32]byte {
	var word [32]byte
	big.NewInt(0).SetUint64(value).FillBytes(word[:])
	return word
}

func TokenID(facilityID uuid.UUID, vintageYear int, activityType string) (string, error) {
	code, known := activityCodes[activityType]
	if !known {
		return "", fmt.Errorf("%w: %s", ErrActivityUnknown, activityType)
	}

	if vintageYear < 0 || vintageYear > 65535 {
		return "", fmt.Errorf("vintage year %d does not fit a uint16", vintageYear)
	}

	domain := Domain()
	facility := facilityWord(facilityID)
	vintage := numberWord(uint64(vintageYear))
	activity := numberWord(uint64(code))

	encoded := make([]byte, 0, 128)
	encoded = append(encoded, domain[:]...)
	encoded = append(encoded, facility[:]...)
	encoded = append(encoded, vintage[:]...)
	encoded = append(encoded, activity[:]...)

	return new(big.Int).SetBytes(crypto.Keccak256(encoded)).String(), nil
}
