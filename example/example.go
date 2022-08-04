package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"

	rd "github.com/saddiqs1/round-detection/v1"
)

// Run like this: go run example.go -demo /path/to/demo.dem
func main() {
	f, err := os.Open(DemoPathFromArgs())
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}
	defer f.Close()

	p := dem.NewParser(f)
	defer p.Close()

	// Find all round start events
	roundStarts := []rd.RoundStart{}
	p.RegisterEventHandler(func(e events.RoundStart) {
		rs := rd.RoundStart{
			Tick: p.GameState().IngameTick(),
		}

		rs.T = p.GameState().TeamTerrorists().Score()
		rs.CT = p.GameState().TeamCounterTerrorists().Score()

		roundStarts = append(roundStarts, rs)
	})

	scoreUpdates := []rd.ScoreUpdate{}
	p.RegisterEventHandler(func(e events.ScoreUpdated) {
		s := rd.ScoreUpdate{
			Tick:  p.GameState().IngameTick(),
			Score: e.NewScore,
		}

		if e.TeamState.Team() == 2 {
			s.Team = "t"
		} else if e.TeamState.Team() == 3 {
			s.Team = "ct"
		}

		scoreUpdates = append(scoreUpdates, s)
	})

	gameHalftimes := []int{}
	p.RegisterEventHandler(func(e events.GameHalfEnded) {
		gameHalftimes = append(gameHalftimes, p.GameState().IngameTick())
	})

	// Parse as much of demo as possible
	lastTick := 0
	for ok, err := p.ParseNextFrame(); ok; ok, err = p.ParseNextFrame() {
		if err != nil {
			fmt.Printf("error while parsing demo: %v \n", err)
		}

		if lastTick < p.GameState().IngameTick() {
			lastTick = p.GameState().IngameTick()
		}
	}

	// Create round graph
	rg := rd.NewRoundGraph()

	// Set rg.rounds
	rg.SetRounds(roundStarts, scoreUpdates, gameHalftimes, lastTick)

	// Set edges for each node
	rg.SetEdges()

	// Find the shortest path from the starting
	matchRounds := rg.GetMatchRounds()

	// This is the set of rounds in a game...
	for _, r := range matchRounds {
		fmt.Printf("demo_goto %v - demo_goto %v, r = %v, t = %v, ct = %v \n", r.StartTick, r.EndTick, r.RoundNumber, r.T, r.CT)
	}
}

// DemoPathFromArgs returns the value of the -demo command line flag.
// Panics if an error occurs.
func DemoPathFromArgs() string {
	fl := new(flag.FlagSet)

	demPathPtr := fl.String("demo", "", "Demo file `path`")

	err := fl.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	demPath := *demPathPtr

	return demPath
}
