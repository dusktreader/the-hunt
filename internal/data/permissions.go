package data

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/dusktreader/the-hunt/internal/types"
	"github.com/lib/pq"
)

type PermissionModel struct {
	DB *sql.DB
	CFG ModelConfig
}

func (m PermissionModel) GetForUser(userID int64) (*types.PermissionSet, error) {
	query := `
		select permissions.code
		from permissions
		join user_permissions on permissions.id = user_permissions.permission_id
		join users on users.id = user_permissions.user_id
		where users.id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := types.NewPermissionSet()
	for rows.Next() {
		var pc types.PermCode
		err := rows.Scan(&pc)
		if err != nil {
			return nil, err
		}
		permissions.Insert(pc)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, perms ...types.PermCode) error {
	slog.Debug("Inserting permissions for user", "userID", userID, "perms", perms)
	query := `
		insert into user_permissions (
			user_id, permission_id
		)
		select $1, permissions.id
		from permissions
		where permissions.code = any($2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), m.CFG.QueryTimeout)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(perms))
	return err
}
