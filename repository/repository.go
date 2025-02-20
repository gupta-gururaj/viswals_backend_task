package repository

import (
	"context"

	"github.com/viswals_backend_task/pkg/models"
	"github.com/viswals_backend_task/pkg/postgres"
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
