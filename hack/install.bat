:: Copyright 2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
:: SPDX-License-Identifier: Apache-2.0

:: Inspired by - https://github.com/vmware-tanzu/community-edition/blob/main/hack/install.bat
:: Script to install tanzu framework 
:: Usage: .\hack\install.bat \path\to\tanzu-framework-binary v0.10.0

@echo off

:: start copy tanzu cli
SET TANZU_CLI_DIR=%ProgramFiles%\tanzu
mkdir "%TANZU_CLI_DIR%" || goto :error
copy /B /Y %1\cli\core\%2\tanzu-core-windows_amd64.exe "%TANZU_CLI_DIR%\tanzu.exe" || goto :error

:: set cli path
set PATH=%PATH%;%TANZU_CLI_DIR%
:: setx /M path "%path%;%TANZU_CLI_DIR%"
:: end copy tanzu cli

:: set plugin path
SET PLUGIN_DIR="%LocalAppData%\tanzu-cli"
SET TANZU_CACHE_DIR="%LocalAppData%\.cache\tanzu"
mkdir %PLUGIN_DIR% || goto :error
:: delete the plugin cache if it exists, before installing new plugins
rmdir /Q /S %TANZU_CACHE_DIR%

:: install plugins
tanzu plugin sync || goto :error
tanzu plugin list || goto :error

echo "Installation complete!"
echo "Please add %TANZU_CLI_DIR% permanently into your system's PATH."
goto :EOF

:error
echo Failed with error #%errorlevel%.
exit /b %errorlevel%
