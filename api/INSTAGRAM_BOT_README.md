# Instagram Bot Configuration

## Environment Variables

### Required Variables
- `INSTAGRAM_BOT_USERNAME` - Instagram account username
- `INSTAGRAM_BOT_PASSWORD` - Instagram account password
- `INSTAGRAM_BOT_TOTP_SECRET` - TOTP secret for 2FA (if enabled)

### Optional Configuration Variables

#### Time Window
- `INSTAGRAM_BOT_MIN_HOUR` - Minimum hour to run (0-23, default: 11)
- `INSTAGRAM_BOT_MAX_HOUR` - Maximum hour to run (0-24, default: 21)

#### Daily Limits
- `INSTAGRAM_BOT_MAX_FOLLOWS_DAILY` - Maximum follows per day (default: 30)
- `INSTAGRAM_BOT_MAX_UNFOLLOWS_DAILY` - Maximum unfollows per day (default: 30)

#### Pause Settings (Randomization)
- `INSTAGRAM_BOT_MIN_PAUSE_BETWEEN_FOLLOWS` - Min seconds between actions (default: 5)
- `INSTAGRAM_BOT_MAX_PAUSE_BETWEEN_FOLLOWS` - Max seconds between actions (default: 15)
- `INSTAGRAM_BOT_MIN_PAUSE_BETWEEN_BATCHES` - Min minutes between batches (default: 1)
- `INSTAGRAM_BOT_MAX_PAUSE_BETWEEN_BATCHES` - Max minutes between batches (default: 3)

#### Batch Sizes (Randomization)
- `INSTAGRAM_BOT_MIN_FOLLOW_BATCH_SIZE` - Min users per follow batch (default: 3)
- `INSTAGRAM_BOT_MAX_FOLLOW_BATCH_SIZE` - Max users per follow batch (default: 7)
- `INSTAGRAM_BOT_MIN_UNFOLLOW_BATCH_SIZE` - Min users per unfollow batch (default: 3)
- `INSTAGRAM_BOT_MAX_UNFOLLOW_BATCH_SIZE` - Max users per unfollow batch (default: 7)

#### Other
- `INSTAGRAM_BOT_HEADLESS` - Run browser in headless mode (default: true for production)

## Files

### Source Accounts (fftt-instagram-accounts.json)
Contains accounts whose followers will be targeted for following:
```json
{
  "source_accounts": [
    "ffttofficiel",
    "fftt_idf"
  ]
}
```

### Blacklist (instagram_blacklist.json)
Contains usernames to unfollow if currently followed:
```json
{
  "usernames": [
    "username1",
    "username2"
  ]
}
```

## Testing

### Local Testing (with visible browser)
```bash
cd api
chmod +x test-bot.sh
./test-bot.sh
```

### Manual Testing Inside Container
```bash
docker exec -it tournois-tt-api-1 /bin/sh
cd /go/src/tournois-tt/api
export INSTAGRAM_BOT_HEADLESS=true
go run cmd/test-instagram-follow/main.go
```

## How It Works

### Follow Process
1. Bot runs every 30-60 minutes (randomized)
2. Checks if current time is within configured time window (default: 11:00-21:00 Paris time)
3. Gets your current following list
4. For each source account:
   - Gets their followers
   - Filters out users you already follow
   - Filters out users in blacklist
5. Follows new users in randomized batches:
   - Random batch size between MIN and MAX batch size
   - Random pause between each follow action
   - Random pause between batches
6. Stops when daily limit is reached

### Unfollow Process
1. Loads blacklist from instagram_blacklist.json
2. Checks which blacklisted users you currently follow
3. Unfollows them in randomized batches with the same random pause logic

### Daily Limits
- The bot tracks follows/unfollows per day
- Counters reset at midnight
- State is persisted to `tmp/bot_data/bot_state.json`

## Best Practices

1. **Start Conservative**: Begin with low daily limits (20-30) and small batches
2. **Realistic Timing**: Keep time window to daytime hours (e.g., 10:00-20:00)
3. **Randomization**: The bot uses randomization to appear more human-like
4. **Session Persistence**: Browser session is saved to avoid repeated logins
5. **2FA**: If you have 2FA enabled, provide TOTP_SECRET for automatic code generation

## Troubleshooting

### Check Bot State
```bash
docker exec tournois-tt-api-1 cat /go/src/tournois-tt/api/tmp/bot_data/bot_state.json
```

### Check Logs
```bash
docker logs -f tournois-tt-api-1
```

### View Screenshots (on login errors)
Login errors create screenshots at: `tmp/bot_data/login_error.png`

### Common Issues

1. **Login Failed**: Check credentials and 2FA setup
2. **Outside Time Window**: Check `INSTAGRAM_BOT_MIN_HOUR` and `INSTAGRAM_BOT_MAX_HOUR`
3. **Daily Limit Reached**: Check bot_state.json and wait for next day
4. **Private Profiles**: Bot cannot get followers from private accounts
