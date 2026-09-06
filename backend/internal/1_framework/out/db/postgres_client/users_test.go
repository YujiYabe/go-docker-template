package postgres_client

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	groupObject "backend/internal/4_domain/group_object"
)

func TestCreateUserPersistsAndReturnsGeneratedIdentity(
	t *testing.T,
) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sql mock: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm mock db: %v", err)
	}

	name := "Alice"
	email := "alice@example.com"
	newUser, err := groupObject.NewUser(&groupObject.NewUserArgs{Name: &name, Email: &email})
	if err != nil {
		t.Fatalf("failed to create draft user: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "users" ("email","full_name") VALUES ($1,$2) RETURNING "id"`,
	)).
		WithArgs(email, name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectCommit()

	client := &PostgresClient{Conn: gormDB}
	createdUser, err := client.CreateUser(context.Background(), *newUser)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if createdUser.ID().GetValue() != 7 {
		t.Fatalf("expected generated ID 7, got: %d", createdUser.ID().GetValue())
	}
	if createdUser.Name().GetValue() != name || createdUser.Email().GetValue() != email {
		t.Fatalf("unexpected created user: %+v", createdUser)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
