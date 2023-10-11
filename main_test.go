package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"encoding/json"
	"bytes"
	"onecv-go-backend/models"
	"testing"
	"log"
	"fmt"
	"strings"
	"context"
	"errors"
	"regexp"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/gin-gonic/gin"
)

var testRouter *gin.Engine

func init() {
	testRouter = router()
}


type registerStudentsTestCase struct {
	testCaseDesc string
	body  models.StudentRegistrationData[string]
	emailsExist models.StudentRegistrationData[bool]
	studentsRegistered []bool
	wantCode int
	wantResponseBody any
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

	if haveInvalidEmails := len(invalidEmails) > 0; haveInvalidEmails { 
		noRequestErrors = false 

	} else {
		addCheckTeacherExistsQuery(mock, teacher, testCase.emailsExist.Teacher)
		addCheckStudentExistsQueries(mock, students, testCase.emailsExist.Students)

		allEmailsExistence := append(testCase.emailsExist.Students, testCase.emailsExist.Teacher)
		for _, emailExist := range allEmailsExistence {
			if !emailExist { 
				noRequestErrors = false 
			}
		}
	}

	if noRequestErrors { // If no errors up to this point, check if teacher student relationships exist
		addCheckTeacherStudentRelationshipExistsQueries(mock, teacher, students, testCase.studentsRegistered)		

		for _, studentRegistered := range testCase.studentsRegistered {
			if studentRegistered {
				noRequestErrors = false
			}
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

	router.ServeHTTP(recorder, request)

	// make sure that all expectations were met
	checkQueryExpectations(mock, t)
	checkStatusAndResponse[registerStudentsSuccessBody](recorder, t, testCaseStruct{testCase.wantCode, testCase.wantResponseBody})
}

func TestRegisterStudents(t *testing.T) {
	testCases := []registerStudentsTestCase{
        {
			"All valid and existent emails", 
			models.StudentRegistrationData[string]{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: true, Students: []bool{true, true}},
			[]bool{false, false},
			204,
			registerStudentsSuccessBody{},
		},
        {
			"Malformed JSON", 
			models.StudentRegistrationData[string]{Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: true, Students: []bool{true, true}},
			[]bool{false, false},
			customErrors["invalidDataType"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidDataType"].Message, errors.New("invalidDataType")).Error() },
		},		
        {
			"One or more invalid emails", 
			models.StudentRegistrationData[string]{Teacher: "tomgmail.com", Students: []string{"jerry@gmailcom", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: true, Students: []bool{true, true}},
			[]bool{false, false},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{"jerry@gmailcom", "tomgmail.com"}, ", ")).Error() },
		},
        {
			"One or more invalid emails & non-existent email(s)", 
			models.StudentRegistrationData[string]{Teacher: "tomgmail.com", Students: []string{"jerry@gmailcom", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: false, Students: []bool{true, true}},
			[]bool{false, false},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{"jerry@gmailcom", "tomgmail.com"}, ", ")).Error() },
		},	
        {
			"Missing email(s)", 
			models.StudentRegistrationData[string]{Teacher: " ", Students: []string{" ", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: false, Students: []bool{false, true}},
			[]bool{false, false},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{" ", " "}, ", ")).Error() },
		},			
		{
			"Non existent teacher email", 
			models.StudentRegistrationData[string]{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: false, Students: []bool{true, true}},
			[]bool{false, false},
			models.CustomErrors["nonExistentTeacher"].Status,
			errorResponseBody{ fmt.Errorf(models.CustomErrors["nonExistentTeacher"].Message, errors.New("nonExistentTeacher"), "tom@gmail.com").Error() },
		},
		{
			"Non existent student emails", 
			models.StudentRegistrationData[string]{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: true, Students: []bool{false, false}},
			[]bool{false, false},
			models.CustomErrors["nonExistentStudents"].Status,
			errorResponseBody{ fmt.Errorf(models.CustomErrors["nonExistentStudents"].Message, errors.New("nonExistentStudents"), strings.Join([]string{"'jerry@gmail.com'", "'spike@gmail.com'"}, ", ")).Error() },
		},	
		{
			"Non existent student & teacher emails", 
			models.StudentRegistrationData[string]{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: false, Students: []bool{false, true}},
			[]bool{false, false},
			models.CustomErrors["nonExistentTeacher&Students"].Status,
			errorResponseBody{ fmt.Errorf(models.CustomErrors["nonExistentTeacher&Students"].Message, errors.New("nonExistentTeacher&Students"), "tom@gmail.com", strings.Join([]string{"'jerry@gmail.com'"}, ", ")).Error() },
		},	
        {
			"Student(s) already registered with Teacher", 
			models.StudentRegistrationData[string]{Teacher: "tom@gmail.com", Students: []string{"jerry@gmail.com", "spike@gmail.com"}},
			models.StudentRegistrationData[bool]{Teacher: true, Students: []bool{true, true}},
			[]bool{true, false},
			models.CustomErrors["studentsAlreadyRegistered"].Status,
			errorResponseBody{ fmt.Errorf(models.CustomErrors["studentsAlreadyRegistered"].Message, errors.New("studentsAlreadyRegistered"), strings.Join([]string{"'jerry@gmail.com'"}, ", "), "tom@gmail.com").Error() },
		},			
    }

	for _, tc := range testCases {
		t.Run(tc.testCaseDesc, func(t *testing.T) {
			OneRegisterStudentTest(t, testRouter, tc)
		})
	}
}

