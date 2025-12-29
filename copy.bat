@echo off
echo Copying...
set "Home=C:\PROGRAMMING\ClipSync\clipsync.exe"
set "Dest=\\ABUNDANCEPC\Users\Abundance\Desktop"
set "Over=\\ABUNDANCEPC\Users\Abundance\Desktop\clipsync.exe"
if exist "%Dest%\" (
    copy /Y "%Home%" "%Dest%"
) else (
    echo Connect your PC to the Other one
    goto end
)

echo Sent Sucessfully
:end

