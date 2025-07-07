package utils

//package for standalone functions that may be called from anywhere

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func IsPortValid(port string) error {
	val, err := strconv.Atoi(strings.TrimPrefix(port, ":"))
	if err != nil {
		return errors.New("port given is not a number")
	}
	if val < 1024 || val > 65535 {
		return errors.New("bad port value given, pick another one")
	}
	return nil
}

func IsAlphaNumeric(str string) bool {
	for _, r := range str {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func Salt() (string, error) {
	salt := make([]byte, 5)
	_, err := rand.Read(salt)
	if err != nil {
		return "", errors.New("salting failed, try pepper")
	}
	return base64.RawURLEncoding.EncodeToString(salt), nil
}

func Hash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func Xor(password, key string) string {
	passBytes := []byte(password)
	keyBytes := []byte(key)
	result := make([]byte, len(passBytes))

	for i, b := range passBytes {
		result[i] = b ^ keyBytes[i%len(keyBytes)]
	}

	return string(result)
}

// Extracts the first part of the email as a username,
// removing all non-alphanumeric characters and padding if needed
func GenerateUsername(input string) string {
	username := strings.Split(input, "@")[0]

	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	username = reg.ReplaceAllString(username, "")

	username = strings.ReplaceAll(username, " ", "_")

	if len(username) < 4 {
		username += fmt.Sprintf("%d", mathrand.Intn(900)+100)
	}

	if len(username) > 15 {
		username = username[:15]
	}

	return username
}

func GenerateOAuthPassword() (string, error) {
	const (
		letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits  = "0123456789"
		special = "!@#$%^&*"
	)

	charset := letters + digits + special

	pass := make([]byte, 12)
	_, err := rand.Read(pass)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	for i := range pass {
		pass[i] = charset[int(pass[i])%len(charset)]
	}

	pass[3] = letters[int(pass[3])%len(letters)]
	pass[7] = digits[int(pass[7])%len(digits)]
	pass[11] = special[int(pass[11])%len(special)]

	mathrand.Shuffle(len(pass), func(i, j int) {
		pass[i], pass[j] = pass[j], pass[i]
	})

	return string(pass), nil
}

func GenerateStateCookie() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state, nil
}

func SetStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func GetStateCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("oauthstate")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func HashPass(password string, salt string, xorKey string) string {
	// Password Hash :necoarcstrangle:
	saltedPass := password + salt
	hashedPass := Hash(saltedPass)
	return Xor(hashedPass, xorKey)
}
