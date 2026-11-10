; Inno Setup script for SourceSight.
; Built by .github/workflows/release.yml:
;   ISCC /DAppVersion=2.1.0 /DDist=..\dist installer\SourceSight.iss
;
; Installs per user (no admin prompt) into %LOCALAPPDATA%\Programs\SourceSight, so the
; in-app updater can replace SourceSight.exe without elevation.

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef Dist
  #define Dist "..\dist"
#endif

[Setup]
AppId={{6F1D3C2A-8B54-4E7A-9C61-5D2B7E0A4F93}
AppName=SourceSight
AppVersion={#AppVersion}
AppVerName=SourceSight {#AppVersion}
AppPublisher=Hdauu
AppPublisherURL=https://github.com/Hdauu/SourceSight
AppSupportURL=https://github.com/Hdauu/SourceSight/issues
AppUpdatesURL=https://github.com/Hdauu/SourceSight/releases
DefaultDirName={localappdata}\Programs\SourceSight
DefaultGroupName=SourceSight
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
OutputDir=..\
OutputBaseFilename=SourceSight-Setup-v{#AppVersion}
SetupIconFile=SourceSight.ico
UninstallDisplayIcon={app}\SourceSight.exe
UninstallDisplayName=SourceSight
LicenseFile={#Dist}\LICENSE
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
CloseApplications=yes
RestartApplications=no

[Languages]
Name: "es"; MessagesFile: "compiler:Languages\Spanish.isl"
Name: "en"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
Source: "{#Dist}\SourceSight.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#Dist}\SourceSightCLI.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#Dist}\SourceSightBot.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#Dist}\LICENSE"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#Dist}\README.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\SourceSight"; Filename: "{app}\SourceSight.exe"
Name: "{autodesktop}\SourceSight"; Filename: "{app}\SourceSight.exe"; Tasks: desktopicon

[Registry]
; "Open with SourceSight" on .dem files, without taking over the default handler
Root: HKCU; Subkey: "Software\Classes\.dem\OpenWithProgids"; ValueType: string; ValueName: "SourceSight.dem"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\SourceSight.dem"; ValueType: string; ValueName: ""; ValueData: "Demo de Counter-Strike"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\SourceSight.dem\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\SourceSight.exe,0"
Root: HKCU; Subkey: "Software\Classes\SourceSight.dem\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\SourceSight.exe"" ""%1"""

[Run]
Filename: "{app}\SourceSight.exe"; Description: "{cm:LaunchProgram,SourceSight}"; Flags: nowait postinstall skipifsilent
