package models

import (
	"context"
	"fmt"
	"strings"
	"slices"
	"sort"
	"errors"

	"github.com/jackc/pgx/v5"
)

var DB PgxIface

type PgxIface interface {
	Close(context.Context) error
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type StudentRegistrationData[T any] struct {
	Teacher  T   `json:"teacher" binding:"required"`
	Students []T `json:"students" binding:"required"`
}

func RegisterStudents(studentRegistrationData StudentRegistrationData[string]) error {
	teacher := studentRegistrationData.Teacher
	students := studentRegistrationData.Students

	err := checkTeacherStudentsExist(teacher, students)
	if err != nil { return err }

	existentStudentTeacherRelationships, err := checkTeacherStudentRelationshipsExist(teacher, students)
	if err != nil { return err }
	if len(existentStudentTeacherRelationships) > 0 {
		return fmt.Errorf(CustomErrors["studentsAlreadyRegistered"].Message, errors.New("studentsAlreadyRegistered"), strings.Join(existentStudentTeacherRelationships, ", "), teacher)
	}
	
	for _, student := range students {
		rows, err := DB.Query(context.Background(), "INSERT INTO teacher_student_relationship(teacher, student) VALUES ($1, $2)", teacher, student)
		if err != nil { return err }

		rows.Close()
	}

	return nil
}

func GetCommonStudents(teachers []string) ([]string, error) {
	nonExistentTeachers, err := checkTeachersExist(teachers)
	if err != nil { return nil, err }

	if len(nonExistentTeachers) > 0 {
		return nil, fmt.Errorf(CustomErrors["nonExistentTeachers"].Message, errors.New("nonExistentTeachers"), strings.Join(nonExistentTeachers, ", "))
	}

	rows, err := DB.Query(context.Background(), `
		SELECT student, array_agg(DISTINCT teacher) AS teachers
		FROM teacher_student_relationship
		WHERE teacher = ANY($1)
		GROUP BY student
	`, teachers)

	if err != nil { return nil, err }
	defer rows.Close()

	commonStudents := []string{}
	for rows.Next() {
		var student string
		var studentTeachers []string
		err := rows.Scan(&student, &studentTeachers)
		if err != nil {
			return nil, err
		}
		
		sort.Strings(studentTeachers)
		sort.Strings(teachers)
		if slices.Equal(studentTeachers, teachers) {
			commonStudents = append(commonStudents, student)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Strings(commonStudents) // For consistency with unit tests
	return commonStudents, nil
}


type StudentSuspensionData[T any] struct {
	Student T `json:"student" binding:"required"`
}

func SuspendStudent(studentSuspensionData StudentSuspensionData[string]) error {
	student := studentSuspensionData.Student

	studentExists, err := checkStudentExists(student)
	if err != nil { return err }
	if !studentExists {
		return fmt.Errorf(CustomErrors["nonExistentStudent"].Message, errors.New("nonExistentStudent"), student)
	}

	rows, err := DB.Query(context.Background(), "UPDATE student SET suspended = true WHERE email = $1", student)
	if err != nil { return err }

	rows.Close()

	return nil
}

type RetrieveForNotificationsData struct {
	Teacher  string   `json:"teacher" binding:"required"`
	Notification string `json:"notification" binding:"required"`
}

type RetrieveForNotificationsProcessedData[T any] struct {
	Teacher  T   `json:"teacher" binding:"required"`
	Students []T `json:"students" binding:"required"`
}

func RetrieveForNotifications(retrieveForNotificationsProcessedData RetrieveForNotificationsProcessedData[string]) ([]string, error) {
	teacher := retrieveForNotificationsProcessedData.Teacher
	students := retrieveForNotificationsProcessedData.Students

	err := checkTeacherStudentsExist(teacher, students)
	if err != nil { return nil, err }

	var registeredStudents []string
	err = DB.QueryRow(context.Background(), `
		SELECT array_agg(DISTINCT student) AS students
		FROM teacher_student_relationship
		WHERE teacher = $1
		GROUP BY teacher
	`, teacher).Scan(&registeredStudents)

	if err == pgx.ErrNoRows {
		registeredStudents = []string{}
	} else if err != nil {
		return nil, err
	}

	candidateRecipients := append(students, registeredStudents...)
	
	recipients := []string{}
	for _, candidate := range candidateRecipients {
		suspended, err := checkStudentSuspended(candidate)
		if err != nil { return nil, err }

		if !suspended {
			recipients = append(recipients, candidate)
		}
	}
	recipients = removeDuplicateStr(recipients)
	sort.Strings(recipients)

	return recipients, nil
}