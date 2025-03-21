package model

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// 打开数据库
func OpenDB(dbPath string, logger *slog.Logger) (DB, error) {
	// 确保目录存在
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %v", err)
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(150)
	db.SetMaxIdleConns(50)

	// 初始化数据库
	sqliteDB := NewSQLiteDB(db, logger)

	// 初始化表结构
	if err := sqliteDB.InitTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化数据库表结构失败: %v", err)
	}

	logger.Info("数据库连接成功", "path", dbPath)
	return sqliteDB, nil
}
