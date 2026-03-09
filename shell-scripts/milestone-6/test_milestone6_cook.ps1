function Wait-For-Data {
    param (
        [string]$Name,
        [scriptblock]$Action,
        [int]$MaxRetries = 20
    )
    Write-Host "Waiting for $Name..." -NoNewline -ForegroundColor Yellow
    for ($i = 0; $i -lt $MaxRetries; $i++) {
        try {
            $result = &$Action
            if ($result) {
                Write-Host " [OK]" -ForegroundColor Green
                return $result
            }
        } catch { }
        Write-Host "." -NoNewline
        Start-Sleep -Milliseconds 500
    }
    Write-Host " [TIMEOUT]" -ForegroundColor Red
    return $null
}

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
$Ingredients = @("Chicken", "Garlic", "Onions")

Write-Host "--- STARTING MILESTONE 3 INTEGRATION TEST ---" -ForegroundColor Magenta

# 1. Identity: Register & Login
Write-Host "[1/6] Registering and Logging in..." -ForegroundColor Cyan
$RegBody = @{ email = $Email; password = $UserKey; display_name = "Master Chef" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "$IdentityUrl/auth/register" -Body $RegBody -ContentType "application/json"
aws cognito-idp admin-confirm-sign-up --user-pool-id $COGNITO_USER_POOL_ID --username $Email --region $AWS_REGION

$LoginBody = @{ email = $Email; password = $UserKey } | ConvertTo-Json
$LoginResponse = Invoke-RestMethod -Method Post -Uri "$IdentityUrl/auth/login" -Body $LoginBody -ContentType "application/json"
$Headers = @{ Authorization = "Bearer $($LoginResponse.access_token)" }
Write-Host "   > Auth Successful." -ForegroundColor Green

# wait for user sync
Wait-For-Data "User Profile Sync" {
    Invoke-RestMethod -Method Get -Uri "$UserUrl/me/inventory" -Headers $Headers
}

Write-Host "[2/6] Creating a Dummy Recipe..." -ForegroundColor Cyan
$CreateRecipeBody = @{
    title = "Kafka Chicken"
    description = "desc"
    instructions = "Cook with code"
    author_id = [guid]::NewGuid().ToString()
    prep_time_min = 5
    calories = 500
    budget_tier = 1
    ingredients = @() 
} | ConvertTo-Json

# Create the recipe in the recipe-service
$CreateRecipeResponse = Wait-For-Data "Creating Dummy Recipe" {
    Invoke-RestMethod -Method Post -Uri "$RecipeUrl/recipes/" -Headers $Headers -Body $CreateRecipeBody -ContentType "application/json"
}
$RecipeID = $CreateRecipeResponse.id
Write-Host "   > Recipe created with ID: $RecipeId" -ForegroundColor Green

# 3. User Service: Setup Pantry
Write-Host "[3/6] Seeding User Pantry..." -ForegroundColor Cyan
$Body = @{ items = $Ingredients } | ConvertTo-Json
Wait-For-Data "verifying user sync" {
    Invoke-RestMethod -Method Post -Uri "$UserUrl/me/pantry/add" -Headers $Headers -Body $Body -ContentType "application/json"
}
Write-Host "   > Pantry stocked with: $($Ingredients -join ', ')" -ForegroundColor Gray

# 4. Interaction Service: The "Cook" Action
Write-Host "[4/7] Executing 'Cook Recipe' action..." -ForegroundColor Cyan
$CookBody = @{ ingredients = $Ingredients } | ConvertTo-Json
Wait-For-Data "verifying user sync" {
    Invoke-RestMethod -Method Post -Uri "$InteractionUrl/interactions/$RecipeID/cook" -Headers $Headers -Body $CookBody -ContentType "application/json"
}
Write-Host "   > Recipe $RecipeID marked as cooked. Kafka events dispatched." -ForegroundColor Green

Write-Host "[5/6] Waiting for Distributed State Consistency..." -ForegroundColor Cyan

# A. Poll Recommendation Service: Score should be 10.0 for a Cook
$ScoreData = Wait-For-Data "Rec Service (Scoring Update)" {
    Invoke-RestMethod -Method Get -Uri "$RecUrl/scores/$RecipeID" -Headers $Headers
    # if ($score.total_score -ge 8) { return $score }
    # return $null
}

# B. Poll Recipe Service: Global 'Times Made' should be 1
$RecipeData = Wait-For-Data "Recipe Service (Global Stats)" {
    $recipe = Invoke-RestMethod -Method Get -Uri "$RecipeUrl/recipes/$RecipeID" -Headers $Headers
    if ($recipe.times_made_globally -gt 0) { return $recipe }
    return $null
}

# 7. Final Report
Write-Host "--- [6/6] VERIFICATION COMPLETE ---" -ForegroundColor Magenta
Write-Host "Rec Score:    $($ScoreData.score)"
Write-Host "Global Cooks: $($RecipeData.times_made_globally)"

Write-Host "--- MILESTONE 3 TEST COMPLETE ---" -ForegroundColor Magenta