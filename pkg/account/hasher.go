package account

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"math/big"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// AzerothCore defaults
var (
	NStr = "894B645E89E1535BBDAD5B8B290650530801B18EBFBF5E8FAB3C82872A3E9BB7"
	N, _ = new(big.Int).SetString(NStr, 16)
	g    = big.NewInt(7)
)

func reverseBytes(b []byte) []byte {
	res := make([]byte, len(b))
	for i, v := range b {
		res[len(b)-1-i] = v
	}
	return res
}

func CalculateSRP6(username, password string) ([]byte, []byte, error) {
	userUpper := strings.ToUpper(username)
	passUpper := strings.ToUpper(password)

	h1Hasher := sha1.New()
	h1Hasher.Write([]byte(userUpper + ":" + passUpper))
	h1 := h1Hasher.Sum(nil)

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, fmt.Errorf("error generating random salt: %w", err)
	}

	h2Hasher := sha1.New()
	h2Hasher.Write(salt)
	h2Hasher.Write(h1)
	h2 := h2Hasher.Sum(nil)

	x := new(big.Int).SetBytes(reverseBytes(h2))
	v := new(big.Int).Exp(g, x, N)

	vBytesBigEndian := make([]byte, 32)
	vBytes := v.Bytes()
	copy(vBytesBigEndian[32-len(vBytes):], vBytes)
	verifier := reverseBytes(vBytesBigEndian)

	return salt, verifier, nil
}
