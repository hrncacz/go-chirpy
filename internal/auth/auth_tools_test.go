package auth

import (
	"fmt"
	"testing"
)

func TestHashPassword(t *testing.T) {
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
