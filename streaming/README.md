# streaming

two containers: **liquidsoap** encodes and mixes audio, **icecast** distributes it to listeners.

```
go broadcast controller
        ↓ unix socket (/var/run/liquidsoap/radio.sock)
    liquidsoap
        ↓ icecast source protocol
      icecast
        ↓ http
     listeners
```

## how it works

**liquidsoap** runs `radio.liq` which sets up three sources:

1. **live harbor** (port 8005) — DJ software (butt, mixxx) connects here for live shows
2. **request queue** (`radio_queue`) — go broadcast controller pushes file paths here
3. **blank** — silence fallback when nothing is scheduled

the go broadcast controller polls the schedule every second (clock-aligned). when an episode starts it pushes the file path to `radio_queue`. when an episode ends it sends `radio_queue.skip`. all source transitions use a 1s crossfade.

**icecast** receives the encoded mp3 stream from liquidsoap on the `/main` mount and serves it to listeners on port 8000.

## control interface

liquidsoap exposes a unix socket at `socket/radio.sock` (bind-mounted into both containers). the go client connects here to send commands.

```bash
# push a track manually
echo "radio_queue.push /media/fif-a-home.mp3" | socat - UNIX-CONNECT:socket/radio.sock

# skip current track
echo "radio_queue.skip" | socat - UNIX-CONNECT:socket/radio.sock
```
