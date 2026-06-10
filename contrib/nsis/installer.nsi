; dcv autosession Installer Script

;--------------------------------
;Includes

    !include "MUI2.nsh"
    !include "LogicLib.nsh"
    !include "x64.nsh"

;--------------------------------
;General
    Name "DCV Autosession"
    OutFile "dcv-autosession-setup.exe"
    Unicode True
    InstallDirRegKey HKCU "Software\dcv-autosession" ""
    InstallDir "$PROGRAMFILES64\dcv-autosession"
    RequestExecutionLevel admin

    SetCompressor /SOLID lzma	; This reduces installer size by approx 30~35%
    ;SetCompressor /FINAL lzma	; This reduces installer size by approx 15~18%
    ; Avoid scaling and blurry text
    ManifestDPIAware true


;--------------------------------
;Version information (passed from build system or defaults to 0.0.0)
    !ifndef VERSION
    !define VERSION "0.0.0"
    !endif
    !ifndef VERSION_NUM
    !define VERSION_NUM "${VERSION}"
    !endif
    VIProductVersion "${VERSION_NUM}.0"
    VIAddVersionKey "ProductName" "DCV Autosession"
    VIAddVersionKey "FileVersion" "${VERSION}"
    VIAddVersionKey "ProductVersion" "${VERSION}"
    VIAddVersionKey "LegalCopyright" "© 2026 Diego Cortassa"
    VIAddVersionKey "FileDescription" "DCV Autosession"

;--------------------------------
;Be sure we are running with admin rights
    Function .onInit
    ; Call UserInfo plugin to get user info.  The plugin puts the result in the stack
    UserInfo::GetAccountType
    # pop the result from the stack into $0
    Pop $0
    ${If} $0 != "admin" ;Require admin rights on NT4+
        MessageBox mb_iconstop "Administrator rights required! Please run as administrator"
        SetErrorLevel 740 ;ERROR_ELEVATION_REQUIRED
        Quit
    ${EndIf}
    FunctionEnd

;--------------------------------
;Modern Interface Settings

  !define MUI_ABORTWARNING
  !define MUI_BGCOLOR "ffffff"
  !define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\orange-install.ico"
  !define MUI_UNICON "${NSISDIR}\Contrib\Graphics\Icons\orange-uninstall.ico"
  !define MUI_HEADERIMAGE_BITMAP "${NSISDIR}\Contrib\Graphics\Header\orange.bmp"
  !define MUI_HEADERIMAGE_UNBITMAP "${NSISDIR}\Contrib\Graphics\Header\orange-uninstall.bmp"
  !define MUI_WELCOMEFINISHPAGE_BITMAP "${NSISDIR}\Contrib\Graphics\Wizard\orange.bmp"
  !define MUI_UNWELCOMEFINISHPAGE_BITMAP "${NSISDIR}\Contrib\Graphics\Wizard\orange-uninstall.bmp"

;--------------------------------
;Pages

    !insertmacro MUI_PAGE_WELCOME
    !insertmacro MUI_PAGE_LICENSE "..\..\LICENSE.md"
    !insertmacro MUI_PAGE_DIRECTORY
    !insertmacro MUI_PAGE_INSTFILES
    ;!insertmacro MUI_PAGE_FINISH

    !insertmacro MUI_UNPAGE_CONFIRM
    !insertmacro MUI_UNPAGE_INSTFILES

;--------------------------------
;Languages
 
    !insertmacro MUI_LANGUAGE "English"

;--------------------------------
;Installer Sections

Section "DCV Autosession" SecMain
    SetOutPath "$INSTDIR"
    
    ; Main executable and configuration
    File "..\..\dist\dcv-autosession-v${VERSION}-windows_amd64\dcv-autosession.exe"
    File "..\..\LICENSE.md"
    
    ; Create the Windows service
    ExecWait '"$INSTDIR\dcv-autosession.exe" --install' $0
    DetailPrint "dcv-autosession.exe returned $0"

    ; Create uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
    
    ; Add uninstall information to Add/Remove Programs
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\dcv-autosession" \
                     "DisplayName" "DCV Autosession"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\dcv-autosession" \
                     "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\dcv-autosession" \
                     "DisplayVersion" "${VERSION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\dcv-autosession" \
                     "Publisher" "dcv-autosession"

SectionEnd

Section "Uninstall"
    ; Stop and remove the service
    ExecWait '"$INSTDIR\dcv-autosession.exe" --uninstall'

    ; Remove installed files
    Delete /REBOOTOK "$INSTDIR\dcv-autosession.exe"
    Delete /REBOOTOK "$INSTDIR\uninstall.exe"
    Delete /REBOOTOK "$INSTDIR\LICENSE.md"

    ; Remove optional config file if it exists
    Delete /REBOOTOK "$INSTDIR\dcv-autosession.conf"

    ; Remove log directory if it exists
    RMDir /r /REBOOTOK "$INSTDIR\log"

    ; Remove install directory ONLY if empty
    RMDir "$INSTDIR"

    ; Remove uninstall information
    DeleteRegKey HKLM \
        "Software\Microsoft\Windows\CurrentVersion\Uninstall\dcv-autosession"
SectionEnd
