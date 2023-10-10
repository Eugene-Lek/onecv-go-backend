package main

import (
	"strings"
	"github.com/gin-gonic/gin"
)

// Need a router factory so that the same router can be assessed by test scripts
func router() *gin.Engine {
	router := gin.Default()
	router.POST("/api/register", registerStudents)
	router.GET("/api/commonstudents", getCommonStudents)
	return router
}

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