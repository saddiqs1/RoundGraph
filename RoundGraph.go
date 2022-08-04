package RoundGraph

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

func NewRoundGraph() RoundGraph {
	return RoundGraph{
		ids:                   make(map[Round]int64),
		rounds:                make(map[int64]Round),
		WeightedDirectedGraph: simple.NewWeightedDirectedGraph(0, 0),
	}
}

func (g RoundGraph) SetRounds(rounds []RoundStart, scores []ScoreUpdate, gameHalftimes []int, lastTick int) {
	for i, round := range rounds {
		r := Round{
			StartTick:   round.Tick,
			RoundNumber: round.T + round.CT + 1,
			T:           round.T,
			CT:          round.CT,
			IsHalfTime:  false,
		}

		if int(i)+1 == len(rounds) {
			r.EndTick = lastTick
		} else {
			r.EndTick = rounds[i+1].Tick - 1
		}

		// Update round scores (some demos don't instantly update round scores when round is started)
		for _, s := range scores {
			if r.StartTick <= s.Tick && s.Tick <= r.StartTick+500 {
				if s.Team == "t" && s.Score != r.T {
					r.T = s.Score
					r.RoundNumber = r.Total() + 1
				} else if s.Team == "ct" && s.Score != r.CT {
					r.CT = s.Score
					r.RoundNumber = r.Total() + 1
				}
			}
		}

		// Set halftime bool using gameHalfEnded events
		for _, gh := range gameHalftimes {
			if r.StartTick <= gh && gh <= r.EndTick {
				r.IsHalfTime = true
			}
		}

		// Some servers (e.g. esea) don't log gameHalfEnded events, so hard code it as below
		if len(gameHalftimes) == 0 {
			x := (r.RoundNumber - 30) % 3 //TODO - check when the scores flip for the first time in OT to figure out it MR 6 or 10
			if r.RoundNumber == 15 {
				r.IsHalfTime = true
			} else if r.RoundNumber > 30 && x == 0 {
				r.IsHalfTime = true
			}
		}

		g.addRound(r)
	}
}

func (g RoundGraph) SetEdges() {
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
			if cr.StartTick < r.StartTick { // cr Tick must be before the next round in the graph to attach edge to it i.e. can't go back in time
				if isHalfTime(cr, r) {
					g.SetWeightedEdge(simple.WeightedEdge{F: n, T: g.nodeAtRound(r), W: 1})
				} else if hasIncreasedByOne(cr, r) {
					g.SetWeightedEdge(simple.WeightedEdge{F: n, T: g.nodeAtRound(r), W: float64(r.StartTick - cr.StartTick)})
				}
			}
		}
	}
}

func (g RoundGraph) GetMatchRounds() []Round {
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
		if len(path) > 0 {
			if firstPass {
				matchNodes = path
				finalRoundsWeight = weight
				firstPass = false
			} else if weight < finalRoundsWeight {
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

func (g RoundGraph) addRound(r Round) {
	n := g.NewNode()
	id := n.ID()
	n = node{
		Round: r,
		Id:    id,
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

func (n node) ID() int64   { return n.Id }
func (r Round) Total() int { return r.T + r.CT }

func hasIncreasedByOne(r1, r2 Round) bool {
	if r1.Total() == r2.Total()-1 { // Ensure round increased by 1
		if r1.T == r2.T || r2.T == r1.T+1 { // Ensure t round only jumped up by 1
			if r1.CT == r2.CT || r2.CT == r1.CT+1 { // Ensure CT round only jumped up by 1
				return true
			}
		}
	}

	return false
}

func isHalfTime(r1, r2 Round) bool {
	if r1.IsHalfTime { // if current round is halftime, then the rounds will switch next round
		// t and CT will switch, one of them will +1 e.g. 4-10 becomes 11-4
		if r1.T == r2.CT && r1.CT == r2.T-1 {
			return true
		}

		if r1.CT == r2.T && r1.T == r2.CT-1 {
			return true
		}
	}

	return false
}
