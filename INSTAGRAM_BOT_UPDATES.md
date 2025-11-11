# Instagram Bot Updates

## Changes Made

### 1. Rate Limit Detection for Following
- **Location**: `api/pkg/instagram/bot/follower_bot.go` - `followAccount()` function
- **What Changed**: Added detection for Instagram's "Try Again Later", "Action Blocked", or "temporarily blocked" alert modals
- **Behavior**: 
  - After clicking follow, waits 2 seconds for modal to appear
  - If rate limit modal detected, closes it and returns error "rate limit detected"
  - This triggers a 2-hour wait period before continuing (handled by existing `RateLimitedUntil` state)

### 2. Silent Rate Limit Detection for Unfollowing  
- **Location**: `api/pkg/instagram/bot/follower_bot.go` - `unfollowAccount()` function
- **What Changed**: Added verification that unfollow actually succeeded
- **Behavior**:
  - After clicking unfollow confirmation, checks if button changed back to "Follow"
  - If button didn't change → unfollow didn't work (silent rate limit)
  - Continues without incrementing unfollow counter
  - No modal appears for unfollow limits (Instagram's behavior)

### 3. Unfollow Enable/Disable Control
- **Location**: `api/internal/crons/instagram/follower.go` - `RunFollowerBot()` function
- **Environment Variable**: `INSTAGRAM_BOT_UNFOLLOW_ENABLED` (default: `true`)
- **What Changed**: Wrapped unfollow logic in a check for this environment variable
- **Behavior**:
  - When `true`: unfollows blacklisted users normally
  - When `false`: skips unfollowing entirely and logs a message
  - Following logic continues regardless

### 4. Random Source Account Selection
- **Location**: `api/internal/crons/instagram/follower.go` - `RunFollowerBot()` function
- **What Changed**: Changed from iterating all source accounts to randomly selecting one per iteration
- **Behavior**:
  - Loads all source accounts from `fftt-instagram-accounts.json`
  - Randomly selects ONE account per bot run
  - Gets followers only from that selected account
  - Logs which account was selected: "🎲 Randomly selected source account: {username}"

## Environment Variables

All variables are already configured in:
- ✅ `docker-compose.yml` 
- ✅ `.env.example`

### Relevant Variables:
```bash
# Enable/disable the entire bot
INSTAGRAM_BOT_ENABLED=true

# Enable/disable unfollowing (new functionality)
INSTAGRAM_BOT_UNFOLLOW_ENABLED=true

# Bot credentials
INSTAGRAM_BOT_USERNAME=your_username
INSTAGRAM_BOT_PASSWORD=your_password
INSTAGRAM_BOT_TOTP_SECRET=ABCD1234EFGH5678IJKL

# Headless mode
INSTAGRAM_BOT_HEADLESS=true
```

## How It Works Together

1. **Bot runs every 30-60 minutes** (random)
2. **Random source selection**: Picks one Instagram account to scrape followers from
3. **Unfollow check**: If enabled, unfollows blacklisted users first
   - Detects silent rate limits (button doesn't change)
4. **Follow new users**: Attempts to follow users from selected source
   - Detects modal alerts for rate limits
   - Sets 2-hour cooldown if detected
5. **Rate limit cooldown**: If triggered, waits 2 hours before next follow attempt
6. **Next iteration**: Selects a different random source account

## Testing Recommendations

1. Monitor logs for:
   - "🎲 Randomly selected source account: {username}"
   - "🚫 Rate limit alert detected" (for following)
   - "⚠️  Unfollow button didn't change" (for silent unfollow limits)
   - "⚠️  Unfollowing is disabled" (when UNFOLLOW_ENABLED=false)

2. Test unfollow disable:
   ```bash
   # In your .env file
   INSTAGRAM_BOT_UNFOLLOW_ENABLED=false
   ```

3. Rate limit behavior should now be much smoother with proper 2-hour waits
