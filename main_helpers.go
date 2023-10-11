package main

import (
	"errors"
	"onecv-go-backend/models"
	"strings"
)

type errorResponseBody struct {
	Message string `json:"message"`
}

type customError struct {
	Message string
	Status int
}

var customErrors = map[string]customError{
	"invalidEmail" : {"%w: You have provided one or more invalid emails: %s ", 400},
	"invalidDataType" : {"%w: The JSON sent does not have the correct structure and/or types", 400},
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
	message := err.Error()

	errorCode := errors.Unwrap(err)
	if errorCode == nil {
		httpStatus = 500
		return httpStatus, message
	}

	customErrorMain, errorCodeMainExists := customErrors[errorCode.Error()]
	customErrorModels, errorCodeModelsExists := models.CustomErrors[errorCode.Error()]

	if !errorCodeMainExists && !errorCodeModelsExists {
		httpStatus = 500
	} else if errorCodeMainExists {
		httpStatus = customErrorMain.Status	
	} else {
		httpStatus = customErrorModels.Status	
	}
	return httpStatus, message
}