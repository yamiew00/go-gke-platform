# Runs Terraform for one stack as a chosen gcloud configuration, without touching this
# machine's default gcloud account or application-default credentials: the Google provider
# and the GCS backend both accept a short-lived token in GOOGLE_OAUTH_ACCESS_TOKEN.
#
# Usage:   ./scripts/tf.ps1 <bootstrap|platform> <terraform arguments...>
# Example: ./scripts/tf.ps1 platform init        (adds -backend-config=backend.hcl for you)
#          ./scripts/tf.ps1 platform plan -out=platform.tfplan
# Set $env:GCLOUD_CONFIG to use a configuration other than "personal".

$stack, $rest = $args
if ($stack -notin @('bootstrap', 'platform')) {
    Write-Host "Usage: ./scripts/tf.ps1 <bootstrap|platform> <terraform arguments...>" -ForegroundColor Yellow
    exit 2
}

# Windows PowerShell splits "-out=plan.tfplan" into "-out=plan" and ".tfplan" before a script
# sees it. Glue such pieces back together so Terraform receives each argument unchanged.
$tfArgs = @()
foreach ($a in $rest) {
    if ($tfArgs.Count -gt 0 -and "$a".StartsWith('.') -and $tfArgs[-1] -like '-*=*') {
        $tfArgs[-1] += "$a"
    } else {
        $tfArgs += "$a"
    }
}

# The platform stack keeps its state in GCS; point init at it so nobody has to remember the flag.
if ($stack -eq 'platform' -and $tfArgs.Count -gt 0 -and $tfArgs[0] -eq 'init' -and -not ($tfArgs -like '-backend-config*')) {
    $tfArgs += '-backend-config=backend.hcl'
}

$config = if ($env:GCLOUD_CONFIG) { $env:GCLOUD_CONFIG } else { 'personal' }
$token = gcloud auth print-access-token --configuration=$config
if ($LASTEXITCODE -ne 0 -or -not $token) {
    Write-Host "Could not get a token for gcloud configuration '$config'." -ForegroundColor Red
    exit 1
}

$env:GOOGLE_OAUTH_ACCESS_TOKEN = $token.Trim()
try {
    terraform "-chdir=$PSScriptRoot/../infra/$stack" @tfArgs
    $code = $LASTEXITCODE
} finally {
    Remove-Item Env:GOOGLE_OAUTH_ACCESS_TOKEN -ErrorAction SilentlyContinue
}
exit $code
