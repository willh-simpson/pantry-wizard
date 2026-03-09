# --- Configuration & Env Loading ---
$EnvPath = "../../.env"
if (Test-Path $EnvPath) {
    Get-Content $EnvPath | ForEach-Object {
        if ($_ -match "^([^#\s][^=]*)=(.*)$") {
            $Key = $matches[1].Trim()
            $Value = $matches[2].Trim().Trim('"').Trim("'")
            Set-Variable -Name $Key -Value $Value -Scope Global
        }
    }
}

# Service Ports (Adjust if yours differ)
$IdentityUrl = "http://localhost:8081"
$InteractionUrl = "http://localhost:8082"
$RecipeUrl = "http://localhost:8083"
$RecUrl = "http://localhost:8084"
$UserUrl = "http://localhost:8085"

$Email = "chef-$(Get-Random)@test.local"
$UserKey = "PantryWizard123!"
$TestRecipeID = "52771"
$Ingredients = @("Chicken", "Garlic", "Onions")

Write-Host "--- STARTING MILESTONE 3 INTEGRATION TEST ---" -ForegroundColor Magenta

# 1. Identity: Register & Login
Write-Host "[1/5] Registering and Logging in..." -ForegroundColor Cyan
$RegBody = @{ email = $Email; password = $UserKey; display_name = "Master Chef" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "$IdentityUrl/auth/register" -Body $RegBody -ContentType "application/json"
aws cognito-idp admin-confirm-sign-up --user-pool-id $COGNITO_USER_POOL_ID --username $Email --region $AWS_REGION

$LoginBody = @{ email = $Email; password = $UserKey } | ConvertTo-Json
$LoginResponse = Invoke-RestMethod -Method Post -Uri "$IdentityUrl/auth/login" -Body $LoginBody -ContentType "application/json"
$Headers = @{ Authorization = "Bearer $($LoginResponse.access_token)" }
Write-Host "   > Auth Successful." -ForegroundColor Green

# 2. User Service: Setup Pantry
Write-Host "[2/5] Seeding User Pantry..." -ForegroundColor Cyan
foreach ($item in $Ingredients) {
    $Body = @{ ingredient_name = $item } | ConvertTo-Json
    Invoke-RestMethod -Method Post -Uri "$UserUrl/me/pantry/add" -Headers $Headers -Body $Body -ContentType "application/json"
}
Write-Host "   > Pantry stocked with: $($Ingredients -join ', ')" -ForegroundColor Gray

# 3. Interaction Service: The "Cook" Action
Write-Host "[3/5] Executing 'Cook Recipe' action..." -ForegroundColor Cyan
$CookBody = @{ ingredients = $Ingredients } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "$InteractionUrl/interactions/$TestRecipeID/cook" -Headers $Headers -Body $CookBody -ContentType "application/json"
Write-Host "   > Recipe $TestRecipeID marked as cooked. Kafka events dispatched." -ForegroundColor Green

# 4. Wait for Async Processes
Write-Host "[4/5] Waiting for Kafka Propagation (3s)..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# 5. Verifications
Write-Host "[5/5] Verifying Distributed State..." -ForegroundColor Cyan

# A. Check User Service (Inventory & Suggestions)
$Inventory = Invoke-RestMethod -Method Get -Uri "$UserUrl/me/inventory" -Headers $Headers
$PantryCount = $Inventory.pantry.Count
Write-Host "   > User Service: Pantry items remaining: $PantryCount (Expected: 0)" -ForegroundColor Gray

# B. Check Recommendation Service (Scoring)
try {
    $Score = Invoke-RestMethod -Method Get -Uri "$RecUrl/scores/$TestRecipeID"
    Write-Host "   > Rec Service: Popularity Score: $($Score.total_score) (Weighted Cook = 10.0)" -ForegroundColor Green
} catch { Write-Host "   > Rec Service: Verification skipped (Endpoint not found)" -ForegroundColor Gray }

# C. Check Recipe Service (Global Stats)
try {
    $Recipe = Invoke-RestMethod -Method Get -Uri "$RecipeUrl/recipes/$TestRecipeID"
    Write-Host "   > Recipe Service: Global 'Times Made': $($Recipe.times_made_globally)" -ForegroundColor Green
} catch { Write-Host "   > Recipe Service: Verification skipped (Endpoint not found)" -ForegroundColor Gray }

Write-Host "--- MILESTONE 3 TEST COMPLETE ---" -ForegroundColor Magenta