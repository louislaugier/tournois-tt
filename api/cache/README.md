# Cache Directory

This directory contains persistent cache files for the tournament system.

## Files

### `data.json`
- Contains all tournament data scraped from FFTT
- Updated regularly by the scraper cron job
- Used by the API and frontend

### `posted_instagram.json` ⭐
- **IMPORTANT**: This file is tracked in git
- Records all tournaments posted to Instagram/Threads
- Prevents duplicate posts
- Synchronized across all instances/rebuilds
- Updated automatically after each successful post
- **Auto-synced with Instagram API** every hour (detects deleted posts)
- **Synced on startup** to validate cache against live API

## Why `posted_instagram.json` is in Git

Unlike temporary data, this file tracks the **posting history** which must be:
- ✅ Shared across all server instances
- ✅ Preserved during container rebuilds
- ✅ Synchronized between deployments
- ✅ Never lost or reset

## Format of `posted_instagram.json`

```json
[
  {
    "tournament_id": 3340,
    "tournament_name": "Tournoi National...",
    "posted_at": "2025-11-04T18:30:00Z",
    "instagram_feed": true,
    "instagram_story": true,
    "threads": true,
    "instagram_post_id": "18123456789",
    "instagram_story_id": "18987654321",
    "threads_post_id": "17555555555"
  }
]
```

## How It Works

1. **On startup**: 
   - Cache is loaded from disk
   - Validates against Instagram/Threads APIs
   - Removes entries for deleted posts

2. **Before posting**: 
   - Cache is checked first (instant)
   - API fallback: If not in cache, check Instagram/Threads APIs (50 recent posts)
   - If found in API but not cache → Updates cache

3. **After posting**: 
   - Cache is updated and saved to disk
   - Entry includes post IDs for verification

4. **Every hour (cron)**:
   - Validates each cached post against Instagram API
   - Removes posts that have been deleted
   - Keeps cache in sync with reality

5. **Manual sync**:
   - `make cache-sync` to force immediate validation
   - Useful after manually deleting posts

6. **In git**: 
   - Changes are committed so all instances stay in sync

