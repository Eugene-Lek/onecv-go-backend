package models

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var DB PgxIface

type PgxIface interface {
	Close(context.Context) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type StudentRegistrationData struct {
	Teacher  string   `json:"teacher" binding:"required"`
	Students []string `json:"students" binding:"required"`
}

func RegisterStudents(studentRegistrationData StudentRegistrationData) error {
	teacher := studentRegistrationData.Teacher

	// Check if teacher's email has been registered
	err := checkTeacherExists(teacher)
	if err != nil { return err }
	
	for _, student := range studentRegistrationData.Students {
		// Check if student's email has been registered
		err := checkStudentExists(student)
		if err != nil { return err }

		rows, err := DB.Query(context.Background(), "INSERT INTO teacher_student_relationship(teacher, student) VALUES ($1, $2)", teacher, student)
		if err != nil { return err }
		rows.Close()
	}

	return nil
}
