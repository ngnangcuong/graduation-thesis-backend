package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type argon2Data struct {
	algoType string
	version  string
	memory   string
	times    string
	threads  string
	salt     string
	hash     string
}

func HashPassword(password []byte) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", nil
	}

	algo := "argon2id"
	threads := uint8(4)
	time := uint32(10)
	memory := uint32(32 * 1024)

	hash := argon2.IDKey(password, salt, time, memory, threads, 32)

	b64Hash := base64.StdEncoding.EncodeToString(hash)
	b64Salt := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("$%s$v=%d$m=%d, t=%d, p=%d$%s%s", algo, argon2.Version, memory, time, threads, b64Salt, b64Hash), nil
}

func HashPasswordSettings(password []byte, salt []byte, algo string, time, memory uint32, threads uint8, keyLength uint32) string {
	var hash []byte
	switch algo {
	case "argon2id":
		hash = argon2.IDKey(password, salt, time, memory, threads, keyLength)
	case "argon2i":
		hash = argon2.Key(password, salt, time, memory, threads, keyLength)
	}

	b64Hash := base64.StdEncoding.EncodeToString(hash)
	b64Salt := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("$%s$v=%d$m=%d, t=%d, p=%d$%s$%s", algo, argon2.Version, memory, time, threads, b64Salt, b64Hash)
}

func generateSalt() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return b, err
	}

	return b, nil
}

func Compare(encodedHash string, password []byte) (bool, error) {
	data, err := split(encodedHash)
	if err != nil {
		return false, err
	}

	salt, err := base64.StdEncoding.DecodeString(data.salt)
	if err != nil {
		return false, err
	}

	saveHash, err := base64.StdEncoding.DecodeString(data.hash)
	if err != nil {
		return false, nil
	}

	mem, err := strconv.Atoi(data.memory)
	if err != nil {
		return false, err
	}

	times, err := strconv.Atoi(data.times)
	if err != nil {
		return false, err
	}

	threads, err := strconv.Atoi(data.threads)
	if err != nil {
		return false, err
	}

	encoded := HashPasswordSettings(password, salt, "argon2id", uint32(times), uint32(mem), uint8(threads), uint32(len(saveHash)))

	return subtle.ConstantTimeCompare([]byte(encoded), []byte(encodedHash)) == 1, nil
}

func split(encodedHash string) (*argon2Data, error) {
	parts := make([]string, 0)
	splits := strings.SplitAfter(encodedHash, "$")
	splits = splits[1:]

	for _, v := range splits {
		parts = append(parts, strings.TrimSuffix(v, "$"))
	}

	versionStr := strings.Split(parts[1], "=")[1]
	_, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil, err
	}

	parameters := strings.Split(parts[2], ",")

	memStr := strings.Split(parameters[0], "=")[1]

	timesStr := strings.Split(parameters[1], "=")[1]

	threadsStr := strings.Split(parameters[2], "=")[1]

	data := argon2Data{
		algoType: parts[0],
		version:  versionStr,
		memory:   memStr,
		times:    timesStr,
		threads:  threadsStr,
		salt:     parts[3],
		hash:     parts[4],
	}

	return &data, nil
}
