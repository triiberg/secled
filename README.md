# SECure LEDger

## Summary

How to keep secrets secret. API Keys, long living tokens etc. Bad practice to keep them in txt file. This software allows to encrypt the secrets and use them in commands. 

Compile this software yourself so you can aware how it works. See compiling options below.

Keep it on an USB stick (recommended) or in your home directory (still better than txt file (in the Git repo :D))

## Use it on your prompt

### Using secled for creating Kubernetes secrets

secret="$(secled get ghcr-password)"          
kubectl -n myapp-sandbox create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=exampleusername \
  --docker-password="$secret" \
  --docker-email=example@example.com
unset secret

### Testing some webhook

TOKEN="$(secled get webhook-token)"
PAYLOAD='{ "instruction": "rollout" }'
SIG=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$TOKEN" | sed 's/^.* //')
curl -X POST http://localhost:8080/deploy/mayapp-sandbox/myapp   -H "X-Hub-Signature-256: sha256=$SIG"   -d "$PAYLOAD"

## How to use it

### Linux/macOS (bash/zsh)
Build:
```sh
go build -o bin/secled ./cmd
```

Add aliases to your `~/.bashrc` or `~/.zshrc`:
```sh
alias secled="~/MYGITHUBDIRS/secled/bin/secled"
alias secled-login='eval "$(secled login)"'
alias secled-logout='eval "$(secled logout)"'
```

Login:
```sh
secled-login
```

List keys (works without password):
```sh
secled list
```

Add a key:
```sh
secled add ghcr-password
```

Update a key:
```sh
secled update ghcr-password
```

Get a key:
```sh
secled get ghcr-password
```

Generate a UUID v4 and store it:
```sh
secled generate-uuid deploy-id
```

Generate a 256-byte JWT secret (base64url) and store it:
```sh
secled generate-256b jwt-secret
```

Remove a key:
```sh
secled remove ghcr-password
```

Logout:
```sh
secled-logout
```

### Windows (PowerShell)
Build:
```powershell
go build -o bin\secled.exe .\cmd
```

Add aliases to your PowerShell profile (create it if missing). Update the path to wherever you cloned `secled`:
```powershell
notepad $PROFILE
```
Add these lines:
```powershell
function secled { & "C:\Users\<user\githubfolders>\secled\bin\secled.exe" @Args }
function secled-login { & secled login | Invoke-Expression }
function secled-logout { & secled logout | Invoke-Expression }
```
Reload your profile:
```powershell
. $PROFILE
```

Login:
```powershell
secled-login
```

List keys:
```powershell
secled list
```

Add a key:
```powershell
secled add ghcr-password
```

Update a key:
```powershell
secled update ghcr-password
```

Get a key:
```powershell
secled get ghcr-password
```

Generate a UUID v4 and store it:
```powershell
secled generate-uuid deploy-id
```

Generate a 256-byte JWT secret (base64url) and store it:
```powershell
secled generate-256b jwt-secret
```

Remove a key:
```powershell
secled remove ghcr-password
```

Logout:
```powershell
secled-logout
```

### Copy to your USB stick
Copy the `bin` directory to your USB drive. The ledger file is stored next to the binary, so keep them together.








