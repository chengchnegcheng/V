package model

import (
	"database/sql"
	"time"
)

// GetAllUsersInternal 内部方法：获取所有用户
func (db *SQLiteDB) GetAllUsersInternal(users *[]*User) error {
	query := `SELECT 
		u.id, u.username, u.email, u.password, u.role, u.status, 
		u.traffic_limit, u.traffic_used, u.is_admin, 
		u.expire_at, u.last_login_at, u.created_at, u.updated_at
		FROM users u`

	rows, err := db.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var result []*User
	for rows.Next() {
		user := &User{}
		var expireAtStr, lastLoginAtStr, createdAtStr, updatedAtStr sql.NullString

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.Status,
			&user.TrafficLimit,
			&user.TrafficUsed,
			&user.IsAdmin,
			&expireAtStr,
			&lastLoginAtStr,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return err
		}

		// 正确处理时间字段
		if expireAtStr.Valid {
			t, err := time.Parse(time.RFC3339, expireAtStr.String)
			if err == nil {
				user.ExpireAt = &t
			}
		}

		if lastLoginAtStr.Valid {
			t, err := time.Parse(time.RFC3339, lastLoginAtStr.String)
			if err == nil {
				user.LastLoginAt = &t
			}
		}

		if createdAtStr.Valid {
			t, err := time.Parse(time.RFC3339, createdAtStr.String)
			if err == nil {
				user.CreatedAt = t
			}
		}

		if updatedAtStr.Valid {
			t, err := time.Parse(time.RFC3339, updatedAtStr.String)
			if err == nil {
				user.UpdatedAt = t
			}
		}

		result = append(result, user)
	}

	*users = result
	return nil
}

// GetProtocolStatsByUserIDPtr 使用指针接收返回值的获取用户协议统计的方法
func (db *SQLiteDB) GetProtocolStatsByUserIDPtr(userID uint, stats *[]*ProtocolStats) error {
	// 使用现有的正确方法实现功能
	result, err := db.ListProtocolStatsByUserID(int64(userID))
	if err != nil {
		return err
	}

	*stats = result
	return nil
}
