$baseUrl = "http://localhost:8082/interactions"
$recUrl = "http://localhost:8084/recommendations"

# valid UUIDs for testing
$recipeA = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
$recipeB = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
$recipeC = "cccccccc-cccc-cccc-cccc-cccccccccccc"
$userId = "11111111-1111-1111-1111-111111111111"

function Send-Interaction($rid, $action) {
    $body = @{
        recipe_id = $rid
        user_id = $userId
        action = $action
        timestamp = [DateTimeOffset]::Now.ToUnixTimeSeconds()
    } | ConvertTo-Json
    
    Write-Host "Sending $action for $rid..." -ForegroundColor Cyan
    Invoke-RestMethod -Uri $baseUrl -Method Post -Body $body -ContentType "application/json"
}


# recipe A gets 5 likes (Score: 5.0)
1..5 | ForEach-Object { Send-Interaction $recipeA "like" }

# recipe B gets 2 saves (Score: 6.0) 
1..2 | ForEach-Object { Send-Interaction $recipeB "save" }

# recipe C gets a like and an unlike (Score: 0.0)
Send-Interaction $recipeC "like"
Send-Interaction $recipeC "unlike"

Write-Host "`nWaiting 10 seconds for Kafka & DB to settle..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# fetch ranking
Write-Host "Fetching Top Recommendations..." -ForegroundColor Green
$results = Invoke-RestMethod -Uri "$($recUrl)?limit=5"
$results | Format-Table -AutoSize