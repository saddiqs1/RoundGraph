package main

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

type RoundGraph struct {
	ids    map[Round]int64
	rounds map[int64]Round
	*simple.WeightedDirectedGraph
}

func newRoundGraph() RoundGraph {
	return RoundGraph{
		ids:                   make(map[Round]int64),
		rounds:                make(map[int64]Round),
		WeightedDirectedGraph: simple.NewWeightedDirectedGraph(0, 0),
	}
}

func (g RoundGraph) addRound(r Round) {
	n := g.NewNode()
	id := n.ID()
	n = node{
		round: r,
		id:    id,
	}
	g.AddNode(n)
	g.ids[r] = id
	g.rounds[id] = r
}

func (g RoundGraph) nodeAtRound(r Round) graph.Node {
	id, ok := g.ids[r]
	if !ok {
		return nil
	}
	return g.WeightedDirectedGraph.Node(id)
}

func (g RoundGraph) roundAtId(id int64) Round {
	return g.rounds[id]
}

func (g RoundGraph) idAtRound(r Round) int64 {
	return g.ids[r]
}

type Round struct {
	startTick   int
	endTick     int
	roundNumber int
	t           int
	ct          int
}

type node struct {
	round Round
	id    int64
}

type RoundStart struct {
	tick int
	t    int
	ct   int
}

func (n node) ID() int64   { return n.id }
func (r Round) Total() int { return r.t + r.ct }

func (g RoundGraph) setRounds(rounds []RoundStart, lastTick int) {
	for i, round := range rounds {
		r := Round{
			startTick:   round.tick,
			roundNumber: round.t + round.ct + 1,
			t:           round.t,
			ct:          round.ct,
		}

		if int(i)+1 == len(rounds) {
			r.endTick = lastTick
		} else {
			r.endTick = rounds[i+1].tick - 1
		}

		g.addRound(r)
	}
}

func (g RoundGraph) setEdges() {
	nodes := g.Nodes()
	for nodes.Len() > 0 {
		nodes.Next()
		n := g.Node(nodes.Node().ID()) // n = current node
		cr := g.roundAtId(n.ID())      // get current round (cr)

		// loop through rounds to attach edges
		for _, r := range g.rounds {
			/*
				TODO:
				Overtime rounds
			*/
			if cr.startTick < r.startTick { // cr tick must be before the next round in the graph to attach edge to it i.e. can't go back in time
				if isFourteen(cr, r) {
					g.SetWeightedEdge(simple.WeightedEdge{F: n, T: g.nodeAtRound(r), W: 1})
				} else if hasIncreasedByOne(cr, r) {
					g.SetWeightedEdge(simple.WeightedEdge{F: n, T: g.nodeAtRound(r), W: float64(r.startTick - cr.startTick)})
				}
			}
		}
	}
}

func hasIncreasedByOne(r1, r2 Round) bool {
	if r1.Total() == r2.Total()-1 { // Ensure round increased by 1
		if r1.t == r2.t || r2.t == r1.t+1 { // Ensure t round only jumped up by 1
			if r1.ct == r2.ct || r2.ct == r1.ct+1 { // Ensure ct round only jumped up by 1
				return true
			}
		}
	}

	return false
}

func isFourteen(r1, r2 Round) bool {
	if r1.Total() == 14 { // if current total is 14, then the rounds will switch next round
		// t and ct will switch, one of them will +1 e.g. 4-10 becomes 11-4
		if r1.t == r2.ct && r1.ct == r2.t-1 {
			return true
		}

		if r1.ct == r2.t && r1.t == r2.ct-1 {
			return true
		}
	}

	return false
}

func (g RoundGraph) getMatchRounds() []Round {
	startRounds := []Round{}
	finalRound := Round{}
	for _, r := range g.rounds {
		//find all 0-0 rounds
		if r.Total() == 0 {
			startRounds = append(startRounds, r)
		}

		//find highest round
		if finalRound.Total() < r.Total() {
			finalRound = r
		}
	}

	// Find the shortest path from each starting round
	matchNodes := []graph.Node{}
	var finalRoundsWeight float64
	firstPass := true
	pt := path.DijkstraAllPaths(g)
	for _, startRound := range startRounds {
		path, weight, _ := pt.Between(g.idAtRound(startRound), g.idAtRound(finalRound))
		if firstPass {
			matchNodes = path
			finalRoundsWeight = weight
			firstPass = false
		} else {
			if weight < finalRoundsWeight {
				matchNodes = path
				finalRoundsWeight = weight
			}
		}
	}

	matchRounds := []Round{}
	for _, rNode := range matchNodes {
		matchRounds = append(matchRounds, g.roundAtId(rNode.ID()))
	}

	return matchRounds
}
