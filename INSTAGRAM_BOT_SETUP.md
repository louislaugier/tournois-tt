# Instagram Bot - Setup and Testing Guide

## Summary of Changes

I've fixed and updated the Instagram follow/unfollow bot with the following improvements:

### 1. **Fixed Bot Implementation** (`api/pkg/instagram/bot/follower_bot.go`)
- ✅ Fixed typo in environment variable name (`MAX_PAUSE_BETWEEN_FOLLOWs` → `MAX_PAUSE_BETWEEN_FOLLOWS`)
- ✅ Improved configuration with proper defaults
- ✅ Added proper time window checking with Paris timezone
- ✅ Enhanced randomization for human-like behavior
- ✅ Better error handling and logging
- ✅ Fixed session management and login flow
- ✅ Improved 2FA handling

### 2. **Updated Cron Implementation** (`api/internal/crons/instagram/follower.go`)
- ✅ Uses proper BotConfig structure
- ✅ Checks time window before running
- ✅ Properly handles blacklist (only unfollows users currently followed)
- ✅ Better logging with emojis and timestamps
- ✅ Randomized sleep between runs (30-60 minutes)

### 3. **Fixed Test Command** (`api/cmd/test-instagram-follow/main.go`)
- ✅ Fixed JSON structure mismatch
- ✅ Better logging and progress tracking
- ✅ Proper filtering of duplicates and blacklisted users
- ✅ Clear test output

### 4. **Configuration Files**
- ✅ Updated `.env.example` with all bot configuration options
- ✅ Updated `docker-compose.yml` with all environment variables
- ✅ Ensured correct JSON structure for `fftt-instagram-accounts.json` and `instagram_blacklist.json`

### 5. **Documentation**
- ✅ Created comprehensive `INSTAGRAM_BOT_README.md`
- ✅ Created test script `test-bot.sh`
- ✅ Added troubleshooting guide

## How to Test

### Step 1: Update Your .env File
Make sure your `.env` file has the correct credentials:

```bash
INSTAGRAM_BOT_ENABLED=true
INSTAGRAM_BOT_USERNAME=your_username
INSTAGRAM_BOT_PASSWORD=your_password
INSTAGRAM_BOT_TOTP_SECRET=YOUR_TOTP_SECRET  # If 2FA is enabled
INSTAGRAM_BOT_HEADLESS=true  # Set to false to see the browser during testing
```

### Step 2: Restart Containers
```bash
docker-compose down
docker-compose up -d
```

### Step 3: Run the Test
```bash
cd api
chmod +x test-bot.sh
./test-bot.sh
```

Alternatively, run manually inside the container:
```bash
docker exec -it tournois-tt-api-1 /bin/sh -c "cd /go/src/tournois-tt/api && go run cmd/test-instagram-follow/main.go"
```

## What the Bot Does

### Follow Process:
1. ✅ Only runs during configured hours (default: 11:00-21:00 Paris time)
2. ✅ Gets followers from source accounts (`ffttofficiel`, `fftt_idf`)
3. ✅ Filters out users already followed
4. ✅ Filters out users in blacklist
5. ✅ Follows users in **randomized batches**:
   - Random batch size: 3-7 users
   - Random pause between follows: 5-15 seconds
   - Random pause between batches: 1-3 minutes
6. ✅ Respects daily limit (default: 30 follows/day)

### Unfollow Process:
1. ✅ Reads blacklist from `instagram_blacklist.json`
2. ✅ Only unfollows users that are:
   - In the blacklist
   - Currently followed by you
3. ✅ Uses same randomization as follow process
4. ✅ Respects daily limit (default: 30 unfollows/day)

## Randomization Features

The bot uses randomization to appear more human-like:

- ✅ **Random batch sizes**: Each batch has 3-7 users (configurable)
- ✅ **Random pauses between actions**: 5-15 seconds (configurable)
- ✅ **Random pauses between batches**: 1-3 minutes (configurable)
- ✅ **Shuffled account order**: Accounts are processed in random order
- ✅ **Random run intervals**: Bot runs every 30-60 minutes

