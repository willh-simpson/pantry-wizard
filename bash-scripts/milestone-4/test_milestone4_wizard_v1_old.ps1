$searchUrl = "http://localhost:8083/recipes/search"

# user has chicken and salt
$myPantry = "Chicken"

function Test-Search($name, $strictness) {
    Write-Host "`nTesting Scenario: $name ($strictness)" -ForegroundColor Cyan
    $uri = "$($searchUrl)?ingredients=$myPantry&strictness=$strictness"
    $results = Invoke-RestMethod -Uri $uri -Method Get
    
    if ($results.Count -eq 0) {
        Write-Host "No results found." -ForegroundColor Yellow
    } else {
        $results | Select-Object title, match_count, total_needed, stars | Format-Table -AutoSize
    }
}

# 1. strict: only gets recipes with chicken and salt as only ingredients (will almost certainly return 0)
Test-Search "Strict Mode" "strict"

# 2. less_trict: returns recipes where user has ~70% of needed ingredients
Test-Search "Less Strict Mode" "less_strict"

# 3. unrestricted: returns anything including chicken and salt
Test-Search "Unrestricted Mode" "unrestricted"