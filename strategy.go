package pokerlib

type Position int

const (
	EarlyPosition Position = iota
	MiddlePosition
	LatePosition
	Blinds
)

func (p Position) String() string {
	return [...]string{"Early", "Middle", "Late", "Blinds"}[p]
}

type GameContext struct {
	Hand            [2]Card
	Community       []Card
	Street          Street
	Pot             int
	CurrentBet      int
	PlayerBet       int
	PlayerChips     int
	Position        Position
	PlayersInHand   int
	RaisesThisRound int
	BigBlind        int
}

func (g GameContext) ToCall() int {
	return g.CurrentBet - g.PlayerBet
}

func (g GameContext) PotOdds() float64 {
	toCall := g.ToCall()
	if toCall == 0 {
		return 0
	}
	return float64(toCall) / float64(g.Pot+toCall)
}

func (g GameContext) StackToPot() float64 {
	if g.Pot == 0 {
		return float64(g.PlayerChips) / float64(g.BigBlind)
	}
	return float64(g.PlayerChips) / float64(g.Pot)
}

type Decision struct {
	Action Action
	Amount int
}

type Strategy interface {
	Decide(ctx GameContext) Decision
	Name() string
}

type PreflopHandStrength int

const (
	PremiumHand PreflopHandStrength = iota
	StrongHand
	PlayableHand
	MarginalHand
	WeakHand
)

func ClassifyPreflopHand(hand [2]Card) PreflopHandStrength {
	high, low := hand[0].Rank, hand[1].Rank
	if low > high {
		high, low = low, high
	}
	suited := hand[0].Suit == hand[1].Suit
	pair := high == low
	gap := int(high) - int(low)
	connector := gap == 1 || (high == Ace && low == Two)

	// Pairs
	if pair {
		if high >= Queen {
			return PremiumHand
		}
		if high >= Ten {
			return StrongHand
		}
		if high >= Six {
			return PlayableHand
		}
		if high >= Two {
			return MarginalHand
		}
	}

	// Ace-high hands
	if high == Ace {
		if low == King {
			if suited {
				return PremiumHand
			}
			return StrongHand
		}
		if low >= Queen {
			if suited {
				return StrongHand
			}
			return PlayableHand
		}
		if low >= Ten {
			if suited {
				return StrongHand
			}
			return PlayableHand
		}
		if suited {
			if low >= Six {
				return PlayableHand
			}
			return MarginalHand
		}
		return MarginalHand
	}

	// King-high hands
	if high == King {
		if low == Queen {
			if suited {
				return StrongHand
			}
			return PlayableHand
		}
		if low == Jack {
			if suited {
				return StrongHand
			}
			return PlayableHand
		}
		if low == Ten {
			if suited {
				return PlayableHand
			}
			return MarginalHand
		}
		if suited && low >= Nine {
			return MarginalHand
		}
		return WeakHand
	}

	// Queen-high hands
	if high == Queen {
		if low == Jack {
			if suited {
				return StrongHand
			}
			return PlayableHand
		}
		if low == Ten {
			if suited {
				return PlayableHand
			}
			return MarginalHand
		}
		if low == Nine && suited {
			return MarginalHand
		}
		return WeakHand
	}

	// Suited connectors and one-gappers
	if suited {
		if connector && high >= Eight {
			return PlayableHand
		}
		if connector && high >= Five {
			return MarginalHand
		}
		if gap == 2 && high >= Nine {
			return MarginalHand
		}
		if high == Jack && low >= Nine {
			return MarginalHand
		}
	}

	// Offsuit connectors
	if !suited {
		if connector && high >= Ten {
			return MarginalHand
		}
		if high == Jack && low == Ten {
			return MarginalHand
		}
	}

	return WeakHand
}

func GetPosition(playerIndex, dealerPos, numPlayers int) Position {
	relativePos := (playerIndex - dealerPos + numPlayers) % numPlayers

	if numPlayers <= 3 {
		if relativePos <= 1 {
			return Blinds
		}
		return LatePosition
	}

	earlyCount := numPlayers / 3
	middleCount := numPlayers / 3

	if relativePos <= 2 {
		return Blinds
	}
	if relativePos <= 2+earlyCount {
		return EarlyPosition
	}
	if relativePos <= 2+earlyCount+middleCount {
		return MiddlePosition
	}
	return LatePosition
}

func BuildGameContext(player *Player, table *Table, playerIndex int) GameContext {
	return GameContext{
		Hand:            player.Hand,
		Community:       table.Community,
		Street:          table.Street,
		Pot:             table.Pot,
		CurrentBet:      table.CurrentBet,
		PlayerBet:       player.Bet,
		PlayerChips:     player.Chips,
		Position:        GetPosition(playerIndex, table.DealerPos, len(table.Players)),
		PlayersInHand:   table.ActivePlayers,
		RaisesThisRound: table.RaisesThisStreet,
		BigBlind:        table.BigBlind,
	}
}
