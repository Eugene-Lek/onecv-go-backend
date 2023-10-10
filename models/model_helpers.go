package models

import (
	"context"
	"fmt"
)

var errorMessages = map[string]string{
	"unregisteredTeacher" : "The email '%v' has not been registered as a teacher.",
	"unregisteredStudent" : "The email '%v' has not been registered as a student.",
}

func checkTeacherExists(teacher string) error {
	rows, err := DB.Query(context.Background(), "SELECT * FROM teacher WHERE email = $1", teacher)
	if err != nil {
		return err
	}
	rowCount := 0
	for rows.Next() {rowCount++}
	if teacherExists := rowCount > 0; !teacherExists {
		return fmt.Errorf(errorMessages["unregisteredTeacher"], teacher)
	}
	defer rows.Close()	

	return nil
}


func checkStudentExists(student string) error {
	rows, err := DB.Query(context.Background(), "SELECT * FROM student WHERE email = $1", student)
	if err != nil {
		return err
	}
	rowCount := 0
	for rows.Next() {rowCount++}
	if studentExists := rowCount > 0; !studentExists {
		return fmt.Errorf(errorMessages["unregisteredStudent"], student)
	}
	defer rows.Close()	

	return nil
}