@echo off
echo Buidling Main and Backup...

start "" /b cmd /c go build . >nul
start "" /b cmd /c go build -o clipsync_backup.exe . >nul