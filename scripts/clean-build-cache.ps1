# clean-build-cache.ps1
# Clears Go build and test caches to reclaim disk space.
# Run after any development session where go build or go test was invoked.

$goBin = "C:\Program Files\Go\bin\go.exe"

$cacheDir = & $goBin env GOCACHE 2>&1
Write-Host "Build cache: $cacheDir"
$cacheSizeMB = (Get-ChildItem $cacheDir -Recurse -File -ErrorAction SilentlyContinue |
    Measure-Object -Property Length -Sum).Sum / 1MB
Write-Host ("Before: {0:N0} MB" -f $cacheSizeMB)

Write-Host "Cleaning Go build and test caches..."
& $goBin clean -cache -testcache

$cacheSizeAfterMB = (Get-ChildItem $cacheDir -Recurse -File -ErrorAction SilentlyContinue |
    Measure-Object -Property Length -Sum).Sum / 1MB
Write-Host ("After:  {0:N0} MB" -f $cacheSizeAfterMB)
Write-Host ("Reclaimed: {0:N0} MB" -f ($cacheSizeMB - $cacheSizeAfterMB))
