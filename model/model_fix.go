package model

import (
	"v/common"
)

// DBSimple 简化版数据库接口，用于满足编译需求
type DBSimple interface {
	// 事务相关
	Begin() error
	Commit() error
	Rollback() error

	// 用户相关基础操作
	CreateUser(user *User) error
	GetUser(id int64) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int64) error
	ListUsers(page, pageSize int) ([]*User, error)

	// 代理相关基础操作
	CreateProxy(proxy *common.Proxy) error
	GetProxy(id int64) (*common.Proxy, error)
	GetProxiesByUserID(userID int64) ([]*common.Proxy, error)
	UpdateProxy(proxy *common.Proxy) error
	DeleteProxy(id int64) error
	ListProxies(page, pageSize int) ([]*common.Proxy, error)

	// 关闭数据库
	Close() error
}
