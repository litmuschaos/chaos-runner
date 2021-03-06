package utils

import (
	"testing"
)

func TestRandomString(t *testing.T) {

	randomString := RandomString()
	if len(randomString) != 6 {
		t.Fatalf("The length of random string is not of 6 digits")
	}
}
