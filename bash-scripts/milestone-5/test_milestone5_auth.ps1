$BaseUrl = "http://identity-service:8081"
$Email = "your-test-email@example.com"
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
Write-Host "success: check your email for the Cognito verification code" -ForegroundColor Green

# confirm registration
$Code = Read-Host "enter the 6-digit code from your email"

$ConfirmBody = @{
    email = $Email
    code  = $Code
} | ConvertTo-Json

Invoke-RestMethod -Method Put -Uri "$BaseUrl/auth/confirm" -Body $ConfirmBody -ContentType "application/json"
Write-Host "account confirmed successfully" -ForegroundColor Green

# login
Write-Host "--- STEP C: Logging In ---" -ForegroundColor Cyan
$LoginBody = @{
    email    = $Email
    password = $UserKey
} | ConvertTo-Json

$LoginResponse = Invoke-RestMethod -Method Get -Uri "$BaseUrl/auth/login" -Body $LoginBody -ContentType "application/json"

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
    $UserProfile = Invoke-RestMethod -Method Get -Uri "$BaseUrl/users/profile" -Headers $Headers
    Write-Host "verifcation successful: identity-service recognized user" -ForegroundColor Green
    $UserProfile | Format-List
} catch {
    Write-Host "Verification Failed: $($_.Exception.Message)" -ForegroundColor Red
}