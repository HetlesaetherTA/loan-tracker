package api

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// Takes in stringified UUID and return UUID
func stringToUUID(uuidStr string) ([16]byte, error) {
	tmp, err := uuid.Parse(uuidStr)

	if err != nil {
		return [16]byte{}, err
	}

	return tmp, nil
}

// Takes a pgtype.Numeric and returns it as a Decimal. If
// parsing numeric fails, returns empty decimal.Decimal
func pgNumericToDecimal(numeric pgtype.Numeric) decimal.Decimal {
	val, err := numeric.Value()

	if err != nil || val == nil {
		return decimal.Decimal{}
	}

	strVal := val.(string)

	decVal, err := decimal.NewFromString(strVal)

	if err != nil {
		return decimal.Decimal{}
	}

	return decVal
}
