$UserService = "http://localhost:8085"
$RecipeService = "http://localhost:8083"
$TestUserId = "550e8400-e29b-41d4-a716-446655440000" 

Write-Host "`n--- STEP 1: Populating User Inventory ---" -ForegroundColor Cyan
$pantryBody = @{ items = @("chicken", "salt", "pepper", "onion", "garlic", "oil") } | ConvertTo-Json
Invoke-RestMethod -Uri "$UserService/users/$TestUserId/pantry/add" -Method Post -Body $pantryBody -ContentType "application/json"

$shoppingBody = @{ items = @("chicken", "basmati rice, bay leaf, cinnamon") } | ConvertTo-Json
Invoke-RestMethod -Uri "$UserService/users/$TestUserId/shopping-list/add" -Method Post -Body $shoppingBody -ContentType "application/json"

$wishlistBody = @{ items = @("chicken") } | ConvertTo-Json
Invoke-RestMethod -Uri "$UserService/users/$TestUserId/wishlist/add" -Method Post -Body $wishlistBody -ContentType "application/json"

Write-Host "`n--- STEP 1.5: Verifying User Service Data ---" -ForegroundColor Yellow
$inventory = Invoke-RestMethod -Uri "$UserService/users/$TestUserId/inventory"
Write-Host "Pantry Items: $($inventory.pantry -join ', ')"
Write-Host "Wishlist Items: $($inventory.wishlist -join ', ')"

Write-Host "`n--- STEP 2: Testing Pantry Search (Broad) ---" -ForegroundColor Cyan
$pantryResults = Invoke-RestMethod -Uri "$RecipeService/recipes/search?user_id=$TestUserId&mode=pantry&strictness=unrestricted"
if ($pantryResults) { $pantryResults | Format-Table -AutoSize } else { Write-Host "No results found." -ForegroundColor Red }

Write-Host "`n--- STEP 4: Moving ['basmati rice', 'bay leaf', 'cinnamon'] ---" -ForegroundColor Cyan
$moveBody = @{ items = @("basmati rice", "bay leaf", "cinnamon") } | ConvertTo-Json
Invoke-RestMethod -Uri "$UserService/users/$TestUserId/pantry/move" -Method Post -Body $moveBody -ContentType "application/json"

Write-Host "`n--- STEP 5: Verify Pantry Search updated ---" -ForegroundColor Cyan
$updatedResults = Invoke-RestMethod -Uri "$RecipeService/recipes/search?user_id=$TestUserId&mode=pantry"
if ($updatedResults) { $updatedResults | Format-Table -AutoSize } else { Write-Host "No results found." -ForegroundColor Red }

Write-Host "`nMilestone 4 Testing Complete." -ForegroundColor Magenta