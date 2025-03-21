package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Backup sqlite.go
	sqliteFile := filepath.Join(cwd, "model", "sqlite.go")
	backupFile := filepath.Join(cwd, "model", "sqlite.go.final_backup")

	sqliteContent, err := ioutil.ReadFile(sqliteFile)
	if err != nil {
		fmt.Printf("Failed to read sqlite.go: %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(backupFile, sqliteContent, 0644)
	if err != nil {
		fmt.Printf("Failed to create backup: %v\n", err)
		os.Exit(1)
	}

	// Generate the fixed sqlite.go file
	sqliteStr := string(sqliteContent)

	// Remove all problematic methods
	methodSignatures := []string{
		"func (db *SQLiteDB) ListProtocolStatsByUserID(userID uint, stats *[]*ProtocolStats)",
		"func (db *SQLiteDB) GetAllUsersInternal(users *[]*User)",
		"func (db *SQLiteDB) GetProtocolStatsByUserIDPtr(userID uint, stats *[]*ProtocolStats)",
	}

	for _, signature := range methodSignatures {
		startIndex := strings.Index(sqliteStr, signature)
		if startIndex != -1 {
			endIndex := strings.Index(sqliteStr[startIndex:], "}")
			if endIndex != -1 {
				sqliteStr = sqliteStr[:startIndex] + sqliteStr[startIndex+endIndex+1:]
			}
		}
	}

	// Save the fixed file
	err = ioutil.WriteFile(sqliteFile, []byte(sqliteStr), 0644)
	if err != nil {
		fmt.Printf("Failed to write fixed file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully fixed sqlite.go by removing duplicate method declarations")
}
