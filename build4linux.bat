@echo off
setlocal

echo Building investool for Linux amd64 without CGO...

REM 设置交叉编译环境变量（禁用CGO）
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

REM 编译不依赖CGO的可执行程序
go build -o investool-linux-amd64 main.go

if %ERRORLEVEL% EQU 0 (
    echo Build successful: investool-linux-amd64
    echo Checking file type...
    
    REM 使用PowerShell检查文件是否为ELF格式
    powershell -Command "try { $bytes = [System.IO.File]::ReadAllBytes('investool-linux-amd64'); if ($bytes.Length -gt 4 -and $bytes[0] -eq 0x7f -and $bytes[1] -eq 0x45 -and $bytes[2] -eq 0x4c -and $bytes[3] -eq 0x46) { Write-Output 'File is confirmed as ELF format' } else { Write-Output 'File is not ELF format' } } catch { Write-Output 'Error reading file' }"
    
    echo.
    echo To test on Linux:
    echo 1. Copy investool-linux-amd64 to your Linux machine
    echo 2. Make it executable: chmod +x investool-linux-amd64
    echo 3. Run it: ./investool-linux-amd64
) else (
    echo Build failed with error code %ERRORLEVEL%
)

endlocal