# Copyright 2020-2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Script to test release artifacts on Windows OS
# Inspired by - https://github.com/vmware-tanzu/community-edition/blob/main/test/release-build-test/check-release-build.ps1

param (
    # Tanzu Framework release version argument
    [Parameter(Mandatory=$True)]
    [string]$version,

    # Path to the signtool
    [Parameter(Mandatory=$True)]
    [string]$signToolPath
)

$ErrorActionPreference = 'Stop';

Set-PSDebug -Trace 1

if ((Test-Path env:GITHUB_TOKEN) -eq $False) {
  throw "GITHUB_TOKEN environment variable is not set"
}

$tempFolderPath = Join-Path $Env:Temp $(New-Guid)
New-Item -Type Directory -Path $tempFolderPath

$tanzuCLIPath = Join-Path $tempFolderPath tanzu
New-Item -Type Directory -Path $tanzuCLIPath

$TF_REPO_URL = "https://github.com/vmware-tanzu/tanzu-framework"

gh release download $version --repo $TF_REPO_URL --pattern "tanzu-framework-windows-amd64.zip" --dir $tempFolderPath

Expand-Archive -LiteralPath "$tempFolderPath\tanzu-framework-windows-amd64.zip" -Destination "$tanzuCLIPath"

# Check if the binaries are all signed
 Get-ChildItem -Path $tanzuCLIPath -File -Recurse -Exclude *.yaml | Foreach-Object {
    & $signToolPath verify /pa $_.FullName
    if ($LastExitCode -ne 0) {
        throw "Error verifying: " + $_.FullName
    }
}

& ".\hack\install.bat" "$tanzuCLIPath\cli\core\$version\tanzu-core-windows_amd64.exe" "$version"

$Env:Path += ";C:\Program Files\tanzu"

tanzu version

if ($LastExitCode -ne 0) {
  throw "Error verifying tanzu CLI using version command: " + $_.FullName
}

tanzu management-cluster version

if ($LastExitCode -ne 0) {
  throw "Error verifying tanzu cluster plugin using version command: " + $_.FullName
}

tanzu package version

if ($LastExitCode -ne 0) {
  throw "Error verifying tanzu package plugin using version command: " + $_.FullName
}

tanzu secret version

if ($LastExitCode -ne 0) {
  throw "Error verifying tanzu conformance plugin using version command: " + $_.FullName
}

tanzu login version

if ($LastExitCode -ne 0) {
  throw "Error verifying tanzu login plugin using version command: " + $_.FullName
}

tanzu pinniped-auth version

if ($LastExitCode -ne 0) {
  throw "Error verifying tanzu auth plugin using version command: " + $_.FullName
}
