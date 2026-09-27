# Copies the platform's Terraform outputs into GitHub Actions repository variables.
# Run it after every `./scripts/tf.ps1 platform apply`: re-creating the platform gives the GitHub
# OIDC pool a new ID, so WIF_PROVIDER changes. Requires `gh auth login`.
#
# Usage: ./scripts/sync-gh-vars.ps1

$json = & "$PSScriptRoot/tf.ps1" platform output -json github_actions_variables
if ($LASTEXITCODE -ne 0) {
    Write-Host "Could not read Terraform outputs. Has the platform been applied?" -ForegroundColor Red
    exit 1
}

$vars = ($json -join "`n") | ConvertFrom-Json
foreach ($p in $vars.PSObject.Properties) {
    # Trim: a trailing "\r" from Windows output once turned the image name into an invalid reference.
    $value = "$($p.Value)".Trim()
    gh variable set $p.Name --body $value | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to set $($p.Name)" -ForegroundColor Red
        exit 1
    }
    Write-Host "set $($p.Name) = $value"
}
