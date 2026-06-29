package bip39

import (
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"strings"
)

//go:embed wordlist.txt
var wordlistRaw string

var wordlist [2048]string

func init() {
	for i, w := range strings.Split(strings.TrimSpace(wordlistRaw), "\n") {
		wordlist[i] = strings.TrimSpace(w)
	}
}

func GenerateMnemonic() (string, error) {
	entropy := make([]byte, 16)
	if _, err := rand.Read(entropy); err != nil {
		return "", err
	}

	hash := sha256.Sum256(entropy)

	bits := make([]byte, 132)
	for i := 0; i < 128; i++ {
		bits[i] = (entropy[i/8] >> uint(7-i%8)) & 1
	}
	for i := 0; i < 4; i++ {
		bits[128+i] = (hash[0] >> uint(7-i)) & 1
	}

	words := make([]string, 12)
	for i := 0; i < 12; i++ {
		idx := 0
		for j := 0; j < 11; j++ {
			idx = (idx << 1) | int(bits[i*11+j])
		}
		words[i] = wordlist[idx]
	}

	return strings.Join(words, " "), nil
}
