package models

import (
	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn


type StudentRegistrationData struct {
	Teacher string `json:"teacher"`
	Students []string `json:"students"`
}
func RegisterStudents(studentRegistrationData StudentRegistrationData) (StudentRegistrationData, error) {

	rows, err := DB.Query("INSERT INTO teacher_student_relationship VALUES ()")
	

	return studentRegistrationData, nil
}