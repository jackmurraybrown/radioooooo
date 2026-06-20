#!/usr/bin/liquidsoap
# ✮⋆‧°—°‧⋆✮ radiooo — liquidsoap entrypoint

# ⋆˙⟡ unix socket — Go controller dials /var/run/liquidsoap/radio.sock
settings.server.socket.set(true)
settings.server.socket.path.set("/var/run/liquidsoap/radio.sock")

# ✮ ⋆ ˚｡𖦹 named queue — Go controller pushes tracks via: radio_queue.push <path>
radio_queue = request.queue(id="radio_queue")

# . ݁₊ ✶ 1s crossfade at every track boundary — covers natural endings and Go-triggered skips
radio_source = crossfade(duration=1., radio_queue)

# ⋆.˚ ⊹₊⟡ live dj input — connect to port 8005, mount /live
# djs connect here, not to icecast directly — keeps them in the fallback chain
# when a dj drops, liquidsoap falls back to radio_source automatically
live = input.harbor("live", port=8005, password="hackme")

# ٩(^ᗜ^ ) 1s fade on all source switches
def fade(a, b) = sequence([fade.out(duration=1., a), b]) end

# ✮⋆‧° fallback playlist — loops when nothing is scheduled
fallback_media = playlist("/media/fallback", mode="randomize", reload_mode="watch")

# ⊹ ࣪ ˖ priority: live → automated queue → fallback playlist → silence
radio = fallback(
  track_sensitive=false,
  transitions=[fade, fade, fade, fade],
  [live, radio_source, fallback_media, blank()]
)

# push to icecast on the 'icecast' container hostname
output.icecast(
  %mp3(bitrate = 256),
  host     = "icecast",
  port     = 8000,
  password = "hackme",
  mount    = "/main",
  name     = "radiooo",
  description = "radiooo fm",
  radio
)
