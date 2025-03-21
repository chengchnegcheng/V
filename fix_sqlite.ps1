# This script removes duplicate method declarations from model/sqlite.go
# Based on the instructions in fixed_sqlite_notes.txt

# Backup the original file
Copy-Item -Path "model/sqlite.go" -Destination "model/sqlite.go.backup_before_fix" -Force

# Find the duplicate method declarations (approximate lines)
$duplicateListProtocolStatsByUserID = 2553  # Line where ListProtocolStatsByUserID duplicate starts
$duplicateListProtocolStatsByUserIDEnd = 2557  # Line where it ends

$duplicateGetAllUsersInternal = 2750  # Approximate line where GetAllUsersInternal duplicate starts 
$duplicateGetAllUsersInternalEnd = 2790  # Approximate line where it ends

$duplicateGetProtocolStatsByUserIDPtr = 2800  # Approximate line where GetProtocolStatsByUserIDPtr duplicate starts
$duplicateGetProtocolStatsByUserIDPtrEnd = 2810  # Approximate line where it ends

# Get content of the file
$content = Get-Content -Path "model/sqlite.go"

# Find the exact line numbers for each duplicate function
$lineIndex = 0
$foundDuplicates = @{
    "ListProtocolStatsByUserID" = @{ Start = 0; End = 0 }
    "GetAllUsersInternal" = @{ Start = 0; End = 0 }
    "GetProtocolStatsByUserIDPtr" = @{ Start = 0; End = 0 }
}

foreach ($line in $content) {
    $lineIndex++
    
    # Find ListProtocolStatsByUserID duplicate
    if ($lineIndex -ge 2500 -and $line -match "func \(db \*SQLiteDB\) ListProtocolStatsByUserID\(userID uint, stats \*\[\]\*ProtocolStats\) error") {
        $foundDuplicates["ListProtocolStatsByUserID"].Start = $lineIndex
        # Find the end of this function (the closing bracket)
        $endLineIndex = $lineIndex
        while ($endLineIndex -lt $content.Length -and $content[$endLineIndex] -ne "}") {
            $endLineIndex++
        }
        $foundDuplicates["ListProtocolStatsByUserID"].End = $endLineIndex + 1
    }
    
    # Find GetAllUsersInternal duplicate
    if ($lineIndex -ge 2700 -and $line -match "func \(db \*SQLiteDB\) GetAllUsersInternal\(users \*\[\]\*User\) error") {
        $foundDuplicates["GetAllUsersInternal"].Start = $lineIndex
        # Find the end of this function (the closing bracket)
        $endLineIndex = $lineIndex
        $bracketCount = 1  # We already found the opening bracket
        while ($endLineIndex -lt $content.Length -and $bracketCount -gt 0) {
            $endLineIndex++
            if ($content[$endLineIndex] -match "{") { $bracketCount++ }
            if ($content[$endLineIndex] -match "}") { $bracketCount-- }
        }
        $foundDuplicates["GetAllUsersInternal"].End = $endLineIndex + 1
    }
    
    # Find GetProtocolStatsByUserIDPtr duplicate
    if ($lineIndex -ge 2800 -and $line -match "func \(db \*SQLiteDB\) GetProtocolStatsByUserIDPtr\(userID uint, stats \*\[\]\*ProtocolStats\) error") {
        $foundDuplicates["GetProtocolStatsByUserIDPtr"].Start = $lineIndex
        # Find the end of this function (the closing bracket)
        $endLineIndex = $lineIndex
        while ($endLineIndex -lt $content.Length -and $content[$endLineIndex] -ne "}") {
            $endLineIndex++
        }
        $foundDuplicates["GetProtocolStatsByUserIDPtr"].End = $endLineIndex + 1
    }
}

# Output what we found
Write-Host "Found duplicate declarations at lines:"
Write-Host "ListProtocolStatsByUserID: Lines $($foundDuplicates['ListProtocolStatsByUserID'].Start)-$($foundDuplicates['ListProtocolStatsByUserID'].End)"
Write-Host "GetAllUsersInternal: Lines $($foundDuplicates['GetAllUsersInternal'].Start)-$($foundDuplicates['GetAllUsersInternal'].End)"
Write-Host "GetProtocolStatsByUserIDPtr: Lines $($foundDuplicates['GetProtocolStatsByUserIDPtr'].Start)-$($foundDuplicates['GetProtocolStatsByUserIDPtr'].End)"

# Remove the duplicate methods by creating a new file without them
if ($foundDuplicates["ListProtocolStatsByUserID"].Start -gt 0 -and $foundDuplicates["GetAllUsersInternal"].Start -gt 0 -and $foundDuplicates["GetProtocolStatsByUserIDPtr"].Start -gt 0) {
    # Extract parts of the file
    $part1 = $content | Select-Object -First ($foundDuplicates["ListProtocolStatsByUserID"].Start - 1)
    $part2 = $content | Select-Object -Skip $foundDuplicates["ListProtocolStatsByUserID"].End -First ($foundDuplicates["GetAllUsersInternal"].Start - $foundDuplicates["ListProtocolStatsByUserID"].End - 1)
    $part3 = $content | Select-Object -Skip $foundDuplicates["GetAllUsersInternal"].End -First ($foundDuplicates["GetProtocolStatsByUserIDPtr"].Start - $foundDuplicates["GetAllUsersInternal"].End - 1)
    $part4 = $content | Select-Object -Skip $foundDuplicates["GetProtocolStatsByUserIDPtr"].End
    
    # Combine parts
    $part1 + $part2 + $part3 + $part4 | Set-Content -Path "model/sqlite.go.fixed"
    
    # Replace the original
    Move-Item -Path "model/sqlite.go.fixed" -Destination "model/sqlite.go" -Force
    
    Write-Host "Successfully removed duplicate methods from sqlite.go"
} else {
    Write-Host "Not all duplicate methods were found. Check the file manually."
} 