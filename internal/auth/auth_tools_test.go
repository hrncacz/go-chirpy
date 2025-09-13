package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashCheckPassword(t *testing.T) {
	cases := []struct {
		inputPasswordToHash  string
		inputPasswordToCheck string
		expected             bool
	}{
		{
			inputPasswordToHash:  "hello world",
			inputPasswordToCheck: "hello world",
			expected:             true,
		},
		{
			inputPasswordToHash:  "fsalfdkjLKJSDFlkjsflksjfALKDFJ.?$$$$$fjsjflasfjd",
			inputPasswordToCheck: "fsalfdkjLKJSDFlkjsflksjfALKDFJ.?$$$$$fjsjflasfjd",
			expected:             true,
		},
		{
			inputPasswordToHash:  "fsalfdkjLKJSDFlkjsflksjfALKDFJ.?$$$$$fjsjflasfjd",
			inputPasswordToCheck: "hello world",
			expected:             false,
		},
	}

	for i, c := range cases {
		hashedPassword, err := HashPassword(c.inputPasswordToHash)
		if err != nil {
			t.Errorf("Error while hashing password: %s", err)
			continue
		}
		err = CheckPasswordHash(c.inputPasswordToCheck, hashedPassword)
		if err != nil {
			if c.expected {
				fmt.Printf("Test nr: %d\n", i)
				t.Errorf("Error while checking password validity: %s", err)
				continue
			}
		}
	}
}

func TestCreateValidateJWT(t *testing.T) {
	cases := []struct {
		userID        uuid.UUID
		signingSecret string
		expiresIn     time.Duration
		waitTime      time.Duration
		expected      bool
	}{
		{
			userID:        uuid.New(),
			signingSecret: "hello world",
			expiresIn:     time.Duration(2 * time.Second),
			waitTime:      time.Duration(1 * time.Second),
			expected:      true,
		},
		{
			userID:        uuid.New(),
			signingSecret: "hello world",
			expiresIn:     time.Duration(2 * time.Second),
			waitTime:      time.Duration(3 * time.Second),
			expected:      false,
		},
	}

	for _, c := range cases {
		token, err := MakeJWT(c.userID, c.signingSecret, c.expiresIn)
		if err != nil {
			t.Errorf("Error while creating JWT: %s", err)
			continue
		}
		time.Sleep(c.waitTime)
		userID, err := ValidateJWT(token, c.signingSecret)
		if err != nil && c.expected {
			t.Errorf("Error while validating JWT: %s", err)
			continue
		}
		if userID != c.userID && c.expected {
			t.Errorf("Returned user ID is not same as input user ID:\n\tInput ID: %s\n\tOutput ID: %s", c.userID.String(), userID.String())
			continue
		}
		continue
	}
}
