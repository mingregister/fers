@echo off
REM ---------------------------------------------------------
REM Build script for current project with CGO_ENABLED=1
REM ---------------------------------------------------------

REM 切换到当前批处理所在目录
cd /d %~dp0

REM 打开 CGO 支持
set CGO_ENABLED=1

REM 如果需要可以指定 gcc 路径（可选）
REM set CC=D:\mingw64\bin\gcc.exe
REM set CXX=D:\mingw64\bin\g++.exe

REM 清理 Go 缓存（可选）
REM go clean -cache

REM 编译当前目录
go build -o app.exe cmd\fers\main.go

REM 提示完成
echo.
echo ============================
echo Build completed: app.exe
echo ============================
pause
