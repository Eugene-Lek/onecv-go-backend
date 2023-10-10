package main

import (
	"strings"
	"strconv"
)

type errorMessage struct {
	Message string `json:"message" binding:"required"`
}

var errorMessages = map[string]string{
	"invalidEmail" : "400: You have provided one or more invalid emails: %v ",
	"invalidDataType" : "400: The JSON sent does not have the correct structure and/or types",
}

func removeDuplicateStr(strSlice []string) []string {
    allKeys := make(map[string]bool)
    list := []string{}
    for _, item := range strSlice {
        if _, value := allKeys[item]; !value {
            allKeys[item] = true
            list = append(list, item)
        }
    }
    return list
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

func getStatusAndMessage(err error) (int, string) {
	var httpStatus int
	var message string
	originalError := err.Error()

	frontSegment, backSegment, delimiterFound := strings.Cut(originalError, ": ")
	if (!delimiterFound) {
		httpStatus = 500 // delimiter not found == not custom error == internal server error
		message = originalError
	} else {
		intConversionAttempt, err := strconv.Atoi(frontSegment)

		if (err != nil) {
			httpStatus = 500 // frontSegment could not be converted to error code == error code malformed or does not exist
			message = originalError
		} else {
			httpStatus = intConversionAttempt
			message = backSegment
		}
	}
	
	return httpStatus, message
}