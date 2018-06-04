# Statistics library for Go

Documentation on [https://godoc.org/github.com/Brickchain/go-stats.v1](https://godoc.org/github.com/Brickchain/go-stats.v1).

Based on [go-metrics](https://github.com/armon/go-metrics).

Simplifies setup of metrics sinks and collecting statistics.

## Usage
```go
package main

import (
    "os"
    "time"
    "path"

    stats "github.com/Brickchain/go-stats.v1"
)

func main() {
    // start an inmem metrics sink that will print metrics once per minute.
    // set the instance name to the name of our binary.
    stats.Setup("inmem", path.Base(os.Args[0]))

    someFunc()

    // Wait a bit more than a minute in order to see the metrics being printed
    time.Sleep(time.Second * 62)
}

func someFunc() {
    t := stats.StartTimer("someFunc.total")
    defer t.Stop()
}
```