# Create backup of sqlite.go
Copy-Item -Path "sqlite.go" -Destination "sqlite.go.backup" -Force

# Read sqlite.go file
$content = Get-Content "sqlite.go" -Raw

# Define patterns for methods to delete
$pattern1 = '(?s)// ListProtocolStatsByUserID.*?\r?\nfunc \(db \*SQLiteDB\) ListProtocolStatsByUserID\(userID uint, stats \*\[\]\*ProtocolStats\) error \{.*?return tx.Error\r?\n\}'
$pattern2 = '(?s)// GetAllUsersInternal.*?\r?\nfunc \(db \*SQLiteDB\) GetAllUsersInternal\(users \*\[\]\*User\) error \{.*?return rows.Err\(\)\r?\n\}'
$pattern3 = '(?s)// GetProtocolStatsByUserIDPtr.*?\r?\nfunc \(db \*SQLiteDB\) GetProtocolStatsByUserIDPtr\(userID uint, stats \*\[\]\*ProtocolStats\) error \{.*?return nil\r?\n\}'

# Delete duplicate method declarations
$content = $content -replace $pattern1, ""
$content = $content -replace $pattern2, ""
$content = $content -replace $pattern3, ""

# Save modified file
Set-Content -Path "sqlite.go" -Value $content

Write-Host "Removed the following duplicate method declarations:"
Write-Host '1. ListProtocolStatsByUserID(userID uint, stats *[]*ProtocolStats)'
Write-Host '2. GetAllUsersInternal(users *[]*User)'
Write-Host '3. GetProtocolStatsByUserIDPtr(userID uint, stats *[]*ProtocolStats)' 