package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"regexp"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/google/go-cmp/cmp"
)

func checkQueryExpectations(mock pgxmock.PgxConnIface, t *testing.T) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}	
}

func getResponseBody[responseBodyType any](recorder *httptest.ResponseRecorder, t *testing.T) responseBodyType {
	var responseBody responseBodyType

	body := recorder.Body.String()
	if body == "" {
		return responseBody
	}
	
	err := json.Unmarshal([]byte(body), &responseBody)
	if err != nil {
		t.Fatalf("parsing json response: %v", err)
	}

	return responseBody
}

type testCaseStruct struct {
	wantCode int
	wantResponseBody any
}

func checkStatusAndResponse[registerStudentsSuccessBody any](recorder *httptest.ResponseRecorder, t *testing.T, testCase testCaseStruct) {
	if recorder.Code != testCase.wantCode {
		t.Errorf("wrong response code:\nwant: %v\n got: %v", testCase.wantCode, recorder.Code)
	}

	if recorder.Code < 300 {
		responseBody := getResponseBody[registerStudentsSuccessBody](recorder, t)
		if !cmp.Equal(responseBody, testCase.wantResponseBody) {
			t.Errorf("wrong response body:\nwant: %v\n got: %v", testCase.wantResponseBody, responseBody)
		}
		
	} else {
		responseBody := getResponseBody[errorResponseBody](recorder, t)
		if !cmp.Equal(responseBody, testCase.wantResponseBody) {
			t.Errorf("wrong response body:\nwant: %s\n got: %s", testCase.wantResponseBody, responseBody)
		}
	}	
}

func addCheckStudentExistsQuery(mock pgxmock.PgxConnIface, student string, studentExists bool) {
	expectedStudentRow := pgxmock.NewRows([]string{"email"})
	if (studentExists) {expectedStudentRow.AddRow(student)}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM student WHERE email = $1")).WithArgs(student).WillReturnRows(expectedStudentRow)	
}

func addCheckStudentExistsQueries(mock pgxmock.PgxConnIface, students []string, studentExistences []bool) {
	for index, student := range students {
		expectedStudentRow := pgxmock.NewRows([]string{"email"})
		if (studentExistences[index]) {
			expectedStudentRow.AddRow(student)
		}
		
		mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM student WHERE email = $1")).WithArgs(student).WillReturnRows(expectedStudentRow)
	}	
}

func addCheckTeacherExistsQuery(mock pgxmock.PgxConnIface, teacher string, teacherExists bool) {
	expectedTeacherRow := pgxmock.NewRows([]string{"email"})
	if (teacherExists) {expectedTeacherRow.AddRow(teacher)}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM teacher WHERE email = $1")).WithArgs(teacher).WillReturnRows(expectedTeacherRow)	
}

func addCheckTeachersExistsQueries(mock pgxmock.PgxConnIface, teachers []string, teacherExistences []bool) {
	for index, teacher := range teachers {
		expectedTeacherRow := pgxmock.NewRows([]string{"email"})
		if (teacherExistences[index]) {
			expectedTeacherRow.AddRow(teacher)
		}
		
		mock.ExpectQuery(regexp.QuoteMeta("SELECT email FROM teacher WHERE email = $1")).WithArgs(teacher).WillReturnRows(expectedTeacherRow)
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

func addCheckStudentSuspendedQuery(mock pgxmock.PgxConnIface, student string, studentSuspended bool) {
	expectedStudentRow := pgxmock.NewRows([]string{"studentSuspended"})
	expectedStudentRow.AddRow(studentSuspended)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT suspended FROM student WHERE email = $1")).WithArgs(student).WillReturnRows(expectedStudentRow)	
}