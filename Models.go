package RoundGraph

type Round struct {
	StartTick   int
	EndTick     int
	RoundNumber int
	T           int
	CT          int
	IsHalfTime  bool
}

type node struct {
	Round Round
	Id    int64
}

type RoundStart struct {
	Tick int
	T    int
	CT   int
}

type ScoreUpdate struct {
	Tick  int
	Team  string
	Score int
}
