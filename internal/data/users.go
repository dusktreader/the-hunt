package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dusktreader/the-hunt/internal/types"
)

type UserModel struct {
	DB *sql.DB
	CFG ModelConfig
}

var UserSearchFields = NewSearchFields("name", "email")
var UserSortFields = NewSortFields("id", "created_at", "updated_at", "name", "email")

func (m UserModel) GetVersion(id int64) (int64, error) {
	query := `
		select version
		from users
		where id = $1
	`
	var version int64

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return version, types.MapError(
		m.DB.QueryRowContext(ctx, query, id).Scan(&version),
		types.ErrorMap{sql.ErrNoRows: types.ErrRecordNotFound},
	)
}

func (m UserModel) Insert(user *types.User) error {
	query := `
		insert into users (name, email, password_hash, activated)
		values ($1, $2, $3, false)
		returning id, created_at, updated_at, version
	`
	args := []any{
		user.Name,
		user.Email,
		user.HashedPassword,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return types.MapError(m.DB.QueryRowContext(
			ctx,
			query,
			args...,
		).Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Version,
		),
		types.ErrorMap{".*duplicate key.*": types.ErrDuplicateKey},
	)

}

func (m UserModel) GetOne(id int64) (*types.User, error) {
	query := `
		select id, created_at, updated_at, activated, name, email, version
		from (
			select id, created_at, updated_at, activated, name, email, version
			from users
			where id = $1
		)
	`
	var u types.User

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return &u, types.MapError(
		m.DB.QueryRowContext(ctx, query, id).Scan(
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Activated,
			&u.Name,
			&u.Email,
			&u.Version,
		),
		types.ErrorMap{sql.ErrNoRows: types.ErrRecordNotFound},
	)
}

func (m UserModel) GetForLogin(l *types.Login) (*types.User, error) {
	query := `
		select id, created_at, updated_at, activated, name, email, version, password_hash
		from users
		where email = $1
	`
	var u types.User

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, l.Email).Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.Activated,
		&u.Name,
		&u.Email,
		&u.Version,
		&u.HashedPassword,
	)
	if err != nil {
		return nil, types.MapError(err, types.ErrorMap{ sql.ErrNoRows: types.ErrRecordNotFound })
	}

	if !u.Activated {
		return nil, types.ErrUserNotActivated
	}

	err = u.HashedPassword.Compare(l.Password)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (m UserModel) GetForToken(t types.Token) (*types.User, error) {
	slog.Debug("Getting user for token", "token", t.Hash, "scope", t.Scope)
	query := `
		select users.id, users.created_at, users.updated_at, users.name, users.email, users.version
		from tokens
		join users on tokens.user_id = users.id
		where tokens.hash = $1
		and tokens.scope = $2
		and tokens.expires_at > $3
	`
	var u types.User

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return &u, types.MapError(
		m.DB.QueryRowContext(ctx, query, t.Hash, t.Scope, time.Now()).Scan(
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Name,
			&u.Email,
			&u.Version,
		),
		types.ErrorMap{sql.ErrNoRows: types.ErrRecordNotFound},
	)
}

func (m UserModel) GetMany(f Filters) ([]*types.User, *ListMetadata, error) {
	args := []any{}
	query_parts := []string{`
		select
			count(*) over (),
			id,
			created_at,
			updated_at,
			name,
			email,
			activated,
			version
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
	users := make([]*types.User, 0, 10)
	for rows.Next() {
		var u types.User
		err := rows.Scan(
			&recordCount,
			&u.ID,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Name,
			&u.Email,
			&u.Activated,
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

func (m UserModel) Update(user *types.User) error {
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

	return types.MapError(
		m.DB.QueryRow(query, args...).Scan(&user.Version),
		types.ErrorMap{
			sql.ErrNoRows: types.ErrEditConflict,
			".*duplicate key.*": types.ErrDuplicateKey,
		},
	)
}

func (m UserModel) Activate(token types.Token) (int64, error) {
	query := `
		update users
		set activated = true
		where
			id = $1
		returning
			id
	`

	var id int64

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return id, types.MapError(
		m.DB.QueryRowContext(
			ctx,
			query,
			token.UserID,
		).Scan(&id),
		types.ErrorMap{ sql.ErrNoRows: types.ErrNoTokenMatch },
	)
}

func (m UserModel) DeleteTokensForUser(userID int64, scope string) error {
	query := `
		with deleted as (
			delete from tokens
			where user_id = $1
			and scope = $2
			returning user_id
		)
		select count(*) from deleted
	`

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	var count int64

	err := m.DB.QueryRowContext(ctx, query, userID, scope).Scan(&count)
	if err != nil {
		return err
	}
	slog.Debug("Deleted tokens", "userID", userID, "count", count)
	return nil
}

func (m UserModel) PartialUpdate(
	id int64,
	version int64,
	partial *types.PartialUser,
) (*types.User, error) {
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
	u := &types.User{
		ID: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	return u, types.MapError(
		m.DB.QueryRowContext(ctx, query, args...).Scan(
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Name,
			&u.Email,
			&u.Version,
		),
		types.ErrorMap{sql.ErrNoRows: types.ErrEditConflict},
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
		return types.ErrRecordNotFound
	}

	return nil
}
