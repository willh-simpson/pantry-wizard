$ingestUrl = "http://localhost:8083/admin/ingest?s="
$categories = @("chicken", "beef", "pasta", "seafood", "dessert")

Write-Host "Starting Bulk Ingestion..." -ForegroundColor Cyan
foreach ($cat in $categories) {
    $res = Invoke-RestMethod -Uri "$($ingestUrl)$cat" -Method Get
    Write-Host "Imported $($res.recipes_imported) recipes for category: $cat" -ForegroundColor Green
}