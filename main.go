package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/traverse"
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

	// Create score graph
	sg := newScoreGraph()

	p.RegisterEventHandler(func(e events.RoundStart) {
		s := Score{
			tick: p.GameState().IngameTick(),
		}

		s.t = p.GameState().TeamTerrorists().Score()
		s.ct = p.GameState().TeamCounterTerrorists().Score()

		//ensure there are no duplicates
		scoreExists := false
		for _, score := range sg.scores {
			if score == s {
				scoreExists = true
				break
			}
		}
		if !scoreExists {
			sg.addScore(s)
		}
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		log.Panic("failed to parse demo: ", err)
	}

	nodes := sg.Nodes()
	for nodes.Len() > 0 {
		nodes.Next()
		n := sg.Node(nodes.Node().ID()) // n = current node
		cs := sg.scoreAtId(n.ID())      // get current score (cs)

		// loop through scores to attach edges
		for _, s := range sg.scores {
			/*
				TODO - EDGE CASES
				Overtime
			*/
			if cs.tick < s.tick { // cs tick must be lower then the next score in the graph to attach edge to it i.e. can't go back in time
				if isHalfTime(cs, s) {
					sg.SetEdge(simple.Edge{F: n, T: sg.nodeAtScore(s)})
				} else if hasIncreasedByOne(cs, s) {
					sg.SetEdge(simple.Edge{F: n, T: sg.nodeAtScore(s)})
				}
			}
		}
	}

	startScores := []Score{}
	finalScore := Score{}

	for _, s := range sg.scores {
		//find all 0-0 scores
		if s.Total() == 0 {
			startScores = append(startScores, s)
		}

		//find highest score
		if finalScore.Total() < s.Total() {
			finalScore = s
		}
	}

	//finding the longest possible path that exists, with the highest starting 0-0 score at the start
	//TODO - WE NEED TO FIND LONGEST PATH
	finalRounds := []Score{}
	finalRounds = append(finalRounds, Score{})
	for _, startScore := range startScores {
		rounds := []Score{}
		df := traverse.DepthFirst{
			Visit: func(n graph.Node) {
				// TODO - check if it's halftime node, do something
				rounds = append(rounds, sg.scoreAtId(n.ID()))
			},
		}
		df.Walk(sg, sg.nodeAtScore(startScore), func(n graph.Node) bool {
			// until score is finalscore
			return sg.scoreAtId(n.ID()) == finalScore
		})

		if len(finalRounds) < len(rounds) {
			finalRounds = rounds
		} else if len(finalRounds) == len(rounds) && finalRounds[0].tick < rounds[0].tick {
			finalRounds = rounds
		}
	}

	/*
		TODO
		- maybe change to weighted graph, change weights of edges based off how close the ticks are. Actually find longest path
		- map these to Round struct, which includes start and end tick for the round
	*/
	// This is the set of rounds in a game...
	for _, r := range finalRounds {
		fmt.Printf("demo_goto %v, t = %v, ct = %v \n", r.tick, r.t, r.ct)
	}
}

func hasIncreasedByOne(s1 Score, s2 Score) bool {
	if s1.Total() == s2.Total()-1 { // Ensure score increased by 1
		if s1.t == s2.t || s2.t == s1.t+1 { // Ensure t score only jumped up by 1
			if s1.ct == s2.ct || s2.ct == s1.ct+1 { // Ensure ct score only jumped up by 1
				return true
			}
		}
	}

	return false
}

func isHalfTime(s1 Score, s2 Score) bool {
	if s1.Total() == 14 { // if current total is 14, then the scores will switch next round
		// t and ct will switch, one of them will +1 e.g. 4-10 becomes 11-4
		if s1.t == s2.ct && s1.ct == s2.t-1 {
			return true
		}

		if s1.ct == s2.t && s1.t == s2.ct-1 {
			return true
		}
	}

	return false
}
