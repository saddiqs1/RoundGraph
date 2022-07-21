# round-detection

A mini project to find the ticks that represent the played rounds within a CS:GO demo using the [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) parser. The RoundStart and RoundEnd events seem to not properly trigger on actual round starts/ends in certain demos for whatever reason, so this repo intends to parse through a demo and return the ticks of rounds that were actually played.

## How to run

1. Install [Go](https://go.dev/doc/install) (at least version 1.18)
2. Create a `demos` folder in the root folder to store any demos you want to parse
3. Create a `.env` file in the root folder, following the format shown in `.env example`
4. Run `go run main.go`