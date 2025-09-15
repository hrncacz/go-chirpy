package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, jwtSignString string, expiresIn time.Duration) (string, error) {
	currentTime := time.Now()
	expirationTime := currentTime.Add(expiresIn)
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	jwtString, err := token.SignedString([]byte(jwtSignString))
	if err != nil {
		return "", errors.New("error while creating Signed jwt string")
	}
	return jwtString, nil
}

func ValidateJWT(tokenString, jwtSignString string) (uuid.UUID, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(jwtSignString), nil
	})
	if err != nil {
		return uuid.Nil, errors.New("invalid token")
	}
	fmt.Println(claims)
	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, err
	}
	return parsedUserID, err
}

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if len(token) == 0 {
		return "", errors.New("authorization header not found")
	}
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.Trim(token, " ")
	return token, nil
}

func MakeRefreshToken() (string, error) {
	randomData := make([]byte, 32)
	rand.Read(randomData)
	randomString := hex.EncodeToString(randomData)
	return randomString, nil
}
func GetAPIKey(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if len(token) == 0 {
		return "", errors.New("authorization header not found")
	}
	token = strings.TrimPrefix(token, "ApiKey ")
	token = strings.Trim(token, " ")
	return token, nil
}
