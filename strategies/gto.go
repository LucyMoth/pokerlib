package strategies

import (
	"math/rand"
	"sort"

	"github.com/LucyMoth/pokerlib"
)

type GTOStrategy struct {
	rng *rand.Rand
}

func NewGTOStrategy() *GTOStrategy {
	return &GTOStrategy{rng: rand.New(rand.NewSource(rand.Int63()))}
}

func NewGTOStrategyWithSeed(seed int64) *GTOStrategy {
	return &GTOStrategy{rng: rand.New(rand.NewSource(seed))}
}

func (s *GTOStrategy) Name() string {
	return "GTO"
}

func (s *GTOStrategy) Decide(ctx pokerlib.GameContext) pokerlib.Decision {
	if ctx.Street == pokerlib.Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *GTOStrategy) decidePreflop(ctx pokerlib.GameContext) pokerlib.Decision {
	strength := pokerlib.ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()
	isHeadsUp := ctx.PlayersInHand <= 2
	openSize := ctx.BigBlind*2 + ctx.BigBlind/2

	// Unopened pot (no raise yet)
	if toCall == 0 {
		switch strength {
		case pokerlib.PremiumHand:
			if s.rng.Float64() < 0.1 {
				return pokerlib.Decision{Action: pokerlib.Check}
			}
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
		case pokerlib.StrongHand:
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
		case pokerlib.PlayableHand:
			switch ctx.Position {
			case pokerlib.EarlyPosition:
				if s.rng.Float64() < 0.3 {
					return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
				}
				return pokerlib.Decision{Action: pokerlib.Check}
			case pokerlib.MiddlePosition:
				if s.rng.Float64() < 0.65 {
					return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
				}
				return pokerlib.Decision{Action: pokerlib.Check}
			default:
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
			}
		case pokerlib.MarginalHand:
			switch ctx.Position {
			case pokerlib.LatePosition:
				if s.rng.Float64() < 0.55 {
					return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
				}
			case pokerlib.Blinds:
				if isHeadsUp && s.rng.Float64() < 0.4 {
					return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
				}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		default:
			if ctx.Position == pokerlib.LatePosition && isHeadsUp && s.rng.Float64() < 0.25 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: openSize}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		}
	}

	// Facing a raise
	threeBetSize := toCall*3 + toCall/2

	switch strength {
	case pokerlib.PremiumHand:
		if ctx.RaisesThisRound < 3 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: threeBetSize}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	case pokerlib.StrongHand:
		if ctx.RaisesThisRound <= 1 && s.rng.Float64() < 0.35 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: threeBetSize}
		}
		if ctx.PotOdds() < 0.30 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
		if ctx.Position == pokerlib.EarlyPosition && ctx.RaisesThisRound >= 2 {
			return pokerlib.Decision{Action: pokerlib.Fold}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	case pokerlib.PlayableHand:
		if ctx.RaisesThisRound >= 2 {
			return pokerlib.Decision{Action: pokerlib.Fold}
		}
		if ctx.Position >= pokerlib.MiddlePosition && ctx.PotOdds() < 0.22 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
		if ctx.Position == pokerlib.LatePosition && s.rng.Float64() < 0.12 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: threeBetSize}
		}
		if ctx.PotOdds() < 0.18 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
		return pokerlib.Decision{Action: pokerlib.Fold}
	case pokerlib.MarginalHand:
		if ctx.Position == pokerlib.Blinds && ctx.PotOdds() < 0.15 {
			if s.rng.Float64() < 0.4 {
				return pokerlib.Decision{Action: pokerlib.Call}
			}
		}
		return pokerlib.Decision{Action: pokerlib.Fold}
	default:
		return pokerlib.Decision{Action: pokerlib.Fold}
	}
}

func (s *GTOStrategy) decidePostflop(ctx pokerlib.GameContext) pokerlib.Decision {
	allCards := make([]pokerlib.Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := pokerlib.EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := pokerlib.AnalyzeDraws(ctx.Hand, ctx.Community)

	handValue := s.assessHandValue(result, ctx, draws)
	isHeadsUp := ctx.PlayersInHand <= 2

	if toCall == 0 {
		switch {
		case handValue >= 0.85:
			if s.rng.Float64() < 0.2 {
				return pokerlib.Decision{Action: pokerlib.Check}
			}
			betSize := ctx.Pot * 3 / 4
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: betSize}
		case handValue >= 0.6:
			betSize := ctx.Pot * 2 / 3
			if s.rng.Float64() < 0.85 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: betSize}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		case handValue >= 0.35:
			freq := 0.55
			if isHeadsUp {
				freq = 0.7
			}
			if s.rng.Float64() < freq {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		case draws.HasDraw():
			freq := 0.45
			if draws.OpenEndedStraight || draws.FlushDraw {
				freq = 0.55
			}
			if isHeadsUp {
				freq += 0.15
			}
			if s.rng.Float64() < freq {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		default:
			bluffFreq := 0.15
			if isHeadsUp {
				bluffFreq = 0.30
			}
			if ctx.Position == pokerlib.LatePosition {
				bluffFreq += 0.10
			}
			if s.rng.Float64() < bluffFreq {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 3}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		}
	}

	potOdds := ctx.PotOdds()
	effectiveValue := handValue
	if draws.HasDraw() {
		effectiveValue += draws.DrawEquity()
		if effectiveValue > 1.0 {
			effectiveValue = 1.0
		}
	}

	if effectiveValue > potOdds {
		if handValue >= 0.7 && s.rng.Float64() < 0.45 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall*2 + ctx.Pot/2}
		}
		if handValue >= 0.5 && s.rng.Float64() < 0.25 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 2}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if draws.HasDraw() && draws.Outs >= 8 && s.rng.Float64() < 0.3 {
		return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 2}
	}

	if effectiveValue > potOdds*0.7 && s.rng.Float64() < 0.25 {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}

