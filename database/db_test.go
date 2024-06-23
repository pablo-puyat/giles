package database

import (
	"database/sql"
	"giles/models"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func Test_GetFilesWithoutHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	expectedRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test.txt")
	mock.ExpectQuery(FilesWithoutHashSql).WillReturnRows(expectedRows)

	want, err := GetFilesWithoutHash(db)

	if err != nil {
		t.Fatalf("Error encoutered while getting list of files: %v", err)
	}

	if len(want) == 0 {
		t.Fatalf("No files to hash")
	}
	if want[0].Id != 1 {
		t.Fatalf("Expected %v, got %v", expectedRows, want)
	}
}

func Test_InsertHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	var hashId int64 = 1
	expect := models.FileData{HashId: hashId}
	mock.ExpectExec("^INSERT").
		WithArgs("test-hash").
		WillReturnResult(sqlmock.NewResult(hashId, 1))

	want := InsertHash(db, models.FileData{Hash: "test-hash"})
	if want.HashId != expect.HashId {
		t.Fatalf("Incorrect id returned: %v", err)
	}
}

func Test_InsertFileIdHashId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	mock.ExpectExec("^INSERT").
		WithArgs(123, 123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	InsertFileIdHashId(db, models.FileData{Id: 123, HashId: 123})
}
