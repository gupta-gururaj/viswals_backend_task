package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
	"github.com/viswals_backend_task/pkg/models"
	"github.com/viswals_backend_task/pkg/postgres"
)

var (
	ErrNoData          = errors.New("requested data does not exist")
	ErrDuplicate       = errors.New("data to create already exists")
	DefaultFieldsCount = 8
)

type Repository struct {
	postgres.Postgres
}

// NewRepository initializes a new repository instance.
func NewRepository(postgres postgres.Postgres) *Repository {
	return &Repository{postgres}
}

// CreateUser inserts a single user record into the database.
func (r *Repository) CreateUser(ctx context.Context, user *models.UserDetails) error {
	// insert data in database.
	_, err := r.DB.ExecContext(ctx, "INSERT INTO user_details (id,first_name,last_name,email_address,created_at,deleted_at,merged_at,parent_user_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8);", user.ID, user.FirstName, user.LastName, user.EmailAddress, user.CreatedAt, user.DeletedAt, user.MergedAt, user.ParentUserId)
	if err != nil {
		return err
	}
	return nil
}

// CreateBulkUsers inserts multiple user records into the database in a single query.
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

// GetUserByID fetches a user by ID from the database.
func (r *Repository) GetUserByID(ctx context.Context, id string) (*models.UserDetails, error) {
	var userDetails models.UserDetails

	row := r.DB.QueryRowContext(ctx, "SELECT id,first_name,last_name,email_address,created_at,deleted_at,merged_at,parent_user_id FROM user_details WHERE id = $1;", id)

	err := row.Scan(&userDetails.ID, &userDetails.FirstName, &userDetails.LastName, &userDetails.EmailAddress, &userDetails.CreatedAt, &userDetails.DeletedAt, &userDetails.MergedAt, &userDetails.ParentUserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoData
		}
		return nil, err
	}
	return &userDetails, nil
}

// GetAllUsers retrieves all user records from the database.
func (r *Repository) GetAllUsers(ctx context.Context) ([]*models.UserDetails, error) {
	var userDetails []*models.UserDetails
	rows, err := r.DB.QueryContext(ctx, "SELECT id,first_name,last_name,email_address,created_at,deleted_at,merged_at,parent_user_id FROM user_details")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var userDetail models.UserDetails
		err := rows.Scan(&userDetail.ID, &userDetail.FirstName, &userDetail.LastName, &userDetail.EmailAddress, &userDetail.CreatedAt, &userDetail.DeletedAt, &userDetail.MergedAt, &userDetail.ParentUserId)
		if err != nil {
			return nil, err
		}

		userDetails = append(userDetails, &userDetail)
	}

	return userDetails, nil
}

// ListUsers fetches a paginated list of users.
func (r *Repository) ListUsers(ctx context.Context, limit, offset int64) ([]*models.UserDetails, error) {
	var userDetails []*models.UserDetails

	rows, err := r.DB.QueryContext(ctx, "SELECT id,first_name,last_name,email_address,created_at,deleted_at,merged_at,parent_user_id FROM user_details ORDER BY id LIMIT $1 OFFSET $2;", limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoData
		}
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var userDetail models.UserDetails
		err := rows.Scan(&userDetail.ID, &userDetail.FirstName, &userDetail.LastName, &userDetail.EmailAddress, &userDetail.CreatedAt, &userDetail.DeletedAt, &userDetail.MergedAt, &userDetail.ParentUserId)
		if err != nil {
			return nil, err
		}

		userDetails = append(userDetails, &userDetail)
	}

	return userDetails, nil
}

// DeleteUser removes a user record by ID.
func (r *Repository)DeleteUser(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, "DELETE FROM user_details WHERE id = $1;", id)
	if err != nil {
		return err
	}

	return nil
}
