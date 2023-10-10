package models

import (
	"context"
	"fmt"
	"strings"
	"slices"
	"sort"

	"github.com/jackc/pgx/v5"
)

var DB PgxIface

type PgxIface interface {
	Close(context.Context) error
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type StudentRegistrationData struct {
	Teacher  string   `json:"teacher" binding:"required"`
	Students []string `json:"students" binding:"required"`
}

func RegisterStudents(studentRegistrationData StudentRegistrationData) error {
	teacher := studentRegistrationData.Teacher
	students := studentRegistrationData.Students

	err := checkTeacherStudentsExist(teacher, students)
	if err != nil { return err }

	existentStudentTeacherRelationships, err := checkTeacherStudentRelationshipsExist(teacher, students)
	if err != nil { return err }
	if len(existentStudentTeacherRelationships) > 0 {
		return fmt.Errorf(ErrorMessages["studentsAlreadyRegistered"], strings.Join(existentStudentTeacherRelationships, ", "), teacher)
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
		return nil, fmt.Errorf(ErrorMessages["nonExistentTeachers"], strings.Join(nonExistentTeachers, ", "))
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

	return commonStudents, nil
}
