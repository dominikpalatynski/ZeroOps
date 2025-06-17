# ZeroOps

ZeroOps is a simple CLI tool to deploy applications to your VPS using Docker Compose. It helps you manage deployment contexts, render docker-compose templates with environment variables, and automate remote deployments.

## Features
- Manage multiple VPS deployment contexts
- Render and upload docker-compose files with environment variables
- Deploy, check status, list, and remove applications on your VPS

## Prerequisites
- Docker and Docker Compose installed on your VPS
- SSH access to your VPS
- `scp` and `ssh` available on your local machine

## MacOS/Linus Installation
You can install ZeroOps using the provided install script:

```sh
curl -sL https://raw.githubusercontent.com/DominikPalatynski/zeroops/main/install.sh | bash
```

This will download the latest release and install it to `~/.local/bin/zeroops`. Make sure `~/.local/bin` is in your `PATH`.

### Windows Installation

1. Go to the [Releases page](https://github.com/DominikPalatynski/zeroops/releases) and download the latest `zeroops_<version>_windows_amd64.zip` (or `arm64.zip` if needed).
2. Extract the `zeroops.exe` file.
3. Add the folder containing `zeroops.exe` to your system `PATH`.
4. You can now use `zeroops` from Command Prompt or PowerShell.

## Usage

### 1. Configure a VPS Context
A context defines the VPS you want to deploy to.

Add a new context:
```sh
zeroops context add my-vps --docker "host=ssh://user@your-vps-ip"
```

List available contexts:
```sh
zeroops context list
```

Set the current context:
```sh
zeroops context use --name my-vps
```

Show the current context:
```sh
zeroops context current
```

Remove a context:
```sh
zeroops context rm --name my-vps
```

### 2. Prepare Your Environment
Create a `.env` file in your project directory with the required environment variables. For example:

```
RESEND_API_KEY=your_api_key_here
```

The variables in `.env` will be injected into `docker-compose.tpl.yml`.

### 3. Prepare Your Docker Compose Template
Edit `docker-compose.tpl.yml` to define your services. Example:

```yaml
version: "3"
services:
  nextjs:
    image: <your-image>
    ports:
      - "3004:3000"
    environment:
      - RESEND_API_KEY={{ .RESEND_API_KEY }}
```

### 4. Deploy an Application
Deploy your app to the current context:
```sh
zeroops deploy add my-app
```
This will:
- Render `docker-compose.tpl.yml` with your `.env` values
- Upload the rendered `docker-compose.yml` and `.env` to your VPS
- Run `docker compose up --build -d` remotely

### 5. Manage Deployments
Check status:
```sh
zeroops deploy status my-app
```

List all deployed apps:
```sh
zeroops deploy list
```

Remove an app:
```sh
zeroops deploy rm my-app
```

## Troubleshooting
If you see errors about `sudo` requiring a password for remote commands, you may need to allow passwordless sudo for your user on the VPS. You can do this by running:

```sh
sudo visudo
```
And adding a line like:
```
youruser ALL=(ALL) NOPASSWD:ALL
```
Replace `youruser` with your VPS username.