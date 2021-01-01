# feature ideas
- more draw modes
  - patterns
    - https://github.com/MarcelMue/konstrukt ?
  - image scaling
  - snake
  - sheep.exe
  - offset rand: random size/color
- proper public api for the fast network handling
  - `Fluter` abstraction, implementing `Reader` to update commands? ringbuffer?
- support animations / frame concept
- support (stackable) effects
- make job distribution fully P2P using [2D CAN / Z-ordercurve adressing](https://git.nroo.de/norwin/geo-dht)

# performance considerations
- server limitations: rendering is bottleneck. maybe artificial limitations (commands per draw, connections per IP, queue)
  - when network isn't bottleneck: fetch each pixel & only send updates for wrong color (?)
  - sync sending with draw frequency (?)
  - use virtual subnets for more IPs (ipv6?) (?)
- client limitations: PCI bus is bottleneck? depends on HW I guess
  - distribute across cores for max PCI bus saturation (?)
  - use userland tcp stack (e.g. https://github.com/google/netstack or even https://github.com/luigirizzo/netmap)
- network limitations: packet size, ACKs, congestion
  - treat benchmarks on `loopback` with care, it has no packet size limitation. real world interfaces will enforce a max size of 1514 bytes [1]
  - avoid packet split if >1514B (?)
  - use `TCP_NODELAY` (?)
  - https://stackoverflow.com/questions/5832308/linux-loopback-performance-with-tcp-nodelay-enabled
- cognitive limitations: draw order
  - randomized pixel order should give a better idea of the image with equal dominance (?)
  - use an energy function like in seam carving to prioritize regions?
