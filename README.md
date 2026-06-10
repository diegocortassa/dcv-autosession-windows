> [!WARNING]
> This project is in beta stage. It is advised that you do not use this in production
> unless you are willing to encounter bugs, paper cuts or your computer on fire along the way.

# Autosession for DCV
Autosession for DCV is a service that creates a simple standalone client-server architecture where DCV session lifecycle is automatically managed without additional running services, management interfaces, or servers.

# How it works
The service constantly monitors the DCV server log for DCV viewer connection requests. It looks for the log at `C:\ProgramData\Amazon\dcv\log\server.log` first, falling back to `C:\ProgramData\NICE\dcv\log\server.log` if the Amazon path does not exist. The log path can be overridden with the `dcvserver_log` key in the configuration file. When a connection request is detected, it creates a new Console session. When no more connections are active, the session is automatically closed.

The configuration file `dcv-autosession.conf` is loaded from the installation directory (`C:\Program Files\dcv-autosession` by default), or a custom path can be specified with the `--conf` flag. Autosession logs are written to `C:\ProgramData\dcv-autosession\logs` by default, or configured under the `[log]` section.

# Limitations
For timing reasons, the session must be created as soon as the DCV viewer connects. Waiting for user authentication is too late and the viewer will not be able to connect. Therefore, the session is created under the `Administrator` user, even though the Windows logon is performed by the real authenticated user without any Administrator privileges. To allow users to connect to a session owned by `Administrator`, the following line must be added to the `[permissions]` section of `C:\Program Files\NICE\DCV\Server\conf\default.perm`:
```
%any% allow builtin
```

# Prerequisites
This service is intended for Windows 10 and later.

DCV server must be installed and, for security, configured to allow only one concurrent session and to automatically lock the session when the user disconnects. To do this, run the following commands in an elevated command prompt:
```powershell
reg add HKEY_USERS\S-1-5-18\Software\GSettings\com\nicesoftware\dcv\session-management /v max-concurrent-clients /t REG_DWORD /d 1 /f
reg add HKEY_USERS\S-1-5-18\Software\GSettings\com\nicesoftware\dcv\security /v os-auto-lock /t REG_DWORD /d 1 /f
```
These two registry keys are important because permissions must be relaxed to allow any user to connect to a session owned by `Administrator`. If multiple concurrent sessions are allowed, users may connect to the wrong session. If auto-lock is disabled, users may connect to a session that is not locked and see the desktop of the previous user.
Edit `C:\Program Files\NICE\DCV\Server\conf\default.perm` and add the following to the `[permissions]` section:
```
%any% allow builtin
```


# Install
Use the provided installer.

# Uninstall
Open "Add or remove programs" and uninstall "DCV Autosession for Windows".

# Build the installer

Install the Go programming language, the make utility and the NSIS installer system.
Run the following command to build the installer:
``` bash
make installer
```

## Releases

The project uses GitHub Actions to automatically build and publish releases. When a new version is ready:

1. Update CHANGELOG.md
2. Commit all changes
3. Create and push a new tag:
```bash
git tag v<version> -m "Release version <version>"
git push origin v<version>
```

The GitHub Actions workflow will automatically:
- Create a source tarball (.tar.gz)
- Build the installer
- Create a GitHub release with all assets attached
- Generate release notes from commits

The following assets will be available in each release:
- Source code (tar.gz)
- Installer for Windows
- A zip file containing the executable and configuration file for manual installation
