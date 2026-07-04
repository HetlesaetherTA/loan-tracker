// package pgutils
//
// import (
// 	"strings"
//
// 	"github.com/google/uuid"
// 	"github.com/jackc/pgx/v5/pgtype"
// )
//
// func StringToUUID(uuidStr string) ([16]byte, error) {
// 	tmp, err := uuid.Parse(uuidStr)
//
// 	if err != nil {
// 		return [16]byte{}, err
// 	}
//
// 	return tmp, nil
// }
//
// func NumericToString(numeric pgtype.Numeric) string {
// 	if !numeric.Valid {
// 		return "0"
// 	}
// }
