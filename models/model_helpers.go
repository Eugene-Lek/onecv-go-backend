package models

import (
	"context"
	"fmt"
	"strings"
)

var errorMessages = map[string]string{
	"nonExistentTeacher" : "The email '%v' does not exist as a teacher.",
	"nonExistentStudents" : "The email(s) '%v' do(es) not exist as student(s).",
	"nonExistentTeacher&Students": "'%v' does not exist as a teacher and '%v' do(es) not exist as student(s)",
}

func checkTeacherExists(teacher string) (bool, error) {
	rows, err := DB.Query(context.Background(), "SELECT * FROM teacher WHERE email = $1", teacher)
	if err != nil {
		return false, err
	}
	rowCount := 0
	for rows.Next() {rowCount++}
	if teacherExists := rowCount > 0; !teacherExists {
		return false, nil //fmt.Errorf(errorMessages["unregisteredTeacher"], teacher)
	}
	defer rows.Close()	

	return true, nil
}

func checkStudentExists(student string) (bool, error) {
	rows, err := DB.Query(context.Background(), "SELECT * FROM student WHERE email = $1", student)
	if err != nil {
		return false, err
	}
	rowCount := 0
	for rows.Next() {rowCount++}
	if studentExists := rowCount > 0; !studentExists {
		return false, nil //fmt.Errorf(errorMessages["unregisteredStudent"], student)
	}
	defer rows.Close()	

	return true, nil
}

func checkStudentsExist(students []string) ([]string, error) {
	nonExistentStudents := []string{}

	for _, student := range students {
		// Check if student's email has been registered
		studentExists, err := checkStudentExists(student)
		if err != nil { return []string{}, err }

		if !studentExists {
			nonExistentStudents = append(nonExistentStudents, student)
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
		return fmt.Errorf(errorMessages["nonExistentTeacher&Students"], teacher, strings.Join(nonExistentStudents, ", "))
	}

	if !teacherExists {
		return fmt.Errorf(errorMessages["nonExistentTeacher"], teacher)
	}

	if len(nonExistentStudents) > 0 {
		return fmt.Errorf(errorMessages["nonExistentStudents"], strings.Join(nonExistentStudents, ", "))
	}

	return nil
}