package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dusktreader/the-hunt/internal/validator"
)

type User struct {
	ID				int64		`json:"id"`
	CreatedAt		time.Time	`json:"created_at"`
	UpdatedAt		time.Time	`json:"updated_at"`
	Name			string		`json:"name"`
	Email			string		`json:"email"`
	Password		Password	`json:"-"`
	Activated		bool		`json:"activated"`
	Version			int64		`json:"version"`
}

type PartialUser struct {
	Name		*string		`json:"name"`
	Email		*string		`json:"email"`
}

type UserModel struct {
	DB *sql.DB
	CFG ModelConfig
}

var UserSearchFields = NewSearchFields("name", "email")
var UserSortFields = NewSortFields("id", "created_at", "updated_at", "name", "email")

func (u *User) Validate(v *validator.Validator) {
	v.Check(u.Name != "", "name", "must be provided")
	v.Check(len(u.Name) <= 128, "name", "must not be more than 128 bytes")

	v.Check(u.Email != "", "email", "must be provided")
	v.Check(validator.IsEmail(u.Email), "email", "must be a valid email address")

	if u.Password.Plaintext != nil {
		v.Check(*u.Password.Plaintext != "", "password", "must be provided")
		v.Check(len(*u.Password.Plaintext) >= 8, "password", "must be at least 8 bytes long")
		v.Check(len(*u.Password.Plaintext) <= 128, "password", "must not be more than 128 bytes")
	}

	if u.Password.Hashed == nil {
		panic("missing hashed password for user")
	}
}

func (uc *PartialUser) Validate(v *validator.Validator) {
	if uc.Name != nil {
		v.Check(*uc.Name != "", "name", "must be provided")
		v.Check(len(*uc.Name) <= 128, "name", "must not be more than 128 bytes")
	}

	if uc.Email != nil {
		v.Check(*uc.Email != "", "email", "must be provided")
		v.Check(validator.IsEmail(*uc.Email), "email", "must be a valid email address")
	}
}

func (m UserModel) GetVersion(id int64) (int64, error) {
	query := `
		select version
		from users
		where id = $1
	`
	var version int64

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return version, MapError(
		m.DB.QueryRowContext(ctx, query, id).Scan(&version),
		ErrorMap{sql.ErrNoRows: ErrRecordNotFound},
	)
}

func (m UserModel) Insert(user *User) error {
	query := `
		insert into users (name, email)
		values ($1, $2)
		returning id, created_at, updated_at, version
	`
	args := []any{
		user.Name,
		user.Email,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return MapError(m.DB.QueryRowContext(
			ctx,
			query,
			args...,
		).Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Version,
		),
		ErrorMap{".*duplicate key.*": ErrDuplicateKey},
	)

}

func (m UserModel) GetOne(id int64) (*User, error) {
	query := `
		select id, created_at, updated_at, name, email, version
		from (
			select id, created_at, updated_at, name, email, version
			from users
			where id = $1
		)
	`
	var u User

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return &u, MapError(
		m.DB.QueryRowContext(ctx, query, id).Scan(
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Name,
			&u.Email,
			&u.Version,
		),
		ErrorMap{sql.ErrNoRows: ErrRecordNotFound},
	)
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		select id, created_at, updated_at, name, email, version
		from users
		where email = email
	`
	var u User

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return &u, MapError(
		m.DB.QueryRowContext(ctx, query, email).Scan(
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Name,
			&u.Email,
			&u.Version,
		),
		ErrorMap{sql.ErrNoRows: ErrRecordNotFound},
	)
}

func (m UserModel) GetMany(f Filters) ([]*User, *ListMetadata, error) {
	args := []any{}
	query_parts := []string{`
		select count(*) over (), id, created_at, updated_at, name, email, version
		from users
	`}

	where_parts := []string{}

	if f.Search != nil {
		for k, v := range *f.Search {
			args = append(args, v)
			where_parts = append(where_parts, fmt.Sprintf("%s ~* $%d", k, len(args)))
		}
	}

	if len(where_parts) > 0 {
		query_parts = append(query_parts, "where", strings.Join(where_parts, " and "))
	}

	sort_parts := []string{}

	if f.Sort != nil {
		for k, v := range f.Sort.FromOldest() {
			sort_parts = append(sort_parts, fmt.Sprintf("%s %s", k, v))
		}
	}

	if len(sort_parts) > 0 {
		query_parts = append(query_parts, "order by", strings.Join(sort_parts, ", "))
	}

	if f.Page != nil && f.PageSize != nil {
		args = append(args, *f.PageSize, (*f.Page - 1) * *f.PageSize)
		query_parts = append(query_parts, fmt.Sprintf("limit $%d offset $%d", len(args) - 1, len(args)))
	}

	query := strings.Join(query_parts, " ")

	slog.Debug("Assembled GetMany query", "query", query, "args", args)

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var recordCount int
	users := make([]*User, 0, 10)
	for rows.Next() {
		var u User
		err := rows.Scan(
			&recordCount,
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Name,
			&u.Email,
			&u.Version,
		)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	metadata := NewListMetadata(f, recordCount)
	return users, &metadata, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		update users
		set name = $1, email = $2, updated_at = $4, version = version + 1
		where id = $5 and version = $6
		returning version
	`
	args := []any{
		user.Name,
		user.Email,
		time.Now(),
		user.ID,
		user.Version,
	}

	return MapError(
		m.DB.QueryRow(query, args...).Scan(&user.Version),
		ErrorMap{
			sql.ErrNoRows: ErrEditConflict,
			".*duplicate key.*": ErrDuplicateKey,
		},
	)
}

func (m UserModel) PartialUpdate(
	id int64,
	version int64,
	partial *PartialUser,
) (*User, error) {
	query := `
		update users
		set updated_at = $1, version = version + 1
	`
	args := []any{
		time.Now(),
	}

	i := 2
	if partial.Name != nil {
		query += fmt.Sprintf(", name = $%d", i)
		args = append(args, *partial.Name)
		i += 1
	}

	if partial.Email != nil {
		query += fmt.Sprintf(", email = $%d", i)
		args = append(args, *partial.Email)
		i += 1
	}

	query += fmt.Sprintf(`
		where id = $%d and version = $%d
		returning created_at, updated_at, name, email, version
	`, i, i+1)
	args = append(args, id, version)
	u := &User{
		ID: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return u, MapError(
		m.DB.QueryRowContext(ctx, query, args...).Scan(
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Name,
			&u.Email,
			&u.Version,
		),
		ErrorMap{sql.ErrNoRows: ErrEditConflict},
	)
}

func (m UserModel) Delete(id int64) error {
	query := `
		delete from users
		where id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
