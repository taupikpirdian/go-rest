package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"try/go-rest/entity"
	"try/go-rest/models"
	"try/go-rest/repository/mysql/mapper"

	"github.com/go-redis/redis/v8"
	"github.com/rocketlaunchr/dbq/v2"
)

type StudentMysqlInteractor struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewStudnetMysql(db *sql.DB, redisClient *redis.Client) *StudentMysqlInteractor {
	return &StudentMysqlInteractor{
		db:          db,
		redisClient: redisClient,
	}
}

func (b *StudentMysqlInteractor) InsertDataStudent(ctx context.Context, dataStudent *entity.Student) error {
	var (
		errMysql error
	)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	insertQuery := "INSERT INTO students (nim, name, gender, dob, pob, jenjang, study_program, faculty, created_at, updated_at)" +
		"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	_, errMysql = b.db.Exec(insertQuery, dataStudent.GetNim(), dataStudent.GetName(), dataStudent.GetGender(),
		dataStudent.GetDob(), dataStudent.GetPob(), dataStudent.GetJenjang(), dataStudent.GetStudyProgram(), dataStudent.GetFaculty(), dataStudent.GetCreatedAt(), dataStudent.GetUpdatedAt())

	if errMysql != nil {
		return errMysql
	}

	return nil
}

func (s *StudentMysqlInteractor) CheckDataStudentByNim(ctx context.Context, nim string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	queryBuku := fmt.Sprintf("SELECT * FROM %s WHERE nim = ?", models.GetStudentTableName())

	opts := &dbq.Options{
		SingleResult:   true,
		ConcreteStruct: models.ModelStudent{},
		DecoderConfig:  dbq.StdTimeConversionConfig(),
	}

	resultStudent, err := dbq.Q(ctx, s.db, queryBuku, opts, nim)

	if err != nil {
		return false, err
	}

	if resultStudent == nil {
		return true, nil
	} else {
		return false, errors.New("DATA STUDENT SUDAH ADA DALAM DATABASE")
	}
}

func (b *StudentMysqlInteractor) ListDataStudent(ctx context.Context) ([]*entity.Student, error) {
	var (
		errMysql error
	)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	sqlQuery := "SELECT * FROM students"
	rows, errMysql := b.db.QueryContext(ctx, sqlQuery)
	if errMysql != nil {
		return nil, errMysql
	}

	dataStudentCollection := make([]*entity.Student, 0)
	for rows.Next() {
		var (
			id            int
			nim           string
			name          string
			gender        string
			dob           string
			pob           string
			jenjang       string
			study_program string
			faculty       string
			created_at    string
			updated_at    string
		)

		err := rows.Scan(&id, &nim, &name, &gender, &dob, &pob, &jenjang, &study_program, &faculty, &created_at, &updated_at)
		if err != nil {
			return nil, err
		}

		dataStudent, errMapper := mapper.DataStudentDbToEntity(entity.DTOStudent{
			Id:           id,
			Nim:          nim,
			Name:         name,
			Gender:       gender,
			Dob:          dob,
			Pob:          pob,
			Jenjang:      jenjang,
			StudyProgram: study_program,
			Faculty:      faculty,
			CreatedAt:    created_at,
			UpdatedAt:    updated_at,
		})

		if errMapper != nil {
			return nil, errMapper
		}

		dataStudentCollection = append(dataStudentCollection, dataStudent)
	}
	defer rows.Close()

	return dataStudentCollection, nil
}
