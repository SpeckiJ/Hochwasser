<h1 align="center" >🌊🌊🌊 Hochwasser 🌊🤽🌊</h1>
<p align="center"><img src="benchmarks/hochwasser_shuffle_vs_ordered.gif"/></p>
<p align="center"><img src="benchmarks/nmzs.gif"/></p>

Highly efficient, distributed [Pixelflut] client.

> Hochwasser brings back the */fun/* in social DDoSing!
>
> No more micro-ddosing: Get Hochwasser now and experience highs never seen before!111!

- Can send static images, text, generated patterns (animations upcoming)
- REPL enables fast iterations
- CnC server + client architecture (it's webscale!) (can also run in a single process)
- Faster than [sturmflut] (in some benchmarks at least)
- No dependencies (pixelflut apparently was considered a primary use case in the design of golang's stdlib 👍)

[pixelflut]: https://cccgoe.de/wiki/Pixelflut
[sturmflut]: https://github.com/TobleMiner/sturmflut

## build / install
1. have a `go` installation >= 1.12
2. `go get github.com/SpeckiJ/Hochwasser`
3. `go install github.com/SpeckiJ/Hochwasser`

## benchmark
The following benchmark was run on a max-spec X280 against version [d4c574b].

I could not figure out what the performance bottleneck is, but it doesn't seem
to be CPU limited, as turbo-boost doesn't kick in.

To reproduce, run the following commands in separate shells:

```sh
iperf -s -p 1234
go run main.go -image benchmark/test.png -connections 10
```

![screenshot: 55 Gbps of hochwasser](benchmarks/benchmark_x280.png)

55 Gbps on average! 🌊🌊🌊

[sturmflut] (`./sturmflut 127.0.0.1:1337 benchmark/test.png -t 10`, version `8ec6ee9`) managed to get 48 Gpbs throughput on this system.

> Hint: Benchmarking throughput against the [pixelnuke][pixelflut_gh] server is
  pointless, as performance is then CPU-limited to ~1 Gbps by the server.
  Using [iperf] removes the server limitation.
  This also means that these metrics of several Gbps are far higher than
  realworld scenarios.

[d4c574b]: https://github.com/SpeckiJ/Hochwasser/commit/d4c574be103a7bad69349f29402694f51058184c
[pixelflut_gh]: https://github.com/defnull/pixelflut
[iperf]: https://iperf.fr/

## future ideas
For future ideas check [IDEAS](https://github.com/SpeckiJ/Hochwasser/blob/master/IDEAS.md)

<p align="center"><img src="benchmarks/hochwasser_vs_sturmflut.gif"/></p>
