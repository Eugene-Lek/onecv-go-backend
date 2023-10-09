package main

import (
	"context"
	"os"
	"fmt"
	"strings"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"onecv-go-backend/models"
)

func main() {
	var dbConnectionError error
	models.DB, dbConnectionError = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if dbConnectionError != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", dbConnectionError)
		os.Exit(1)
	}
	defer models.DB.Close(context.Background())

	router := gin.Default()
	router.POST("/api/register", registerStudents)
	router.GET("/api/commonstudents", getCommonStudents)

	router.Run("localhost:8080")
}

func registerStudents(c *gin.Context) {
	var studentRegistrationData models.StudentRegistrationData
    if err := c.BindJSON(&studentRegistrationData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, errorMessage{message: "The json sent does not have the correct structure and/or types"})
        return
    }

	//Parameter validation (check for @gmail.com)
	allEmails := append(studentRegistrationData.Students, studentRegistrationData.Teacher)
	invalidEmails := getInvalidEmails(allEmails)

	if (len(invalidEmails) > 0) {
		c.IndentedJSON(http.StatusBadRequest, errorMessage{message: fmt.Sprintf(errorMessages["invalidEmail"], strings.Join(invalidEmails, ", "))})
        return
	}

	//Register the student
	res, err = models.RegisterStudents(studentRegistrationData)

    c.Status(http.StatusNoContent)

}

func getCommonStudents(c *gin.Context) {
	teachers := c.Request.URL.Query()
	fmt.Println(teachers)
}