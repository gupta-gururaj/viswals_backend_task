package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
	"github.com/viswals_backend_task/pkg/models"
	"github.com/viswals_backend_task/pkg/postgres"
)
var(
	ErrNoData          = errors.New("requested data does not exist")
	ErrDuplicate       = errors.New("data to create already exists")
	DefaultFieldsCount = 8
)

type Repository struct {
	postgres.Postgres
}

func NewRepository(postgres postgres.Postgres) *Repository {
	return &Repository{postgres}
}

func (r *Repository) CreateUser(ctx context.Context, user *models.UserDetails) error {
	// insert data in database.
	_, err := r.DB.ExecContext(ctx, "INSERT INTO user_details (id,first_name,last_name,email_address,created_at,deleted_at,merged_at,parent_user_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);", user.ID, user.FirstName, user.LastName, user.EmailAddress, user.CreatedAt, user.DeletedAt, user.MergedAt, user.ParentUserId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) CreateBulkUsers(ctx context.Context, users []*models.UserDetails) error {

	query := "INSERT INTO user_details (id,first_name,last_name,email_address,created_at,deleted_at,merged_at,parent_user_id) VALUES "

	var queryHolders = make([]string, 0, len(users)*DefaultFieldsCount)
	var valueHolder = make([]interface{}, 0, len(users)*DefaultFieldsCount)

	for i, user := range users {
		queryHolders = append(queryHolders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", i*DefaultFieldsCount+1, i*DefaultFieldsCount+2, i*DefaultFieldsCount+3, i*DefaultFieldsCount+4, i*DefaultFieldsCount+5, i*DefaultFieldsCount+6, i*DefaultFieldsCount+7, i*DefaultFieldsCount+8))
		valueHolder = append(valueHolder, user.ID, user.FirstName, user.LastName, user.EmailAddress, user.CreatedAt, user.DeletedAt, user.MergedAt, user.ParentUserId)
	}

	query += strings.Join(queryHolders, ",")

	// insert data in database.
	_, err := r.DB.ExecContext(ctx, query, valueHolder...)
	if err != nil {
		log.Println(err)
		// check for data already exists.
		var e *pq.Error
		if errors.As(err, &e) && e.Code == "23505" {
			return ErrDuplicate
		}
		return err
	}
	return nil
}
