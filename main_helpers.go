package main

import (
	"strings"
)

type errorMessage struct {
	Message string `json:"message" binding:"required"`
}

var errorMessages = map[string]string{
	"invalidEmail" : "You have provided one or more invalid emails: %v ",
	"invalidDataType" : "The JSON sent does not have the correct structure and/or types",
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