type commonStudentsTestCase struct {
	testCaseDesc string
	teachers []string
	emailsExist []bool
	studentToTeachers map[string][]string 
	wantCode int
	wantResponseBody any
}

func OneCommonStudentsTest(t *testing.T, router *gin.Engine, testCase commonStudentsTestCase) {
	// First, add expected queries and results to the mock DB
	teachers := testCase.teachers

	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	noRequestErrors := true

	invalidEmails := getInvalidEmails(teachers)

	if haveInvalidEmails := len(invalidEmails) > 0 ; haveInvalidEmails { 
		noRequestErrors = false 

	} else {
		addCheckTeachersExistsQueries(mock, teachers, testCase.emailsExist)

		allEmailsExistence := testCase.emailsExist
		for _, emailExist := range allEmailsExistence {
			if !emailExist { 
				noRequestErrors = false 
			}
		}		
	}

	if noRequestErrors {
		expectedRows := pgxmock.NewRows([]string{"student", "teachers"})
		for student, teachers := range testCase.studentToTeachers {
			expectedRows.AddRow(student, teachers)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT student, array_agg(DISTINCT teacher) AS teachers
		FROM teacher_student_relationship
		WHERE teacher = ANY($1)
		GROUP BY student
	`)).WithArgs(teachers).WillReturnRows(expectedRows)

	}

	models.DB = mock // assign the mock connection's pointer to models.DB so it can be used by the API endpoints

	// Now, we make the API call
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/api/commonstudents", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	q := url.Values{}
	for _, teacher := range teachers {
		q.Add("teacher", teacher)
	}

    request.URL.RawQuery = q.Encode()
	router.ServeHTTP(recorder, request)

	// make sure that all expectations were met
	checkQueryExpectations(mock, t)
	checkStatusAndResponse[commonStudentsSuccessBody](recorder, t, testCaseStruct{testCase.wantCode, testCase.wantResponseBody})
}

func TestCommonStudents(t *testing.T) {
	testCases := []commonStudentsTestCase{
        {
			"All valid and existent emails, 1 teacher", 
			[]string{"tom@gmail.com"},
			[]bool{true, true},
			map[string][]string{
				"jerry@gmail.com": {"tom@gmail.com"},
				"spike@gmail.com": {"tom@gmail.com"},
			},
			200,
			commonStudentsSuccessBody{ []string{"jerry@gmail.com", "spike@gmail.com"} },
		},
        {
			"All valid and existent emails, 2 teachers", 
			[]string{"tom@gmail.com", "quacker@gmail.com"},
			[]bool{true, true},
			map[string][]string{
				"jerry@gmail.com": {"tom@gmail.com", "quacker@gmail.com"},
				"spike@gmail.com": {"tom@gmail.com"},
			},
			200,
			commonStudentsSuccessBody{ []string{"jerry@gmail.com"} },
		},		
        {
			"One or more invalid emails",
			[]string{"tomgmail.com", "quacker@gmail.com"},
			[]bool{true, true},
			map[string][]string{
				"jerry@gmail.com": {"tom@gmail.com", "quacker@gmail.com"},
				"spike@gmail.com": {"tom@gmail.com"},
			},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{"tomgmail.com"}, ", ")).Error() },		
		},
        {
			"One or more invalid emails & non-existent email(s)",
			[]string{"tomgmail.com", "quacker@gmail.com"},
			[]bool{false, true},
			map[string][]string{
				"jerry@gmail.com": {"tom@gmail.com", "quacker@gmail.com"},
				"spike@gmail.com": {"tom@gmail.com"},
			},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{"tomgmail.com"}, ", ")).Error() },
		},		
		{
			"Non existent teacher(s) email", 
			[]string{"tom@gmail.com", "quacker@gmail.com"},
			[]bool{false, true},
			map[string][]string{
				"jerry@gmail.com": {"tom@gmail.com", "quacker@gmail.com"},
				"spike@gmail.com": {"tom@gmail.com"},
			},
			models.CustomErrors["nonExistentTeachers"].Status,
			errorResponseBody{ fmt.Errorf(models.CustomErrors["nonExistentTeachers"].Message, errors.New("nonExistentTeachers"), strings.Join([]string{"'tom@gmail.com'"}, ", ")).Error() },
		},
    }

	for _, tc := range testCases {
		t.Run(tc.testCaseDesc, func(t *testing.T) {
			OneCommonStudentsTest(t, testRouter, tc)
		})
	}
}

type suspendStudentTestCase struct {
	testCaseDesc string
	body  models.StudentSuspensionData[string]
	emailsExist models.StudentSuspensionData[bool]
	wantCode int
	wantResponseBody any
}

func OneSuspendStudentTest(t *testing.T, router *gin.Engine, testCase suspendStudentTestCase) {
	// First, add expected queries and results to the mock DB
	student := testCase.body.Student
	emailExists := testCase.emailsExist.Student

	mock, err := pgxmock.NewConn()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close(context.Background())

	noRequestErrors := true

	if !validateEmail(student) { 
		noRequestErrors = false 
	} else {
		addCheckStudentExistsQuery(mock, student, emailExists)

		if !emailExists { 
			noRequestErrors = false 
		}
	}

	if noRequestErrors {
		mock.ExpectQuery(regexp.QuoteMeta("UPDATE student SET suspended = true WHERE email = $1")).WithArgs(student).WillReturnRows(pgxmock.NewRows([]string{"student", "suspended"}))
	}

	models.DB = mock // assign the mock connection's pointer to models.DB so it can be used by the API endpoints

	// Now, we make the API call
    out, err := json.Marshal(testCase.body)
    if err != nil {
        log.Fatal(err)
    }

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/api/suspend", bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	router.ServeHTTP(recorder, request)

	// make sure that all expectations were met
	checkQueryExpectations(mock, t)
	checkStatusAndResponse[suspendStudentSuccessBody](recorder, t, testCaseStruct{testCase.wantCode, testCase.wantResponseBody})
}

func TestSuspendStudent(t *testing.T) {
	testCases := []suspendStudentTestCase{
        {
			"Valid and existent student email", 
			models.StudentSuspensionData[string]{Student: "jerry@gmail.com"},
			models.StudentSuspensionData[bool]{Student: true},
			204,
			suspendStudentSuccessBody{},
		},
        {
			"Malformed JSON", 
			models.StudentSuspensionData[string]{},
			models.StudentSuspensionData[bool]{Student: true},
			customErrors["invalidDataType"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidDataType"].Message, errors.New("invalidDataType")).Error() },
		},		
        {
			"Invalid student email", 
			models.StudentSuspensionData[string]{Student: "jerry@gmailcom"},
			models.StudentSuspensionData[bool]{Student: true},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{"jerry@gmailcom"}, ", ")).Error() },
		},
        {
			"Invalid and non-existent student email", 
			models.StudentSuspensionData[string]{Student: "jerry@gail.com"},
			models.StudentSuspensionData[bool]{Student: false},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{"jerry@gail.com"}, ", ")).Error() },
		},	
        {
			"Missing email(s)", 
			models.StudentSuspensionData[string]{Student: " "},
			models.StudentSuspensionData[bool]{Student: false},
			customErrors["invalidEmail"].Status,
			errorResponseBody{ fmt.Errorf(customErrors["invalidEmail"].Message, errors.New("invalidEmail"), strings.Join([]string{" "}, ", ")).Error() },
		},			
		{
			"Non existent student email", 
			models.StudentSuspensionData[string]{Student: "jerry@gmail.com"},
			models.StudentSuspensionData[bool]{Student: false},
			models.CustomErrors["nonExistentStudent"].Status,
			errorResponseBody{ fmt.Errorf(models.CustomErrors["nonExistentStudent"].Message, errors.New("nonExistentStudent"), "jerry@gmail.com").Error() },
		},
    }

	for _, tc := range testCases {
		t.Run(tc.testCaseDesc, func(t *testing.T) {
			OneSuspendStudentTest(t, testRouter, tc)
		})
	}
}