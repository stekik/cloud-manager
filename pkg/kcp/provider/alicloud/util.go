package alicloud

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GeneratePassword returns a 32-char password satisfying AliCloud r-kvstore
// requirements: at least one uppercase, one lowercase, one digit.
func GeneratePassword() string {
	const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const digits = "0123456789"
	const all = upper + lower + digits

	randChar := func(charset string) byte {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failure: %v", err))
		}
		return charset[n.Int64()]
	}

	b := make([]byte, 32)
	b[0] = randChar(upper)
	b[1] = randChar(lower)
	b[2] = randChar(digits)
	for i := 3; i < 32; i++ {
		b[i] = randChar(all)
	}
	for i := len(b) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			panic(fmt.Sprintf("crypto/rand failure: %v", err))
		}
		j := int(jBig.Int64())
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}