## Configuration Options

All settings can be customized via environment variables:

### Time Window
- `INSTAGRAM_BOT_MIN_HOUR=11` (start hour, 24h format)
- `INSTAGRAM_BOT_MAX_HOUR=21` (end hour, 24h format)

### Daily Limits
- `INSTAGRAM_BOT_MAX_FOLLOWS_DAILY=30`
- `INSTAGRAM_BOT_MAX_UNFOLLOWS_DAILY=30`

### Randomization Ranges
- `INSTAGRAM_BOT_MIN_PAUSE_BETWEEN_FOLLOWS=5` (seconds)
- `INSTAGRAM_BOT_MAX_PAUSE_BETWEEN_FOLLOWS=15` (seconds)
- `INSTAGRAM_BOT_MIN_PAUSE_BETWEEN_BATCHES=1` (minutes)
- `INSTAGRAM_BOT_MAX_PAUSE_BETWEEN_BATCHES=3` (minutes)
- `INSTAGRAM_BOT_MIN_FOLLOW_BATCH_SIZE=3`
- `INSTAGRAM_BOT_MAX_FOLLOW_BATCH_SIZE=7`
- `INSTAGRAM_BOT_MIN_UNFOLLOW_BATCH_SIZE=3`
- `INSTAGRAM_BOT_MAX_UNFOLLOW_BATCH_SIZE=7`

## Monitoring

### Check Bot State
```bash
docker exec tournois-tt-api-1 cat /go/src/tournois-tt/api/tmp/bot_data/bot_state.json
```

This shows:
- Last run date
- Follows today
- Unfollows today

### Check Logs
```bash
docker logs -f tournois-tt-api-1
```

Look for:
- `🤖 Starting new follower bot session...`
- `✅ Successfully followed/unfollowed...`
- `⏰ Follower bot will run again in...`

## Troubleshooting

### Bot not running?
Check:
1. ✅ `INSTAGRAM_BOT_ENABLED=true` in `.env`
2. ✅ Container is running: `docker ps | grep api`
3. ✅ Time is within window: Check logs for current Paris time
4. ✅ Daily limit not reached: Check `bot_state.json`

### Login errors?
Check:
1. ✅ Username and password are correct
2. ✅ 2FA secret is correct (if using 2FA)
3. ✅ Screenshot at `tmp/bot_data/login_error.png`

### No users to follow?
- ✅ Check that source accounts have public followers
- ✅ Verify you're not already following everyone
- ✅ Check blacklist isn't too large

## File Structure

```
api/
├── fftt-instagram-accounts.json      # Source accounts to get followers from
├── instagram_blacklist.json          # Users to unfollow
├── tmp/bot_data/
│   ├── session.json                  # Browser session (auto-created)
│   ├── bot_state.json               # Daily counters (auto-created)
│   └── login_error.png              # Error screenshot (if login fails)
├── pkg/instagram/bot/
│   └── follower_bot.go              # Main bot implementation
├── internal/crons/instagram/
│   └── follower.go                  # Cron job that runs the bot
└── cmd/test-instagram-follow/
    └── main.go                       # Test command
```

## Next Steps

1. ✅ Make sure `.env` has correct credentials
2. ✅ Restart containers: `docker-compose down && docker-compose up -d`
3. ✅ Run test: `cd api && ./test-bot.sh`
4. ✅ Monitor logs: `docker logs -f tournois-tt-api-1`
5. ✅ Adjust configuration as needed

The bot will now:
- ✅ Run automatically every 30-60 minutes
- ✅ Only during configured hours (11:00-21:00 Paris time by default)
- ✅ Follow users from source accounts in randomized batches
- ✅ Unfollow blacklisted users
- ✅ Respect daily limits
- ✅ Use human-like randomization

## Need Help?

Check:
- `INSTAGRAM_BOT_README.md` for detailed documentation
- Container logs: `docker logs tournois-tt-api-1`
- Bot state: `tmp/bot_data/bot_state.json`
- Error screenshots: `tmp/bot_data/login_error.png`
