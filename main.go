package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	dem "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v3/pkg/demoinfocs/events"
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

	scores := []Score{}
	firstScore := Score{
		tick: 0,
		t:    0,
		ct:   0,
	}
	scores = append(scores, firstScore)

	p.RegisterEventHandler(func(e events.ScoreUpdated) {
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

		//ensure there are no duplicates
		scoreExists := false
		for _, score := range scores {
			if score == s {
				scoreExists = true
				break
			}
		}
		if !scoreExists {
			scores = append(scores, s)
		}
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		log.Panic("failed to parse demo: ", err)
	}

	// Create score graph
	sg := newScoreGraph()
	for _, s := range scores {
		u := sg.NewNode()
		uid := u.ID()
		u = node{
			score: s,
			id:    uid,
		}
		sg.AddNode(u)
		sg.ids[s] = uid
		sg.scores[uid] = s

		fmt.Printf("id = %v, tick = %v, t = %v, ct = %v \n", uid, s.tick, s.t, s.ct)
	}

	nodes := sg.Nodes()

	for nodes.Len() > 0 {
		nodes.Next()
		n := sg.Node(nodes.Node().ID()) // n = current node
		cs := sg.scoreAtId(n.ID())      // get current score (cs)

		// loop through scores to attach edges
		for _, s := range sg.scores {
			//TODO - cleanup how to pick the starting 0-0 point (some logic to say what is the last 0 total score with lower tick maybe)
			//TODO - will have to account for OT scores eventually
			if cs.tick < s.tick { // cs tick must be lower then the next score in the graph to attach edge to it i.e. can't go back in time
				if cs.Total() == s.Total()-1 { // if score equals currentTotal+1 then add edge to it
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

		fmt.Printf("node id = %v, total = %v \n", n.ID(), sg.scoreAtId(n.ID()).Total())
	}

	//TODO - find longest possible path of nodes in sg
	//TODO - we want the longest possible path where round total 0 has highest tick number possible
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
