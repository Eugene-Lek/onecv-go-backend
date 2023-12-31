package models

import (
	"context"
	"fmt"
	"strings"
	"errors"

	"github.com/jackc/pgx/v5"
)

type customError struct {
	Message string
	Status int
}

var CustomErrors = map[string]customError{
	"nonExistentTeacher" : {"%w: The email '%v' does not exist as a teacher", 400},
	"nonExistentTeachers" : {"%w: The email(s) %v do(es) not exist as teacher(s)", 400},
	"nonExistentStudent" : {"%w: The email '%v' does not exist as a student", 400},
	"nonExistentStudents" : {"%w: The email(s) %v do(es) not exist as student(s)", 400},
	"nonExistentTeacher&Students": {"%w: '%v' does not exist as a teacher and %v do(es) not exist as student(s)", 400},
	"studentsAlreadyRegistered": {"%w: Student(s) %v has/have already been registered with the teacher '%v'", 409},
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
		return fmt.Errorf(CustomErrors["nonExistentTeacher&Students"].Message, errors.New("nonExistentTeacher&Students"), teacher, strings.Join(nonExistentStudents, ", "))
	}

	if !teacherExists {
		return fmt.Errorf(CustomErrors["nonExistentTeacher"].Message, errors.New("nonExistentTeacher"), teacher)
	}

	if len(nonExistentStudents) > 0 {
		return fmt.Errorf(CustomErrors["nonExistentStudents"].Message, errors.New("nonExistentStudents"), strings.Join(nonExistentStudents, ", "))
	}

	return nil
}

func checkTeacherStudentRelationshipsExist(teacher string, students []string) ([]string, error) {
	existentStudentTeacherRelationships := []string{}
	for _, student := range students {
		// Check if the teacher, student relationship exists
		var relationshipID string
		err := DB.QueryRow(context.Background(), "SELECT student FROM teacher_student_relationship WHERE teacher = $1 AND student = $2", teacher, student).Scan(&relationshipID)
	
		if err == pgx.ErrNoRows {
			//Do nothing
		} else if err != nil {
			return nil, err
		} else {
			existentStudentTeacherRelationships = append(existentStudentTeacherRelationships, fmt.Sprintf("'%v'", student))
		}
	}	

	return existentStudentTeacherRelationships, nil	
}

func checkStudentSuspended(student string) (bool, error) {
	var suspended bool
	err := DB.QueryRow(context.Background(), "SELECT suspended FROM student WHERE email = $1", student).Scan(&suspended)

	if err != nil {
		return true, err
	}	

	return suspended, nil
}

func removeDuplicateStr(strSlice []string) []string {
    allKeys := make(map[string]bool)
    list := []string{}
    for _, item := range strSlice {
        if _, value := allKeys[item]; !value {
            allKeys[item] = true
            list = append(list, item)
        }
    }
    return list
}