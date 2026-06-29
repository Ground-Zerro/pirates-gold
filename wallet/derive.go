package wallet

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ripemd160"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

type hdKey struct {
	key       [32]byte
	chainCode [32]byte
}

func MnemonicToAddress(mnemonic string) (string, error) {
	seed := pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"), 2048, 64, sha512.New)

	master := masterKey(seed)

	path := []uint32{
		44 | 0x80000000,
		0 | 0x80000000,
		0 | 0x80000000,
		0,
		0,
	}

	child := master
	for _, index := range path {
		child = deriveChild(child, index)
	}

	return privToP2PKH(child.key[:]), nil
}

func masterKey(seed []byte) hdKey {
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	mac.Write(seed)
	I := mac.Sum(nil)

	var k hdKey
	copy(k.key[:], I[:32])
	copy(k.chainCode[:], I[32:])
	return k
}

func deriveChild(parent hdKey, index uint32) hdKey {
	mac := hmac.New(sha512.New, parent.chainCode[:])

	if index >= 0x80000000 {
		mac.Write([]byte{0x00})
		mac.Write(parent.key[:])
	} else {
		mac.Write(compressedPubKey(parent.key[:]))
	}

	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], index)
	mac.Write(buf[:])

	I := mac.Sum(nil)

	n := btcec.S256().N
	IL := new(big.Int).SetBytes(I[:32])
	parentKey := new(big.Int).SetBytes(parent.key[:])
	childKey := new(big.Int).Mod(new(big.Int).Add(IL, parentKey), n)

	var k hdKey
	b := childKey.Bytes()
	copy(k.key[32-len(b):], b)
	copy(k.chainCode[:], I[32:])
	return k
}

func compressedPubKey(privKey []byte) []byte {
	_, pub := btcec.PrivKeyFromBytes(privKey)
	return pub.SerializeCompressed()
}

func privToP2PKH(privKey []byte) string {
	pubKey := compressedPubKey(privKey)

	sha := sha256.Sum256(pubKey)
	ripe := ripemd160.New()
	ripe.Write(sha[:])
	pubKeyHash := ripe.Sum(nil)

	versioned := make([]byte, 1+len(pubKeyHash))
	versioned[0] = 0x00
	copy(versioned[1:], pubKeyHash)

	c1 := sha256.Sum256(versioned)
	c2 := sha256.Sum256(c1[:])

	full := append(versioned, c2[:4]...)
	return base58Encode(full)
}

func base58Encode(input []byte) string {
	n := new(big.Int).SetBytes(input)
	zero := new(big.Int)
	mod := new(big.Int)
	base := big.NewInt(58)

	var result []byte
	for n.Cmp(zero) > 0 {
		n.DivMod(n, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	for _, b := range input {
		if b != 0 {
			break
		}
		result = append(result, base58Alphabet[0])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}
