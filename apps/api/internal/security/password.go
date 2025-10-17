package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

var defaultParams = Argon2Params{
	Time:    3,
	Memory:  64 * 1024,
	Threads: 2,
	KeyLen:  32,
	SaltLen: 16,
}

func HashPassword(password string) ([]byte, error) {
	return HashPasswordWithParams(password, defaultParams)
}

func HashPasswordWithParams(password string, params Argon2Params) ([]byte, error) {
	salt := make([]byte, params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	encoded := base64.StdEncoding.EncodeToString(hash)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	result := fmt.Sprintf("$argon2id$v=19$t=%d,m=%d,p=%d$%s$%s",
		params.Time, params.Memory, params.Threads, encodedSalt, encoded)

	return []byte(result), nil
}

func VerifyPassword(password string, encodedHash []byte) (bool, error) {
	var (
		time    uint32
		memory  uint32
		threads uint8
		saltB64 string
		hashB64 string
	)

	_, err := fmt.Sscanf(string(encodedHash), "$argon2id$v=19$t=%d,m=%d,p=%d$%s$%s",
		&time, &memory, &threads, &saltB64, &hashB64)
	if err != nil {
		return false, fmt.Errorf("parse hash: %w", err)
	}

	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}

	hash, err := base64.StdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	params := Argon2Params{
		Time:    time,
		Memory:  memory,
		Threads: threads,
		KeyLen:  uint32(len(hash)),
		SaltLen: uint32(len(salt)),
	}

	computed := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, params.Threads, params.KeyLen)

	if subtle.ConstantTimeCompare(hash, computed) == 1 {
		return true, nil
	}
	return false, nil
}
