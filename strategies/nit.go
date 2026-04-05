package strategies

import (
	"math/rand"

	"github.com/LucyMoth/pokerlib"
)

type NitStrategy struct {
	rng *rand.Rand
}

func NewNitStrategy() *NitStrategy {
	return &NitStrategy{rng: rand.New(rand.NewSource(rand.Int63()))}
}

func (s *NitStrategy) Name() string {
	return "Nit"
}

func (s *NitStrategy) Decide(ctx pokerlib.GameContext) pokerlib.Decision {
	if ctx.Street == pokerlib.Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *NitStrategy) decidePreflop(ctx pokerlib.GameContext) pokerlib.Decision {
	strength := pokerlib.ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if strength <= pokerlib.StrongHand {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind * 4}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	if strength == pokerlib.PremiumHand {
		return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 3}
	}
	if strength == pokerlib.StrongHand && ctx.PotOdds() < 0.2 {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}

func (s *NitStrategy) decidePostflop(ctx pokerlib.GameContext) pokerlib.Decision {
	allCards := make([]pokerlib.Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := pokerlib.EvaluateHand(allCards)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if result.Rank >= pokerlib.ThreeOfAKind {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot * 2 / 3}
		}
		if result.Rank >= pokerlib.TwoPair {
			if s.rng.Float64() < 0.6 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		}
		if result.Rank == pokerlib.OnePair && len(result.HighCards) > 0 {
			maxBoard := pokerlib.Rank(0)
			for _, c := range ctx.Community {
				if c.Rank > maxBoard {
					maxBoard = c.Rank
				}
			}
			if result.HighCards[0] > maxBoard && s.rng.Float64() < 0.4 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
			}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	if result.Rank >= pokerlib.ThreeOfAKind {
		if s.rng.Float64() < 0.5 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 2}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if result.Rank >= pokerlib.TwoPair {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if result.Rank == pokerlib.OnePair && len(result.HighCards) > 0 && result.HighCards[0] >= pokerlib.Queen {
		if ctx.PotOdds() < 0.25 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}
