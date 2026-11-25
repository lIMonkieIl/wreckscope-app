Unicode true

####
## WreckScope themed NSIS (Wails v3 layout under build/windows/nsis).
## Stock macros from wails_tools.nsh; branding + AppData/exe fixes retained from v2.
####
## !define INFO_PROJECTNAME    "MyProject"
## !define INFO_COMPANYNAME    "MyCompany"
## !define INFO_PRODUCTNAME    "MyProduct"
## !define INFO_PRODUCTVERSION "1.0.0"
## !define INFO_COPYRIGHT      "Copyright"
###
## !define REQUEST_EXECUTION_LEVEL "admin"
## !define WAILS_INSTALL_SCOPE     "user"
####
# Match output binary name.
!define PRODUCT_EXECUTABLE "WreckScope.exe"
####
!include "wails_tools.nsh"

VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"

# WreckScope brand: #111820 / #52C3FF / #007DAA
!define MUI_BGCOLOR 111820
!define MUI_TEXTCOLOR FFFFFF
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_BITMAP "resources\header.bmp"
!define MUI_HEADERIMAGE_RIGHT
!define MUI_WELCOMEFINISHPAGE_BITMAP "resources\welcome.bmp"
!define MUI_UNWELCOMEFINISHPAGE_BITMAP "resources\welcome.bmp"
!define MUI_INSTFILESPAGE_COLORS "FFFFFF 111820"
!define MUI_LICENSEPAGE_BGCOLOR FFFFFF
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!define MUI_WELCOMEPAGE_TITLE "Welcome to $(^NameDA) Setup"
!define MUI_WELCOMEPAGE_TEXT "Setup will install $(^NameDA) on your computer.$\r$\n$\r$\nClick Next to continue."

!define MUI_FINISHPAGE_TITLE "$(^NameDA) is installed"
!define MUI_FINISHPAGE_TEXT "$(^NameDA) has been installed on your computer.$\r$\n$\r$\nClick Finish to close Setup."
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "Run $(^NameDA)"
!define MUI_FINISHPAGE_SHOWREADME ""
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Create desktop shortcut"
!define MUI_FINISHPAGE_SHOWREADME_FUNCTION finish.CreateDesktopShortcut
!define MUI_FINISHPAGE_LINK_COLOR 52C3FF

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "..\..\..\LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!define MUI_PAGE_CUSTOMFUNCTION_SHOW finish.FixCheckboxColors
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

BrandingText "$(^Name)"

#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
!if "${WAILS_INSTALL_SCOPE}" == "user"
  InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
  # Flat install dir — avoid Program Files\WreckScope\WreckScope nesting.
  InstallDir "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
!endif
ShowInstDetails show

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Function finish.CreateDesktopShortcut
  SetShellVarContext current
  CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
FunctionEnd

Function finish.FixCheckboxColors
  GetDlgItem $R0 $MUI_HWND 1203
  System::Call 'uxtheme::SetWindowTheme(p$R0,w" ",w" ")'
  SetCtlColors $R0 FFFFFF 111820
  GetDlgItem $R0 $MUI_HWND 1204
  System::Call 'uxtheme::SetWindowTheme(p$R0,w" ",w" ")'
  SetCtlColors $R0 FFFFFF 111820
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    # Stop a running instance so files/folders are not locked.
    nsExec::ExecToLog 'taskkill /F /IM ${PRODUCT_EXECUTABLE}'
    Pop $0
    Sleep 500

    # User data is always under the current user's Roaming AppData.
    # Machine installs use SetShellVarContext all, which would make $APPDATA
    # point at ProgramData and miss %AppData%\WreckScope entirely.
    SetShellVarContext current
    RMDir /r "$APPDATA\${INFO_PRODUCTNAME}"

    # Install tree: current flat path, plus legacy nested company\product.
    RMDir /r $INSTDIR
    RMDir /r "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
    RMDir "$PROGRAMFILES64\${INFO_COMPANYNAME}"

    # Start Menu shortcut is all-users on machine installs.
    SetShellVarContext all
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    # Finish-page desktop shortcut was created with current-user context.
    SetShellVarContext current
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
