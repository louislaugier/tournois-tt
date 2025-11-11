# Instagram Follow Bot - Docker Path Fix

## Problem
The `make test-instagram-follow` command was failing with an error because the bot couldn't find the required JSON configuration files:
- `fftt-instagram-accounts.json`
- `instagram_blacklist.json`

This was happening because when running `go run` inside Docker, the working directory wasn't correctly set to where these files are located.

## Solution
Fixed the path resolution in three places:

### 1. Test Command (`cmd/test-instagram-follow/main.go`)
- Added fallback path resolution to try multiple possible file locations
- The code now tries these paths in order:
  1. `fftt-instagram-accounts.json` (current directory)
  2. `/go/src/tournois-tt/api/fftt-instagram-accounts.json` (absolute Docker path)
  3. `./api/fftt-instagram-accounts.json` (relative path)
- Added logging to show which path was successfully found

### 2. Cron Job (`internal/crons/instagram/follower.go`)
- Applied the same fallback path resolution for the automated cron jobs
- This ensures the bot works both in tests and when running automatically via crons

### 3. Makefile
- Updated the `test-instagram-follow` target to explicitly set the working directory:
  ```makefile
  docker-compose exec -w /go/src/tournois-tt/api api go run cmd/test-instagram-follow/main.go
  ```
- The `-w` flag ensures the command runs from the correct directory

## Testing
To test the fix:

```bash
# Test the Instagram follow bot
make test-instagram-follow

# The bot should now:
# ✅ Find and load fftt-instagram-accounts.json
# ✅ Find and load instagram_blacklist.json
# ✅ Run the follow/unfollow logic successfully
```

## Files Modified
1. `/api/cmd/test-instagram-follow/main.go`
   - Updated `getSourceAccounts()` function
   - Updated `getAccountsToUnfollow()` function
   - Added logging for loaded file paths

2. `/api/internal/crons/instagram/follower.go`
   - Updated `loadSourceAccounts()` function
   - Updated `loadBlacklist()` function
   - Added logging for loaded file paths

3. `/Makefile`
   - Updated `test-instagram-follow` target to set working directory

## Benefits
- ✅ Works in Docker containers with any working directory
- ✅ Works in local development
- ✅ Works in automated cron jobs
- ✅ Clear logging shows which path was used
- ✅ No need to modify docker-compose.yml or environment variables
