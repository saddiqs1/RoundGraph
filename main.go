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

	scores := []Score{}
	firstScore := Score{
		tick: 0,
		t:    0,
		ct:   0,
	}
	scores = append(scores, firstScore)

	p.RegisterEventHandler(func(e events.ScoreUpdated) {

		/*
			graph collection the right way to go?
			find 0-0 start -- maybe find 1-0 and backtrack to most recent 0-0?
			keep track of score's incrementing by 1 until they total 15
			score's will flip sides
			keep track of score's until one side hit's 16
			or if both equal 15, then OT logic...
		*/

		s := Score{
			tick: p.CurrentFrame(),
		}

		// 2 = t, 3 = ct
		if e.TeamState.Team() == 2 {
			s.t = e.NewScore
			s.ct = scores[len(scores)-1].ct
		} else if e.TeamState.Team() == 3 {
			s.t = scores[len(scores)-1].t
			s.ct = e.NewScore
		}

		scores = append(scores, s)

		fmt.Printf("tick = %v, t = %v, ct = %v \n", s.tick, s.t, s.ct)
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		log.Panic("failed to parse demo: ", err)
	}
}

type Score struct {
	tick int
	t    int
	ct   int
}

func (s Score) Total() int { return s.t + s.ct }