func (s *GTOStrategy) assessHandValue(result pokerlib.HandResult, ctx pokerlib.GameContext, draws pokerlib.DrawInfo) float64 {
	baseValue := 0.0
	switch result.Rank {
	case pokerlib.RoyalFlush, pokerlib.StraightFlush:
		return 1.0
	case pokerlib.FourOfAKind:
		return 0.97
	case pokerlib.FullHouse:
		baseValue = 0.90
	case pokerlib.Flush:
		baseValue = 0.80
		if len(result.HighCards) > 0 && result.HighCards[0] >= pokerlib.Queen {
			baseValue = 0.85
		}
	case pokerlib.Straight:
		baseValue = 0.72
	case pokerlib.ThreeOfAKind:
		baseValue = 0.62
		if s.isSetOnBoard(result, ctx) {
			baseValue = 0.55
		}
	case pokerlib.TwoPair:
		baseValue = 0.48
		if s.isTopTwoPair(result, ctx) {
			baseValue = 0.55
		}
	case pokerlib.OnePair:
		baseValue = s.assessPairValue(result, ctx)
	default:
		if len(result.HighCards) > 0 && result.HighCards[0] >= pokerlib.Queen {
			baseValue = 0.12
		} else {
			baseValue = 0.05
		}
	}

	if len(ctx.Community) >= 3 {
		paired := s.boardPaired(ctx.Community)
		monotone := s.boardMonotone(ctx.Community)
		connected := s.boardConnected(ctx.Community)

		if paired && result.Rank <= pokerlib.OnePair {
			baseValue *= 0.85
		}
		if monotone && result.Rank < pokerlib.Flush {
			baseValue *= 0.80
		}
		if connected && result.Rank < pokerlib.Straight {
			baseValue *= 0.90
		}
	}

	if ctx.PlayersInHand > 2 && result.Rank <= pokerlib.OnePair {
		baseValue *= 0.85
	}
	if ctx.PlayersInHand > 3 && result.Rank <= pokerlib.TwoPair {
		baseValue *= 0.90
	}

	return baseValue
}

func (s *GTOStrategy) assessPairValue(result pokerlib.HandResult, ctx pokerlib.GameContext) float64 {
	if len(result.HighCards) == 0 || len(ctx.Community) == 0 {
		return 0.25
	}
	pairRank := result.HighCards[0]

	maxBoard := pokerlib.Rank(0)
	for _, c := range ctx.Community {
		if c.Rank > maxBoard {
			maxBoard = c.Rank
		}
	}

	if pairRank > maxBoard {
		if pairRank >= pokerlib.Queen {
			return 0.45
		}
		return 0.38
	}

	if pairRank == maxBoard {
		if len(result.HighCards) > 1 && result.HighCards[1] >= pokerlib.Ten {
			return 0.35
		}
		return 0.28
	}

	return 0.18
}

func (s *GTOStrategy) isSetOnBoard(result pokerlib.HandResult, ctx pokerlib.GameContext) bool {
	if result.Rank != pokerlib.ThreeOfAKind || len(result.HighCards) == 0 {
		return false
	}
	tripRank := result.HighCards[0]
	count := 0
	for _, c := range ctx.Community {
		if c.Rank == tripRank {
			count++
		}
	}
	return count >= 2
}

func (s *GTOStrategy) isTopTwoPair(result pokerlib.HandResult, ctx pokerlib.GameContext) bool {
	if result.Rank != pokerlib.TwoPair || len(result.HighCards) < 2 {
		return false
	}
	maxBoard := pokerlib.Rank(0)
	for _, c := range ctx.Community {
		if c.Rank > maxBoard {
			maxBoard = c.Rank
		}
	}
	return result.HighCards[0] >= maxBoard
}

func (s *GTOStrategy) boardPaired(community []pokerlib.Card) bool {
	seen := make(map[pokerlib.Rank]bool)
	for _, c := range community {
		if seen[c.Rank] {
			return true
		}
		seen[c.Rank] = true
	}
	return false
}

func (s *GTOStrategy) boardMonotone(community []pokerlib.Card) bool {
	if len(community) < 3 {
		return false
	}
	suit := community[0].Suit
	count := 0
	for _, c := range community {
		if c.Suit == suit {
			count++
		}
	}
	return count >= 3
}

func (s *GTOStrategy) boardConnected(community []pokerlib.Card) bool {
	if len(community) < 3 {
		return false
	}
	ranks := make([]int, len(community))
	for i, c := range community {
		ranks[i] = int(c.Rank)
	}
	sort.Ints(ranks)
	connCount := 0
	for i := 1; i < len(ranks); i++ {
		if ranks[i]-ranks[i-1] <= 2 {
			connCount++
		}
	}
	return connCount >= 2
}
