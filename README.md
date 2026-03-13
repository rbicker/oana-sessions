# oana-sessions

## Build & Deploy
```bash
docker build -t registry.asdf.ooo/oana . && docker push registry.asdf.ooo/oana
```

## Oana sync

On startup, the app now:

1. Runs app DB migrations (including creation of the `categories` collection).
2. Runs an initial category sync from:
   `https://pointbreak.coremanager.info/shop/index/index/categorie/4`
3. Runs an initial sessions sync into the `sessions` collection by calling:
   `https://pointbreak.coremanager.info/shop/index/geteventbox`
   for every active category and every day in the range `today ... today + 40 days`.
4. Registers cron sync jobs (server local time):
   - categories: daily at `02:00`
   - sessions: hourly at minute `00`
   - session participants:
     - every `5` minutes for sessions in `today + tomorrow`
     - hourly for sessions in `day +2 ... day +40`

The sync upserts categories by a stable `external_key` and soft-deletes stale
records by setting `is_deleted=true` (instead of removing them).

Each collection record has:
- `created` and `updated` (PocketBase system fields)
- `last_synced_at` (set by the sync jobs)

The `sessions` collection also stores:
- `participants_count`
- `participants_synced_at`
- `event_date` (`YYYY-MM-DD`)
- `event_weekday` (`1=Mon ... 7=Sun`)
- `event_start_time` (`HH:mm`)
- `event_start_hour` (`0..23`)
- `event_start_minute` (`0..1439`)
- `event_time_known` (`true` when start time could be parsed)

These participant fields are refreshed by the participant sync jobs.

The `session_participants` collection stores compact snapshots:
- `session` (relation to `sessions`)
- `recorded_at`
- `participants_count`
- `event_start_at`
- `minutes_before_start`

The participant sync inserts a new snapshot only when the count changed.

## Frontend (SvelteKit + DaisyUI)

The project includes a frontend in `frontend/`.
It reads `sessions` directly from PocketBase (`/api/collections/sessions/records`),
keeps the original per-day table colors, and enables realtime participant updates
for today's sessions.

Run locally:

1. `cd frontend`
2. `npm install`
3. `npm run dev`

Build static assets for PocketBase:

1. `cd frontend`
2. `npm run build`

This writes the production frontend to `pb_public/`, which is served by the app.
