package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ErrorMessages = map[string]string{
	"nonExistentTeacher" : "The email '%v' does not exist as a teacher",
	"nonExistentTeachers" : "The email(s) %v do(es) not exist as teacher(s)",
	"nonExistentStudents" : "The email(s) %v do(es) not exist as student(s)",
	"nonExistentTeacher&Students": "'%v' does not exist as a teacher and %v do(es) not exist as student(s)",
}

func checkTeacherExists(teacher string) (bool, error) {
	var email string
	err := DB.QueryRow(context.Background(), "SELECT email FROM teacher WHERE email = $1", teacher).Scan(&email)

	if err == pgx.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func checkTeachersExist(teachers []string) ([]string, error) {
	nonExistentTeachers := []string{}

	for _, teacher := range teachers {
		// Check if teacher's email has been registered
		teacherExists, err := checkTeacherExists(teacher)
		if err != nil { return []string{}, err }

		if !teacherExists {
			nonExistentTeachers = append(nonExistentTeachers, fmt.Sprintf("'%v'", teacher))
		}
	}	

	return nonExistentTeachers, nil
}

func checkStudentExists(student string) (bool, error) {
	var email string
	err := DB.QueryRow(context.Background(), "SELECT email FROM student WHERE email = $1", student).Scan(&email)

	if err == pgx.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}	

	return true, nil
}

func checkStudentsExist(students []string) ([]string, error) {
	nonExistentStudents := []string{}

	for _, student := range students {
		// Check if student's email has been registered
		studentExists, err := checkStudentExists(student)
		if err != nil { return []string{}, err }

		if !studentExists {
			nonExistentStudents = append(nonExistentStudents, fmt.Sprintf("'%v'", student))
		}
	}	

	return nonExistentStudents, nil
}

func checkTeacherStudentsExist(teacher string, students []string) error {
	var err error

	teacherExists, err := checkTeacherExists(teacher)
	if err != nil { return err }
	
	nonExistentStudents, err := checkStudentsExist(students)
	if err != nil { return err }

	if !teacherExists && len(nonExistentStudents) > 0 {
		return fmt.Errorf(ErrorMessages["nonExistentTeacher&Students"], teacher, strings.Join(nonExistentStudents, ", "))
	}

	if !teacherExists {
		return fmt.Errorf(ErrorMessages["nonExistentTeacher"], teacher)
	}

	if len(nonExistentStudents) > 0 {
		return fmt.Errorf(ErrorMessages["nonExistentStudents"], strings.Join(nonExistentStudents, ", "))
	}

	return nil
}