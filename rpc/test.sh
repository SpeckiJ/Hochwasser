#!/bin/bash

# runs two instances of hochwasser (rpc server & client each) against a fake (iperf)
# pixelflut server for a few seconds, and opens pprof afterwar

runtime=${1:-"2s"}

function cleanup {
  kill ${pids[@]} > /dev/null 2>&1
}
trap cleanup EXIT

wd=$(dirname "$0")
pids=()

iperf -p 1337 -s > /dev/null 2>&1 &
pids+=($!)

go run . -image $wd/../benchmarks/black_small.png -rán :1234 &
pids+=($!)
sleep 0.2

go run . -image $wd/../benchmarks/white_small.png -hevring :1234 -runtime "$runtime" -cpuprofile hevring.prof
pids+=($!)

cleanup

go tool pprof -http :8080 Hochwasser hevring.prof
#go tool pprof -http :8081 Hochwasser ran.prof
