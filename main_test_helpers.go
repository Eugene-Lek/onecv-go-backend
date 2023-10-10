package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v3"
)

type Response struct {
	Status int `json:"status" binding:"required"`
	Message string `json:"message"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

func getResponse(request *http.Request, router *gin.Engine, recorder *httptest.ResponseRecorder, t *testing.T) Response {

	router.ServeHTTP(recorder, request)

	var responseBody ResponseBody
	body := recorder.Body.String()
	
	if recorder.Code < 300 {
		responseBody = ResponseBody{""}
	} else {
		if err := json.Unmarshal([]byte(body), &responseBody); err != nil {
			t.Fatalf("parsing json response: %v", err)
		}
	}

	return Response {
		Status: recorder.Code,
		Message: responseBody.Message,
	}
}


func addCheckTeacherExistsQuery(mock pgxmock.PgxConnIface, teacher string, teacherExists string) {
	expectedTeacherRow := pgxmock.NewRows([]string{"email"})
	if (teacherExists == "true") {expectedTeacherRow.AddRow(teacher)}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM teacher WHERE email = $1")).WithArgs(teacher).WillReturnRows(expectedTeacherRow)	
}

func addCheckStudentExistsQueries(mock pgxmock.PgxConnIface, students []string, studentExistences []string) {
	for index, student := range students {
		expectedStudentRow := pgxmock.NewRows([]string{"email"})
		if (studentExistences[index] == "true") {
			expectedStudentRow.AddRow(student)
		}
		
		mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM student WHERE email = $1")).WithArgs(student).WillReturnRows(expectedStudentRow)
	}	
}

func addCheckTeacherStudentRelationshipExistsQueries(mock pgxmock.PgxConnIface, teacher string, students []string, studentsRegistered []bool) {
	for index, student := range students {
		expectedRelationshipRow := pgxmock.NewRows([]string{"student"})
		if (studentsRegistered[index]) {
			expectedRelationshipRow.AddRow(student)
		}
		
		mock.ExpectQuery(regexp.QuoteMeta("SELECT student FROM teacher_student_relationship WHERE teacher = $1 AND student = $2")).WithArgs(teacher, student).WillReturnRows(expectedRelationshipRow)
	}		
}