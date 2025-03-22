# Hashicorp vault integration

As a keystore backend for the **EdgeCloud** we'll be using Hashicorp Vault. 
This will allow us to store secrets and other sensitive information securely. I'll be baking support for it into the cli.

## Features

- [ ] Save secrets
- [ ] Fetch secrets


## Install the cli

- **windows:**
    ```powershell
    choco install vault
    ```

## Set env vars

- **windows:**
    ```powershell
    $env:VAULT_ADDR = "https://edgevault.duckdns.org"
    $env:VAULT_TOKEN = ""
    ```

## RKE2 integration