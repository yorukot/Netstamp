package system

import (
	"encoding/json"
	"time"
)

type Setting struct {
	Key                 string
	Value               json.RawMessage
	EncryptedValue      []byte
	EncryptedValueNonce []byte
	Secret              bool
	UpdatedByUserID     *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
