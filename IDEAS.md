# feature ideas
- pluggable cli: commands for image, text, shader rendering
- support animations / frame concept
- visualization client
- CnC: server distributes jobs to connected clients
- webassembly port?

# performance considerations
- server limitations: rendering is bottleneck. maybe artificial limitations (commands per draw, connections per IP, queue)
  - when network isn't bottleneck: fetch each pixel & only send updates for wrong color (?)
  - sync sending with draw frequency (?)
  - use virtual subnets for more IPs (ipv6?) (?)
- client limitations: PCI bus is bottleneck? depends on HW I guess
  - precompute everything
  - distribute across cores for max PCI bus saturation (?)
- network limitations: packet size, ACKs, congestion
  - treat benchmarks on `loopback` with care, it has no packet size limitation. real world interfaces will enforce a max size of 1514 bytes [1]
  - avoid packet split if >1514B (?)
  - use `TCP_NODELAY` (?)
  - https://stackoverflow.com/questions/5832308/linux-loopback-performance-with-tcp-nodelay-enabled
- cognitive limitations: draw order
  - randomized pixel order should give a better idea of the image with equal dominance (?)

# concept: CLI for distributed hochwasser v2

> pixelflut endlich *durchgespielt*

```
hochwasser     --server
    provide    [type] [input] --effect --offset --scale --port --nosend
    subscribe  --connections --shuffle --diffmode
    view       --fullscreen
```

- CLI via `github.com/spf13/cobra`
  - `--server` refers to pixelflut server or hochwasser jobprovider, depending on mode

- jobprovider has different input types (`image`, `text`, `shader`?), each is parsed into an `image.GIF`
  - jobprovider also sends image itself?

- when subscriber connects to jobprovider, `GIF` is split up, and (re)distributed to all subscribers
  - protocol: (address,offset,imgdata) serialized with `gob`?

- viewer fetches into framebuffer, renders via opengl?
