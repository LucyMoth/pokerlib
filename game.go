package pokerlib

import (
	"math/rand"
	"time"
)

type GameConfig struct {
	SmallBlind    int
	BigBlind      int
	StartingChips int
	MaxPlayers    int
}

func DefaultConfig() GameConfig {
	return GameConfig{
		SmallBlind:    25,
		BigBlind:      50,
		StartingChips: 1000,
		MaxPlayers:    9,
	}
}

type Game struct {
	Table  *Table
	Config GameConfig
	Rng    *rand.Rand
}

func NewGame(config GameConfig) *Game {
	return &Game{
		Table:  NewTable(config.SmallBlind, config.BigBlind),
		Config: config,
		Rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func NewGameWithSeed(config GameConfig, seed int64) *Game {
	return &Game{
		Table:  NewTable(config.SmallBlind, config.BigBlind),
		Config: config,
		Rng:    rand.New(rand.NewSource(seed)),
	}
}

func (g *Game) AddPlayer(name string) *Player {
	if len(g.Table.Players) >= g.Config.MaxPlayers {
		return nil
	}
	p := NewPlayer(name, g.Config.StartingChips)
	g.Table.AddPlayer(p)
	return p
}

func (g *Game) PlayHand() {
	g.Table.StartHand()
}

func (g *Game) GetTable() *Table {
	return g.Table
}

type SimulationResult struct {
	Wins   int
	Ties   int
	Losses int
	Total  int
}

func (s SimulationResult) WinRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Wins) / float64(s.Total)
}

func (s SimulationResult) Equity() float64 {
	if s.Total == 0 {
		return 0
	}
	return (float64(s.Wins) + float64(s.Ties)/2) / float64(s.Total)
}

func SimulateHeadsUp(hand1, hand2 [2]Card, community []Card, iterations int) (SimulationResult, SimulationResult) {
	var result1, result2 SimulationResult

	for i := 0; i < iterations; i++ {
		deck := NewDeck()

		deck.RemoveCard(hand1[0])
		deck.RemoveCard(hand1[1])
		deck.RemoveCard(hand2[0])
		deck.RemoveCard(hand2[1])

		for _, c := range community {
			deck.RemoveCard(c)
		}

		deck.Shuffle()

		fullCommunity := make([]Card, len(community))
		copy(fullCommunity, community)

		for len(fullCommunity) < 5 {
			fullCommunity = append(fullCommunity, deck.Deal())
		}

		cards1 := append(hand1[:], fullCommunity...)
		cards2 := append(hand2[:], fullCommunity...)

		eval1 := EvaluateHand(cards1)
		eval2 := EvaluateHand(cards2)

		cmp := eval1.Compare(eval2)
		result1.Total++
		result2.Total++

		if cmp > 0 {
			result1.Wins++
			result2.Losses++
		} else if cmp < 0 {
			result1.Losses++
			result2.Wins++
		} else {
			result1.Ties++
			result2.Ties++
		}
	}

	return result1, result2
}

func SimulateHand(hand [2]Card, community []Card, opponents int, iterations int) SimulationResult {
	var result SimulationResult

	for i := 0; i < iterations; i++ {
		deck := NewDeck()
		deck.RemoveCard(hand[0])
		deck.RemoveCard(hand[1])

		for _, c := range community {
			deck.RemoveCard(c)
		}

		deck.Shuffle()

		oppHands := make([][2]Card, opponents)
		for j := 0; j < opponents; j++ {
			oppHands[j][0] = deck.Deal()
			oppHands[j][1] = deck.Deal()
		}

		fullCommunity := make([]Card, len(community))
		copy(fullCommunity, community)

		for len(fullCommunity) < 5 {
			fullCommunity = append(fullCommunity, deck.Deal())
		}

		heroCards := append(hand[:], fullCommunity...)
		heroEval := EvaluateHand(heroCards)

		won := true
		tied := false
		for _, oppHand := range oppHands {
			oppCards := append(oppHand[:], fullCommunity...)
			oppEval := EvaluateHand(oppCards)
			cmp := heroEval.Compare(oppEval)
			if cmp < 0 {
				won = false
				tied = false
				break
			} else if cmp == 0 {
				tied = true
			}
		}

		result.Total++
		if won && !tied {
			result.Wins++
		} else if tied {
			result.Ties++
		} else {
			result.Losses++
		}
	}

	return result
}
