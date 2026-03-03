$baseUrl = "http://localhost:8083/recipes/search"

$myPantry = "chicken,salt,pepper,onion,garlic,oil,water,basmati rice,cinnamon,coriander,cloves,cardamom,green chilli,turmeric powder,bay leaf,ginger,spring onions"
$myShoppingList = "chicken,salt"
$myWishlist = "chicken" # every result must have chicken

function Test-Search($testName, $mode, $strictness, $pantry, $shopping, $wish) {
    Write-Host "`n--- TEST: $testName ---" -ForegroundColor Cyan
    Write-Host "Mode: $mode | Strictness: $strictness | Wishlist: $wish" -ForegroundColor Gray
    
    $uri = "$($baseUrl)?mode=$mode&strictness=$strictness&pantry=$pantry&shopping_list=$shopping&wishlist=$wish"
    
    try {
        $results = Invoke-RestMethod -Uri $uri -Method Get
        if ($results.Count -eq 0) {
            Write-Host "No recipes found matching these criteria." -ForegroundColor Yellow
        } else {
            $results | Select-Object title, match_count, total_needed, stars | Format-Table -AutoSize
        }
    } catch {
        $err = $_.Exception.Message
        Write-Host "Error: $err" -ForegroundColor Red
    }
}

# pantry Mode - unrestricted
# 10% match still required
Test-Search "Broad Pantry Search" "pantry" "unrestricted" $myPantry "" $myWishlist

# pantry Mode - less_strict
# 70% match required
Test-Search "Efficient Pantry Search" "pantry" "less_strict" $myPantry "" $myWishlist

# shopping list mode - strict
# expect 0 results
Test-Search "Minimalist Shopping Search" "shopping_list" "strict" "" $myShoppingList $myWishlist

# beef in wishlist should result in no matches for chicken dishes
Test-Search "Wishlist Exclusion Test" "pantry" "unrestricted" $myPantry "" "beef"