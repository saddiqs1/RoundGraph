package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
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

	// Set edges for each node
	nodes := sg.Nodes()
	for nodes.Len() > 0 {
		nodes.Next()
		n := sg.Node(nodes.Node().ID()) // n = current node
		cs := sg.scoreAtId(n.ID())      // get current score (cs)

		// loop through scores to attach edges
		for _, s := range sg.scores {
			/*
				TODO:
				Overtime scores
			*/
			if cs.tick < s.tick { // cs tick must be lower then the next score in the graph to attach edge to it i.e. can't go back in time
				if needsSetWeight(cs, s) {
					sg.SetWeightedEdge(simple.WeightedEdge{F: n, T: sg.nodeAtScore(s), W: float64(s.tick - cs.tick)})
				} else if isFourteen(cs, s) {
					sg.SetWeightedEdge(simple.WeightedEdge{F: n, T: sg.nodeAtScore(s), W: 1})
				} else if hasIncreasedByOne(cs, s) {
					sg.SetWeightedEdge(simple.WeightedEdge{F: n, T: sg.nodeAtScore(s), W: 1})
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

	// Find the shortest path from all starting scores
	finalScores := []graph.Node{}
	var finalRoundsWeight float64
	firstPass := true
	pt := path.DijkstraAllPaths(sg)
	for _, startScore := range startScores {
		path, weight, _ := pt.Between(sg.idAtScore(startScore), sg.idAtScore(finalScore))
		if firstPass {
			finalScores = path
			finalRoundsWeight = weight
			firstPass = false
		} else {
			if weight < finalRoundsWeight {
				finalScores = path
				finalRoundsWeight = weight
			}
		}
	}

	/*
		TODO
		- map these to Round struct, which includes start and end tick for the round
	*/
	// This is the set of rounds in a game...
	for _, s := range finalScores {
		r := sg.scoreToRound(sg.scoreAtId(s.ID()), p.Header().PlaybackTicks)
		fmt.Printf("demo_goto %v - demo_goto %v, r = %v, t = %v, ct = %v \n", r.startTick, r.endTick, r.roundNumber, r.t, r.ct)
	}
}

func hasIncreasedByOne(s1, s2 Score) bool {
	if s1.Total() == s2.Total()-1 { // Ensure score increased by 1
		if s1.t == s2.t || s2.t == s1.t+1 { // Ensure t score only jumped up by 1
			if s1.ct == s2.ct || s2.ct == s1.ct+1 { // Ensure ct score only jumped up by 1
				return true
			}
		}
	}

	return false
}

func isFourteen(s1, s2 Score) bool {
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

func needsSetWeight(s1, s2 Score) bool {
	s1Total := s1.Total()
	if s1Total == 0 || s1Total == 15 {
		if hasIncreasedByOne(s1, s2) {
			return true
		}
	}

	return false
}
