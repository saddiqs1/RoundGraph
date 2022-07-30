package main

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type ScoreGraph struct {
	ids    map[Score]int64 //TODO - maybe could move id into Score object, would remove need for these maps
	scores map[int64]Score
	*simple.WeightedDirectedGraph
}

func newScoreGraph() ScoreGraph {
	return ScoreGraph{
		ids:                   make(map[Score]int64),
		scores:                make(map[int64]Score),
		WeightedDirectedGraph: simple.NewWeightedDirectedGraph(0, 0),
	}
}

func (g ScoreGraph) addScore(s Score) {
	n := g.NewNode()
	id := n.ID()
	n = node{
		score: s,
		id:    id,
	}
	g.AddNode(n)
	g.ids[s] = id
	g.scores[id] = s
}

func (g ScoreGraph) nodeAtScore(s Score) graph.Node {
	id, ok := g.ids[s]
	if !ok {
		return nil
	}
	return g.WeightedDirectedGraph.Node(id)
}

func (g ScoreGraph) scoreAtId(id int64) Score {
	return g.scores[id]
}

func (g ScoreGraph) idAtScore(s Score) int64 {
	return g.ids[s]
}

type node struct {
	score Score
	id    int64
}

func (n node) ID() int64 { return n.id }

type Score struct {
	tick int
	t    int
	ct   int
}

func (s Score) Total() int { return s.t + s.ct }

func (sg ScoreGraph) scoreToRound(s Score, lastTick int) Round {
	for i, score := range sg.scores {
		if score == s {
			if int(i)+1 == len(sg.scores) {
				return Round{
					startTick:   s.tick,
					endTick:     lastTick,
					roundNumber: s.t+ s.ct + 1,
					t:           s.t,
					ct:          s.ct,
				}
			}

			return Round{
				startTick:   s.tick,
				endTick:     sg.scores[i+1].tick - 1,
				roundNumber: s.t + s.ct + 1,
				t:           s.t,
				ct:          s.ct,
			}
		}
	}

	return Round{}
}

type Round struct {
	startTick   int
	endTick     int
	roundNumber int
	t           int
	ct          int
}
