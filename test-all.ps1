# xit 功能测试脚本
# 运行: powershell -ExecutionPolicy Bypass -File test-all.ps1
$ErrorActionPreference = "Continue"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  xit Function Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 1. init
Write-Host "`n=== 1. xit init ===" -ForegroundColor Yellow
Remove-Item -Recurse -Force .xit -ErrorAction SilentlyContinue
./xit.exe init

# 2. no init check
Write-Host "`n=== 2. Error on uninit repo ===" -ForegroundColor Yellow
$tmp = New-Item -ItemType Directory -Force tmp_test -ErrorAction SilentlyContinue
Push-Location tmp_test
../xit.exe status
Pop-Location
Remove-Item -Recurse -Force tmp_test

# 3. create test files
Write-Host "`n=== 3. Create test files ===" -ForegroundColor Yellow
"file A" | Set-Content test_a.txt -NoNewline
"file B" | Set-Content test_b.txt -NoNewline
"echo hello" | Set-Content test.sh -NoNewline
Write-Host "Created: test_a.txt, test_b.txt, test.sh"

# 4. hash-object
Write-Host "`n=== 4. hash-object ===" -ForegroundColor Yellow
$h = ./xit.exe hash-object test_a.txt
Write-Host "Hash of test_a.txt: $h"

# 5. cat-file
Write-Host "`n=== 5. cat-file ===" -ForegroundColor Yellow
./xit.exe cat-file $h
Write-Host ""

# 6. add single
Write-Host "`n=== 6. add single file ===" -ForegroundColor Yellow
./xit.exe add test_a.txt

# 7. add wildcard
Write-Host "`n=== 7. add *.txt (wildcard) ===" -ForegroundColor Yellow
./xit.exe add *.txt

# 8. add shell script
Write-Host "`n=== 8. add test.sh ===" -ForegroundColor Yellow
./xit.exe add test.sh

# 9. status (staged)
Write-Host "`n=== 9. status (staged files) ===" -ForegroundColor Yellow
./xit.exe status

# 10. commit
Write-Host "`n=== 10. commit ===" -ForegroundColor Yellow
./xit.exe commit -m "first commit: add test files"

# 11. commit multi-line
Write-Host "`n=== 11. commit with multi-line -m ===" -ForegroundColor Yellow
"modified A" | Set-Content test_a.txt -NoNewline
./xit.exe add test_a.txt
./xit.exe commit -m "second commit" -m "this is the second paragraph"

# 12. log
Write-Host "`n=== 12. log ===" -ForegroundColor Yellow
./xit.exe log

# 13. XIT_AUTHOR env var
Write-Host "`n=== 13. XIT_AUTHOR env var ===" -ForegroundColor Yellow
$env:XIT_AUTHOR = "Zhang San <zhangsan@test.com>"
"env test" | Set-Content test_env.txt -NoNewline
./xit.exe add test_env.txt
./xit.exe commit -m "custom author test"
Remove-Item env:XIT_AUTHOR -ErrorAction SilentlyContinue
./xit.exe log | Select-String "Zhang San"

# 14. status (clean)
Write-Host "`n=== 14. status (clean) ===" -ForegroundColor Yellow
./xit.exe status

# 15. status (modified)
Write-Host "`n=== 15. status (modified file) ===" -ForegroundColor Yellow
"modified content" | Set-Content test_b.txt -NoNewline
./xit.exe status

# 16. diff
Write-Host "`n=== 16. diff ===" -ForegroundColor Yellow
"hello" | Set-Content diff_a.txt -NoNewline
"hello world" | Set-Content diff_b.txt -NoNewline
$h1 = ./xit.exe hash-object diff_a.txt
$h2 = ./xit.exe hash-object diff_b.txt
./xit.exe diff $h1 $h2

# 17. checkout file restore
Write-Host "`n=== 17. checkout restore file ===" -ForegroundColor Yellow
$firstCommit = ./xit.exe log | Select-String "提交:" | Select-Object -Last 1
$firstHash = ($firstCommit -split " ")[1]
Write-Host "First commit: $firstHash"
./xit.exe checkout $firstHash "--" test_b.txt
Get-Content test_b.txt

# 18. checkout branch switch
Write-Host "`n=== 18. checkout switch branch ===" -ForegroundColor Yellow
./xit.exe checkout main

# 19. cleanup
Write-Host "`n=== 19. Cleanup ===" -ForegroundColor Yellow
Remove-Item -Force test_a.txt, test_b.txt, test.sh, test_env.txt, diff_a.txt, diff_b.txt -ErrorAction SilentlyContinue
Write-Host "Cleaned up test files"

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  ALL TESTS COMPLETED" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
