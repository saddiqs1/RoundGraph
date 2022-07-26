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
	firstScore := Score{
		tick: 0,
		t:    0,
		ct:   0,
	}
	sg.addScore(firstScore)
	previousScore := firstScore

	p.RegisterEventHandler(func(e events.ScoreUpdated) {
		s := Score{
			tick: p.CurrentFrame(),
		}

		// 2 = t, 3 = ct
		if e.TeamState.Team() == 2 {
			s.t = e.NewScore
			s.ct = previousScore.ct
		} else if e.TeamState.Team() == 3 {
			s.t = previousScore.t
			s.ct = e.NewScore
		}

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
			previousScore = s
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
			//TODO - will have to account for OT scores eventually
			if cs.tick < s.tick { // cs tick must be lower then the next score in the graph to attach edge to it i.e. can't go back in time
				if cs.Total() == s.Total()-1 { // if score equals currentTotal+1 then add edge to it // TODO - maybe this check doesn't need to be here?
					if hasIncreasedByOne(cs, s) {
						sg.SetEdge(simple.Edge{F: n, T: sg.nodeAtScore(s)})
					}
				} else if cs.Total() == 15 { // if current total is 15, then the scores will switch
					if cs.t == s.ct && cs.ct == s.t {
						sg.SetEdge(simple.Edge{F: n, T: sg.nodeAtScore(s)})
					}
				}
			}
		}
	}

	//find all 0-0 scores
	startScores := []Score{}
	for _, s := range sg.scores {
		if s.Total() == 0 {
			startScores = append(startScores, s)
		}
	}

	//find highest score
	finalScore := Score{}
	for _, s := range sg.scores {
		if finalScore.Total() < s.Total() {
			finalScore = s
		}
	}

	//finding the longest possible path that exists, with the highest starting 0-0 score at the start
	finalRounds := []Score{}
	finalRounds = append(finalRounds, Score{})
	for _, startScore := range startScores {
		rounds := []Score{}
		df := traverse.DepthFirst{
			Visit: func(n graph.Node) {
				rounds = append(rounds, sg.scoreAtId(n.ID()))
			},
		}
		df.Walk(sg, sg.nodeAtScore(startScore), func(n graph.Node) bool {
			// until score is finalscore
			return sg.scoreAtId(n.ID()) == finalScore
		})

		if len(finalRounds) <= len(rounds) {
			if finalRounds[0].tick < rounds[0].tick {
				finalRounds = rounds
			}
		}
	}

	// This is the set of rounds in a game...
	for _, r := range finalRounds {
		fmt.Printf("demo_goto %v, t = %v, ct = %v \n", r.tick, r.t, r.ct)
	}
}

func hasIncreasedByOne(s1 Score, s2 Score) bool {
	result := false

	if s1.t == s2.t || s2.t == s1.t+1 { // Ensure t score only jumped up by 1
		if s1.ct == s2.ct || s2.ct == s1.ct+1 { // Ensure ct score only jumped up by 1
			result = true
		}
	}

	return result
}
