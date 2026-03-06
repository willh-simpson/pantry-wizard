$EnvPath = "../../.env"

# 1. Load Environment Variables
if (Test-Path $EnvPath) {
    Get-Content $EnvPath | ForEach-Object {
        if ($_ -match "^([^#\s][^=]*)=(.*)$") {
            $Key = $matches[1].Trim()
            $Value = $matches[2].Trim().Trim('"').Trim("'")
            Set-Variable -Name $Key -Value $Value -Scope Global
        }
    }
    Write-Host "--- .env loaded successfully ---" -ForegroundColor Gray
}

$IdentityUrl = "http://localhost:8081"
$UserUrl = "http://localhost:8085"
$Email = "tester-$(Get-Random)@test.local" # Random email to avoid conflict
$UserKey = "PantryWizard123!" 
$DisplayName = "Pantry Hero"

# --- STEP A: Register & Confirm ---
Write-Host "--- STEP A: Registering via Identity Service ---" -ForegroundColor Cyan
$RegBody = @{ email = $Email; password = $UserKey; display_name = $DisplayName } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "$IdentityUrl/auth/register" -Body $RegBody -ContentType "application/json"

aws cognito-idp admin-confirm-sign-up --user-pool-id $COGNITO_USER_POOL_ID --username $Email --region $AWS_REGION
aws cognito-idp admin-update-user-attributes --user-pool-id $COGNITO_USER_POOL_ID --username $Email --user-attributes Name=email_verified,Value=true --region $AWS_REGION

# --- STEP B: Login ---
Write-Host "--- STEP B: Logging In ---" -ForegroundColor Cyan
$LoginBody = @{ email = $Email; password = $UserKey } | ConvertTo-Json
$LoginResponse = Invoke-RestMethod -Method Post -Uri "$IdentityUrl/auth/login" -Body $LoginBody -ContentType "application/json"
$Global:AccessToken = $LoginResponse.access_token
Write-Host "Logged in. Token captured." -ForegroundColor Green

# --- STEP C: Propagation Wait ---
Write-Host "--- STEP C: Waiting for Kafka Sync (2s) ---" -ForegroundColor Yellow
Start-Sleep -Seconds 2

# --- STEP D: Verify User Service Profile ---
Write-Host "--- STEP D: Checking Profile in User Service ---" -ForegroundColor Cyan
$Headers = @{ Authorization = "Bearer $Global:AccessToken" }

try {
    $UserProfile = Invoke-RestMethod -Method Get -Uri "$UserUrl/me/profile" -Headers $Headers
    Write-Host "User Sync Success! Local User Record:" -ForegroundColor Green
    $UserProfile | Format-List
} catch {
    Write-Host "Failed to find user in User Service. Kafka might be slow or consumer failed." -ForegroundColor Red
    return
}

# --- STEP E: Testing Inventory Management ---
Write-Host "--- STEP E: Adding Items to Pantry & Shopping List ---" -ForegroundColor Cyan

$Items = @(
    @{ Path = "pantry/add"; Ingredient = "Onions" },
    @{ Path = "pantry/add"; Ingredient = "Garlic" },
    @{ Path = "shopping-list/add"; Ingredient = "Milk" }
)

foreach ($Item in $Items) {
    $Body = @{ ingredient_name = $Item.Ingredient } | ConvertTo-Json
    Invoke-RestMethod -Method Post -Uri "$UserUrl/me/$($Item.Path)" -Headers $Headers -Body $Body -ContentType "application/json"
    Write-Host "Added $($Item.Ingredient) to $($Item.Path.Split('/')[0])" -ForegroundColor Gray
}

# --- STEP F: Fetch Full Inventory ---
Write-Host "--- STEP F: Fetching Full Combined Inventory ---" -ForegroundColor Cyan
try {
    $Inventory = Invoke-RestMethod -Method Get -Uri "$UserUrl/me/inventory" -Headers $Headers
    Write-Host "Inventory Retrieved:" -ForegroundColor Green
    $Inventory | ConvertTo-Json
} catch {
    Write-Host "Error fetching inventory." -ForegroundColor Red
}

# --- STEP G: Update Preferences ---
Write-Host "--- STEP G: Updating Dietary Flags ---" -ForegroundColor Cyan
$PrefBody = @{ dietary_flags = @("Vegan", "Nut-Free") } | ConvertTo-Json
try {
    Invoke-RestMethod -Method Put -Uri "$UserUrl/me/profile/preferences" -Headers $Headers -Body $PrefBody -ContentType "application/json"
    Write-Host "Preferences Updated!" -ForegroundColor Green
    
    # Verify the update
    $UpdatedProfile = Invoke-RestMethod -Method Get -Uri "$UserUrl/me/profile" -Headers $Headers
    Write-Host "Current Flags: $($UpdatedProfile.dietary_flags -join ', ')" -ForegroundColor Gray
} catch {
    Write-Host "Preference update failed." -ForegroundColor Red
}

Write-Host "`n--- MILESTONE TEST COMPLETE ---" -ForegroundColor Green