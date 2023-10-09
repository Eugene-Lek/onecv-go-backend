package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

type StudentRegistrationData struct {
	Teacher  string   `json:"teacher"`
	Students []string `json:"students"`
}

func RegisterStudents(studentRegistrationData StudentRegistrationData) error {
	var err error
	teacher := studentRegistrationData.Teacher

	_, err = DB.Query(context.Background(), "SELECT * FROM teacher WHERE email = $1", teacher)
	if err != nil {
		return errors.New(fmt.Sprintf(errorMessages["unregisteredTeacher"], teacher))
	}

	for _, student := range studentRegistrationData.Students {
		_, err = DB.Query(context.Background(), "SELECT * FROM student WHERE email = $1", student)
		if err != nil {
			return errors.New(fmt.Sprintf(errorMessages["unregisteredStudent"], student))
		}
		_, err = DB.Query(context.Background(), "INSERT INTO teacher_student_relationship(teacher, student) VALUES ($1, $2)", teacher, student)
		if err != nil {
			return err
		}
	}

	return nil
}
