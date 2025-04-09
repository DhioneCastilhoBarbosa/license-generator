package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GerarCódigo gera um código de licença único
func GerarCodigo(meses int) string {
	bytes := make([]byte, 10)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	randomPart := hex.EncodeToString(bytes)

	return fmt.Sprintf("%dM-%s", meses, randomPart)
}
