package model

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

func FixSQLiteFile() error {
	// 读取原始文件
	data, err := ioutil.ReadFile("sqlite.go")
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 创建备份
	if err := ioutil.WriteFile("sqlite.go.backup2", data, 0644); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}

	content := string(data)

	// 删除重复声明的方法
	// 1. ListProtocolStatsByUserID(userID uint, stats *[]*ProtocolStats)
	pattern1 := `(?s)// ListProtocolStatsByUserID 获取用户的所有协议统计\nfunc \(db \*SQLiteDB\) ListProtocolStatsByUserID\(userID uint, stats \*\[\]\*ProtocolStats\) error \{.*?return tx.Error\n\}`

	// 2. GetAllUsersInternal(users *[]*User)
	pattern2 := `(?s)// GetAllUsersInternal 内部方法：获取所有用户\nfunc \(db \*SQLiteDB\) GetAllUsersInternal\(users \*\[\]\*User\) error \{.*?return rows.Err\(\)\n\}`

	// 3. GetProtocolStatsByUserIDPtr(userID uint, stats *[]*ProtocolStats)
	pattern3 := `(?s)// GetProtocolStatsByUserIDPtr 使用指针接收返回值的获取用户协议统计的方法\nfunc \(db \*SQLiteDB\) GetProtocolStatsByUserIDPtr\(userID uint, stats \*\[\]\*ProtocolStats\) error \{.*?return nil\n\}`

	// 编译正则表达式
	re1, _ := regexp.Compile(pattern1)
	re2, _ := regexp.Compile(pattern2)
	re3, _ := regexp.Compile(pattern3)

	// 替换内容
	content = re1.ReplaceAllString(content, "")
	content = re2.ReplaceAllString(content, "")
	content = re3.ReplaceAllString(content, "")

	// 写入修改后的文件
	if err := ioutil.WriteFile("sqlite.go.fixed", []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Println("成功删除重复的方法声明，结果保存在 sqlite.go.fixed")
	return nil
}
