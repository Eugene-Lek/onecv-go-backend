package main

import (
	"context"
	"fmt"
	"net/http"
	"onecv-go-backend/models"
	"os"
	"strings"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	var dbConnectionError error
	models.DB, dbConnectionError = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if dbConnectionError != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", dbConnectionError)
		os.Exit(1)
	}
	defer models.DB.Close(context.Background())

	router := router()
	router.Run("localhost:8080")
}

func registerStudents(c *gin.Context) {
	var studentRegistrationData models.StudentRegistrationData
	if err := c.BindJSON(&studentRegistrationData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, errorMessage{Message: errorMessages["invalidDataType"]})
		return
	}

	//Parameter validation (check for @gmail.com)
	allEmails := append(studentRegistrationData.Students, studentRegistrationData.Teacher)
	invalidEmails := getInvalidEmails(allEmails)

	if len(invalidEmails) > 0 {
		c.IndentedJSON(http.StatusBadRequest, errorMessage{Message: fmt.Sprintf(errorMessages["invalidEmail"], strings.Join(invalidEmails, ", "))})
		return
	}

	//Register the student
	err := models.RegisterStudents(studentRegistrationData)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, errorMessage{Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)

}

func getCommonStudents(c *gin.Context) {
	teachers := c.Request.URL.Query()
	fmt.Println(teachers)
}
