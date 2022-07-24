package main

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type scoreGraph struct {
	ids    map[Score]int64
	scores map[int64]Score
	*simple.DirectedGraph
}

func newScoreGraph() scoreGraph {
	return scoreGraph{
		ids:           make(map[Score]int64),
		scores:        make(map[int64]Score),
		DirectedGraph: simple.NewDirectedGraph(),
	}
}

func (g scoreGraph) nodeAtScore(score Score) graph.Node {
	id, ok := g.ids[score]
	if !ok {
		return nil
	}
	return g.DirectedGraph.Node(id)
}

func (g scoreGraph) scoreAtId(id int64) Score {
	return g.scores[id]
}

type node struct {
	score Score
	id    int64
}

func (n node) ID() int64    { return n.id }
func (n node) Score() Score { return n.score }

type Score struct {
	tick int
	t    int
	ct   int
}

func (s Score) Total() int { return s.t + s.ct }
