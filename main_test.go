package main

import (
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"bytes"
	"onecv-go-backend/models"
	"testing"
	"log"
	"fmt"
	"strings"
	"context"
	"regexp"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/gin-gonic/gin"
)

type registerStudentsTestCase struct {
	testCaseDesc string
	body  models.StudentRegistrationData
	emailsExist models.StudentRegistrationData
	studentsRegistered []bool
	wantCode int
	wantMessage string
}

func OneRegisterStudentTest(t *testing.T, router *gin.Engine, testCase registerStudentsTestCase) {
	// First, add expected queries and results to the mock DB
	teacher := testCase.body.Teacher
	students := testCase.body.Students

	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	noRequestErrors := true

	allEmails := append(students, teacher)
	invalidEmails := getInvalidEmails(allEmails)
	if len(invalidEmails) > 0 { 
		noRequestErrors = false 
	} else {
		addCheckTeacherExistsQuery(mock, teacher, testCase.emailsExist.Teacher)
		addCheckStudentExistsQueries(mock, students, testCase.emailsExist.Students)
	}

	allEmailsExistence := append(testCase.emailsExist.Students, testCase.emailsExist.Teacher)
	for _, emailExist := range allEmailsExistence {
		if emailExist == "false" { 
			noRequestErrors = false 
		}
	}

	if noRequestErrors { // If no errors up to this point, check if teacher student relationships exist
		addCheckTeacherStudentRelationshipExistsQueries(mock, teacher, students, testCase.studentsRegistered)		
	}

	for _, studentRegistered := range testCase.studentsRegistered {
		if studentRegistered {
			noRequestErrors = false
		}
	}

	if noRequestErrors {
		for _, student := range students {
			mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO teacher_student_relationship(teacher, student) VALUES ($1, $2)")).WithArgs(teacher, student).WillReturnRows(pgxmock.NewRows([]string{"id", "teacher", "student"}))
		}
	}

	models.DB = mock // assign the mock connection's pointer to models.DB so it can be used by the API endpoints

	// Now, we make the API call
    out, err := json.Marshal(testCase.body)
    if err != nil {
        log.Fatal(err)
    }

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/api/register", bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	response := getResponse(request, router, recorder, t)

	// make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}	

	if response.Status != testCase.wantCode {
		t.Errorf("wrong response code:\nwant: %v\n got: %v", testCase.wantCode, response.Status)
	}
	if response.Message != testCase.wantMessage {
		t.Errorf("wrong response message:\nwant: %s\n got: %s", testCase.wantMessage, response.Message)
	}
}

func TestRegisterStudents(t *testing.T) {
	testCases := []registerStudentsTestCase{
        {
			"All valid and existent emails", 
			models.StudentRegistrationData{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "true", Students: []string{"true", "true"}},
			[]bool{false, false},
			204,
			"",
		},
        {
			"One or more invalid emails", 
			models.StudentRegistrationData{Teacher: "tomgmail.com", Students: []string{"jerry@gmailcom", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "true", Students: []string{"true", "true"}},
			[]bool{false, false},
			400,
			strings.Join(strings.Split(fmt.Sprintf(errorMessages["invalidEmail"], strings.Join([]string{"jerry@gmailcom", "tomgmail.com"}, ", ")), ": ")[1:], ": "),
		},
        {
			"One or more invalid emails & non-existent email(s)", 
			models.StudentRegistrationData{Teacher: "tomgmail.com", Students: []string{"jerry@gmailcom", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "false", Students: []string{"true", "true"}},
			[]bool{false, false},
			400,
			strings.Join(strings.Split(fmt.Sprintf(errorMessages["invalidEmail"], strings.Join([]string{"jerry@gmailcom", "tomgmail.com"}, ", ")), ": ")[1:], ": "),
		},	
        {
			"Missing email(s)", 
			models.StudentRegistrationData{Teacher: " ", Students: []string{" ", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "false", Students: []string{"false", "true"}},
			[]bool{false, false},
			400,
			strings.Join(strings.Split(fmt.Sprintf(errorMessages["invalidEmail"], strings.Join([]string{" ", " "}, ", ")), ": ")[1:], ": "),
		},			
		{
			"Non existent teacher email", 
			models.StudentRegistrationData{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "false", Students: []string{"true", "true"}},
			[]bool{false, false},
			400,
			strings.Join(strings.Split(fmt.Sprintf(models.ErrorMessages["nonExistentTeacher"], "tom@gmail.com"), ": ")[1:], ": "),
		},
		{
			"Non existent student emails", 
			models.StudentRegistrationData{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "true", Students: []string{"false", "false"}},
			[]bool{false, false},
			400,
			strings.Join(strings.Split(fmt.Sprintf(models.ErrorMessages["nonExistentStudents"], strings.Join([]string{"'jerry@gmail.com'", "'spike@gmail.com'"}, ", ")), ": ")[1:], ": "),
		},	
		{
			"Non existent student & teacher emails", 
			models.StudentRegistrationData{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "false", Students: []string{"false", "true"}},
			[]bool{false, false},
			400,
			strings.Join(strings.Split(fmt.Sprintf(models.ErrorMessages["nonExistentTeacher&Students"], "tom@gmail.com", strings.Join([]string{"'jerry@gmail.com'"}, ", ")), ": ")[1:], ": "),
		},	
        {
			"Student(s) already registered with Teacher", 
			models.StudentRegistrationData{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData{Teacher: "true", Students: []string{"true", "true"}},
			[]bool{true, false},
			409,
			strings.Join(strings.Split(fmt.Sprintf(models.ErrorMessages["studentsAlreadyRegistered"], strings.Join([]string{"'jerry@gmail.com'"}, ", "), "tom@gmail.com"), ": ")[1:], ": "),
		},			
    }

	router := router()
	for _, tc := range testCases {
		t.Run(tc.testCaseDesc, func(t *testing.T) {
			OneRegisterStudentTest(t, router, tc)
		})
	}
}
