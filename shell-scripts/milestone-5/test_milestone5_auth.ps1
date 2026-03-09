$EnvPath = "../../.env"

if (Test-Path $EnvPath) {
    Get-Content $EnvPath | ForEach-Object {
        if ($_ -match "^([^#\s][^=]*)=(.*)$") {
            $Key = $matches[1].Trim()
            $Value = $matches[2].Trim().Trim('"').Trim("'")
            Set-Variable -Name $Key -Value $Value -Scope Global
        }
    }

    Write-Host "--- .env loaded successfully ---" -ForegroundColor Gray
} else {
    Write-Host "--- warning: .env not found at $EnvPath ---" -ForegroundColor Yellow
}

$BaseUrl = "http://localhost:8081"
$Email = "dummy-email@test.local"
$UserKey = "PantryWizard123!" # workaround for detect-secrets since this is not a real user
$DisplayName = "pantry-wizard user"

# sign up
Write-Host "--- STEP A: Registering User ---" -ForegroundColor Cyan
$RegBody = @{
    email        = $Email
    password     = $UserKey
    display_name = $DisplayName
} | ConvertTo-Json

$RegResponse = Invoke-RestMethod -Method Post -Uri "$BaseUrl/auth/register" -Body $RegBody -ContentType "application/json"
$RegResponse | Format-Table
Write-Host "success: moving to confirm user (normally 6-digit code sent to email)" -ForegroundColor Green

# confirm registration
Write-Host "--- STEP B: Admin Force-Confirming User ---" -ForegroundColor Cyan

if (-not $COGNITO_USER_POOL_ID) {
    Write-Host "error: COGNITO_USER_POOL_ID not found in .env" -ForegroundColor Red
    exit
}

# aws CLI command
aws cognito-idp admin-confirm-sign-up `
    --user-pool-id $COGNITO_USER_POOL_ID `
    --username $Email `
    --region $AWS_REGION

if ($LASTEXITCODE -eq 0) {
    Write-Host "user $Email successfully confirmed via Admin API" -ForegroundColor Green
} else {
    Write-Host "failed to confirm user: make sure AWS CLI is configured" -ForegroundColor Red
    exit
}

Write-Host "--- STEP B.5: Verifying Email Attribute ---" -ForegroundColor Cyan
aws cognito-idp admin-update-user-attributes `
    --user-pool-id $COGNITO_USER_POOL_ID `
    --username $Email `
    --user-attributes Name=email_verified,Value=true `
    --region $AWS_REGION

if ($LASTEXITCODE -eq 0) {
    Write-Host "email verified" -ForegroundColor Green
}

# login
Write-Host "--- STEP C: Logging In ---" -ForegroundColor Cyan
$LoginBody = @{
    email    = $Email
    password = $UserKey
} | ConvertTo-Json

$LoginResponse = Invoke-RestMethod -Method Post -Uri "$BaseUrl/auth/login" -Body $LoginBody -ContentType "application/json"

# Save the token for the next request
$Global:AccessToken = $LoginResponse.access_token

Write-Host "logged in: token captured" -ForegroundColor Green
$LoginResponse.user | Format-List

# access protected profile
Write-Host "--- STEP D: Testing Auth Middleware ---" -ForegroundColor Cyan

$Headers = @{
    Authorization = "Bearer $Global:AccessToken"
}

try {
    $UserProfile = Invoke-WebRequest -Method Get -Uri "$BaseUrl/users/profile" -Headers $Headers -UseBasicParsing
    Write-Host "Success!" -ForegroundColor Green
    $UserProfile.Content | ConvertFrom-Json | Format-List
} catch {
    $rawResponse = $_.Exception.Response
    if ($rawResponse) {
        $reader = New-Object System.IO.StreamReader($rawResponse.GetResponseStream())
        $body = $reader.ReadToEnd()
        Write-Host "CRITICAL SERVER ERROR: $body" -ForegroundColor Red
    } else {
        Write-Host "No response from server. Is the service down?" -ForegroundColor Red
    }
}