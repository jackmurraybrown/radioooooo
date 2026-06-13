#!/usr/bin/liquidsoap
# ✮⋆‧°—°‧⋆✮ radiooo — liquidsoap entrypoint
# step 1: fallback silence only, no schedule integration yet

# ⋆˙⟡ telnet server so Go can send commands later
settings.server.telnet.set(true)
settings.server.telnet.bind_addr.set("0.0.0.0")
settings.server.telnet.port.set(1234)

# . ݁₊ ✶ play files from /media, fall back to silence if empty
# ⋆˙⟡ single loops one file infallibly — prod will use request.queue via Go controller
radio = single("/media/fif-a-home.mp3")

# ✮ ⋆ ˚｡𖦹 push to icecast on the 'icecast' container hostname
output.icecast(
  %mp3(bitrate = 128),
  host     = "icecast",
  port     = 8000,
  password = "hackme",
  mount    = "/main",
  name     = "radiooo",
  description = "radiooo fm",
  radio
)
