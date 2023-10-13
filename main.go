package main

import (
	"context"
	"fmt"
	"net/http"
	"onecv-go-backend/models"
	"os"
	"strings"
	"log"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil && os.Getenv("GIN_MODE") == "debug" {
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
	router.Run(":8080")
}

// Need a router factory so that the same router can be assessed by test scripts
func router() *gin.Engine {
	router := gin.Default()
	router.POST("/api/register", registerStudents)
	router.GET("/api/commonstudents", getCommonStudents)
	router.POST("/api/suspend", suspendStudent)
	router.POST("/api/retrievefornotifications", retrieveForNotifications)
	return router
}

type registerStudentsSuccessBody struct {}

func registerStudents(c *gin.Context) {
	var studentRegistrationData models.StudentRegistrationData[string]
	if err := c.BindJSON(&studentRegistrationData); err != nil {
		err := fmt.Errorf(customErrors["invalidDataType"].Message, errors.New("invalidDataType"))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{message})			
		return
	}

	//Parameter validation (remove duplicates, check for @gmail.com))
	studentRegistrationData.Students = removeDuplicateStr(studentRegistrationData.Students)

	allEmails := append(studentRegistrationData.Students, studentRegistrationData.Teacher)
	invalidEmails := getInvalidEmails(allEmails)

	if haveInvalidEmails := len(invalidEmails) > 0; haveInvalidEmails {
		err := fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join(invalidEmails, ", "))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{message})				
		return
	}

	//Register the student
	err := models.RegisterStudents(studentRegistrationData)

	if err != nil {
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{Message: message})		
		return
	}

	c.Status(http.StatusNoContent)

}

type commonStudentsSuccessBody struct {
	Students []string `json:"students"`
}

func getCommonStudents(c *gin.Context) {
	queryParams := c.Request.URL.Query()
	teachers := queryParams["teacher"]

	//Parameter validation (remove duplicates, check for @gmail.com)
	teachers = removeDuplicateStr(teachers)

	invalidEmails := getInvalidEmails(teachers)
	if haveInvalidEmails := len(invalidEmails) > 0; haveInvalidEmails {
		err := fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join(invalidEmails, ", "))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{Message: message})				
		return
	}

	//Get common students
	commonStudents, err := models.GetCommonStudents(teachers)
	if err != nil {
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{Message: message})
		return
	}

	c.IndentedJSON(http.StatusOK, commonStudentsSuccessBody{commonStudents})

}

type suspendStudentSuccessBody struct {}

func suspendStudent(c *gin.Context) {
	var studentSuspensionData models.StudentSuspensionData[string]
	if err := c.BindJSON(&studentSuspensionData); err != nil {
		err := fmt.Errorf(customErrors["invalidDataType"].Message, errors.New("invalidDataType"))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{message})			
		return
	}

	//Parameter validation (check for @gmail.com)
	invalidEmails := getInvalidEmails([]string{studentSuspensionData.Student})

	if haveInvalidEmails := len(invalidEmails) > 0; haveInvalidEmails {
		err := fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join(invalidEmails, ", "))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{message})				
		return
	}

	//Register the student
	err := models.SuspendStudent(studentSuspensionData)
	if err != nil {
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{Message: message})		
		return
	}

	c.Status(http.StatusNoContent)
}

type retrieveForNotificationsSuccessBody struct {
	Recipients []string `json:"recipients"`
}

func retrieveForNotifications(c *gin.Context) {
	var retrieveForNotificationsData models.RetrieveForNotificationsData

	if err := c.BindJSON(&retrieveForNotificationsData); err != nil {
		err := fmt.Errorf(customErrors["invalidDataType"].Message, errors.New("invalidDataType"))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{message})			
		return
	}

	teacher := retrieveForNotificationsData.Teacher
	notification := retrieveForNotificationsData.Notification
	notificationWords := strings.Split(notification, " ")

	students := []string{}
	for _, word := range notificationWords {
		if strings.HasPrefix(word, "@") {
			students = append(students, word[1:])
		}
	}

	//Parameter validation (remove duplicates, check for @gmail.com))
	students = removeDuplicateStr(students)

	allEmails := append(students, teacher)
	invalidEmails := getInvalidEmails(allEmails)

	if haveInvalidEmails := len(invalidEmails) > 0; haveInvalidEmails {
		err := fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join(invalidEmails, ", "))
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{message})				
		return
	}

	//Register the student
	retrieveForNotificationsProcessedData := models.RetrieveForNotificationsProcessedData[string] {
		Teacher: teacher,
		Students: students,
	}

	recipients, err := models.RetrieveForNotifications(retrieveForNotificationsProcessedData)

	if err != nil {
		httpStatus, message := getStatusAndMessage(err)
		c.IndentedJSON(httpStatus, errorResponseBody{Message: message})		
		return
	}

	c.IndentedJSON(http.StatusOK, retrieveForNotificationsSuccessBody{recipients})

}