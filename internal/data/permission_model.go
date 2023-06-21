package data

import (
	"context"
	"database/sql"
	"fmt"
)

type PermissionModel struct {
	DB *sql.DB
}

func (m *PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `SELECT p.code
    FROM users u INNER JOIN user_permissions up ON u.ID = up.user_id
    INNER JOIN permissions p ON up.permission_id = p.id
    WHERE u.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query all permissions of user with id=%d: %s", userID, err)
	}
	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var p string

		err = rows.Scan(&p)
		if err != nil {
			return permissions, fmt.Errorf("scan a permission of user with id=%d: %s", userID, err)
		}

		permissions = append(permissions, p)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate all permissions of user with id=%d: %s", userID, err)
	}

	return permissions, nil
}
