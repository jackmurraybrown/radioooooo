# todo

## later
- [ ] rate limiting on authenticated routes (defense-in-depth, low priority while auth gate is sufficient)

## notes

### contact_email: copy-on-expansion strategy
when the expansion job generates episodes from a show, it copies the show's `contact_email` directly onto each episode. this means:
- changing a show's contact_email after expansion does NOT update past episodes — each episode owns its email independently
- this is intentional: if a guest DJ does a one-off slot, the admin can override `contact_email` on that specific episode without touching the show
- the tradeoff is that bulk email changes require updating episodes manually or re-expanding (but re-expansion skips already-existing episodes, so that doesn't help)
- alternative considered: always look up via show at email-send time. rejected because episodes can outlive shows (show deleted, episodes remain) and we want the reminder/tracklist email to still work

### metadata enrichment: where does MusicBrainz/Discogs fit?
recommendation: radiooo API, not the wails app.

reasons:
- the wails app is local-first and offline during a live set — network calls to external APIs mid-show is unreliable
- enrichment is a good fit for a river background job: triggered after `SetTracks` saves a tracklist, runs async, updates tracks with MBIDs/Discogs IDs/links without blocking the DJ
- the enriched data (labels, catalog numbers, links) lives in the radiooo DB alongside the tracklist, not in the wails SQLite which is ephemeral (pruned after 48h)
- all input paths (form + wails webhook) converge on `SetTracks`, so one enrichment job covers both

the wails app could do local enrichment for display during the show (e.g. show artwork), but the canonical enriched record should live in radiooo.

when to build: after the basic tracklist flow is working end-to-end. the `episode_tracks` table already has `album` column; can add `musicbrainz_id`, `discogs_id`, `label`, `catalog_number` columns when ready.
