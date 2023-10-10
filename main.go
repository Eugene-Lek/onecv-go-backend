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

// Need a router factory so that the same router can be assessed by test scripts
func router() *gin.Engine {
	router := gin.Default()
	router.POST("/api/register", registerStudents)
	router.GET("/api/commonstudents", getCommonStudents)
	return router
}

func registerStudents(c *gin.Context) {
	var studentRegistrationData models.StudentRegistrationData
	if err := c.BindJSON(&studentRegistrationData); err != nil {
		err := fmt.Errorf(errorMessages["invalidDataType"])
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorMessage{Message: message})			
		return
	}

	//Parameter validation (remove duplicates, check for @gmail.com))
	studentRegistrationData.Students = removeDuplicateStr(studentRegistrationData.Students)

	allEmails := append(studentRegistrationData.Students, studentRegistrationData.Teacher)
	invalidEmails := getInvalidEmails(allEmails)

	if len(invalidEmails) > 0 {
		err := fmt.Errorf(errorMessages["invalidEmail"], strings.Join(invalidEmails, ", "))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorMessage{Message: message})				
		return
	}

	//Register the student
	err := models.RegisterStudents(studentRegistrationData)
	fmt.Println(err)
	if err != nil {
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorMessage{Message: message})		
		return
	}

	c.Status(http.StatusNoContent)

}

type commonStudentsResponse struct {
	Students []string `json:"students"`
}

func getCommonStudents(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	teachers := queryParams["teacher"]

	//Parameter validation (remove duplicates, check for @gmail.com)
	teachers = removeDuplicateStr(teachers)

	invalidEmails := getInvalidEmails(teachers)
	if len(invalidEmails) > 0 {
		c.IndentedJSON(http.StatusBadRequest, errorMessage{Message: fmt.Sprintf(errorMessages["invalidEmail"], strings.Join(invalidEmails, ", "))})
		return
	}

	//Get common students
	commonStudents, err := models.GetCommonStudents(teachers)
	if err != nil {
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorMessage{Message: message})
		return
	}

	c.IndentedJSON(http.StatusOK, commonStudentsResponse{commonStudents})

}
