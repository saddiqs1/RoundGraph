package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	demoFileName := os.Getenv("DEMO1")
	f, err := os.Open(demoFileName)
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}
	defer f.Close()

	p := dem.NewParser(f)
	defer p.Close()

	// Find all round start events
	roundStarts := []RoundStart{}
	p.RegisterEventHandler(func(e events.RoundStart) {
		rs := RoundStart{
			tick: p.GameState().IngameTick(),
		}

		rs.t = p.GameState().TeamTerrorists().Score()
		rs.ct = p.GameState().TeamCounterTerrorists().Score()

		roundStarts = append(roundStarts, rs)
	})

	scoreUpdates := []ScoreUpdate{}
	p.RegisterEventHandler(func(e events.ScoreUpdated) {
		s := ScoreUpdate{
			tick:  p.GameState().IngameTick(),
			score: e.NewScore,
		}

		if e.TeamState.Team() == 2 {
			s.team = "t"
		} else if e.TeamState.Team() == 3 {
			s.team = "ct"
		}

		scoreUpdates = append(scoreUpdates, s)
	})

	gameHalftimes := []int{}
	p.RegisterEventHandler(func(e events.GameHalfEnded) {
		gameHalftimes = append(gameHalftimes, p.GameState().IngameTick())
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		log.Panic("failed to parse demo: ", err)
	}

	// Create round graph
	rg := newRoundGraph()

	// Set rg.rounds
	rg.setRounds(roundStarts, scoreUpdates, gameHalftimes, p.Header().PlaybackTicks)

	// Set edges for each node
	rg.setEdges()

	// Find the shortest path from the starting
	matchRounds := rg.getMatchRounds()

	// This is the set of rounds in a game...
	for _, r := range matchRounds {
		fmt.Printf("demo_goto %v - demo_goto %v, r = %v, t = %v, ct = %v \n", r.startTick, r.endTick, r.roundNumber, r.t, r.ct)
	}
}
