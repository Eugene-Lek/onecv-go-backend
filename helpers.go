package main

import (
	"strings"
)

type errorMessage struct {
	message string
}

var errorMessages = map[string]string{
	"invalidEmail" : "You have provided one or more invalid emails:\n%v ",
}

func validateEmail (email string) bool {
	return strings.HasSuffix(email, "@gmail.com")
}

func getInvalidEmails (allEmails []string) []string {

	invalidEmails := []string{}
	for _, email := range allEmails {
		if (!validateEmail(email)) {
			invalidEmails = append(invalidEmails, email)
		}
	}	
	return invalidEmails
}