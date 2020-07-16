; -- Example1.iss --
; Demonstrates copying 3 files and creating an icon.

; SEE THE DOCUMENTATION FOR DETAILS ON CREATING .ISS SCRIPT FILES!      

[Setup]
AppName= e掌通智控
AppVersion=v1.1
WizardStyle=modern
DefaultDirName=D:\e掌通智控
DefaultGroupName=e掌通智控
UninstallDisplayIcon={app}\Attendance.exe
Compression=lzma2
SolidCompression=yes
OutputDir=userdocs:Inno Setup Examples Output
SetupIconFile = 

[Components]
Name: Server; Description:"服务端";Types:full
Name: Client;Description:"客户端";Types:full

[Files]
Source: "beegoServer.db"; DestDir: "{app}" ;Components: Server ;  Flags: onlyifdoesntexist  uninsneveruninstall
Source: "InstallServer.bat"; DestDir: "{app}" ;Components: Server
Source: "Uninstallserver.bat"; DestDir: "{app}";Components: Server
Source: "使用说明.docx"; DestDir: "{app}";Components: Server
Source: "web.exe"; DestDir: "{app}" ;Components: Server
Source:"conf\*"; DestDir: "{app}/conf";Components: Server
Source:"manager\*"  ;  DestDir: "{app}/manager";Components: Server
Source:"manager\static\css\*";      DestDir: "{app}/manager/static/css";Components: Server
Source:"manager\static\fonts\*";      DestDir: "{app}/manager/static/fonts";Components: Server
Source:"manager\static\img\*";      DestDir: "{app}/manager/static/img";Components: Server
Source:"manager\static\js\*";      DestDir: "{app}/manager/static/js";Components: Server
Source:"views\*";      DestDir: "{app}/views";Components: Server
Source:"webClient.exe"  ;DestDir: "{app}";Components: Client

[run]
Filename: "{app}/InstallServer.bat";Components: Server ; Flags:runhidden;

[UninstallRun]
Filename: "{app}/Uninstallserver.bat";Components: Server ;Flags:runhidden;


[Icons]
Name: "{commondesktop}\e掌通智控"; Filename: "{app}\webClient.exe";Components: Client




