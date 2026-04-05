package strategies

import (
	"math/rand"

	"github.com/LucyMoth/pokerlib"
)

type LAGStrategy struct {
	rng        *rand.Rand
	aggression float64
}

func NewLAGStrategy() *LAGStrategy {
	return &LAGStrategy{
		rng:        rand.New(rand.NewSource(rand.Int63())),
		aggression: 0.75,
	}
}

func (s *LAGStrategy) Name() string {
	return "LAG"
}

func (s *LAGStrategy) Decide(ctx pokerlib.GameContext) pokerlib.Decision {
	if ctx.Street == pokerlib.Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *LAGStrategy) decidePreflop(ctx pokerlib.GameContext) pokerlib.Decision {
	strength := pokerlib.ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if strength <= pokerlib.MarginalHand || s.rng.Float64() < 0.4 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind * 3}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	if strength <= pokerlib.PlayableHand {
		if s.rng.Float64() < s.aggression {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 3}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if strength == pokerlib.MarginalHand && ctx.Position >= pokerlib.MiddlePosition {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if s.rng.Float64() < 0.2 {
		return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 3}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}

func (s *LAGStrategy) decidePostflop(ctx pokerlib.GameContext) pokerlib.Decision {
	allCards := make([]pokerlib.Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := pokerlib.EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := pokerlib.AnalyzeDraws(ctx.Hand, ctx.Community)

	if toCall == 0 {
		if result.Rank >= pokerlib.TwoPair {
			betSize := ctx.Pot*2/3 + ctx.Pot/4
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: betSize}
		}
		if result.Rank >= pokerlib.OnePair {
			betSize := ctx.Pot/2 + ctx.Pot/4
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: betSize}
		}
		if draws.HasDraw() {
			if s.rng.Float64() < s.aggression*0.7 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot * 2 / 3}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		}
		bluffFreq := s.aggression * 0.5
		if ctx.Position == pokerlib.LatePosition {
			bluffFreq += 0.15
		}
		if ctx.PlayersInHand <= 2 {
			bluffFreq += 0.10
		}
		if s.rng.Float64() < bluffFreq {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	if result.Rank >= pokerlib.TwoPair {
		if s.rng.Float64() < s.aggression {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 3}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if result.Rank >= pokerlib.OnePair {
		if s.rng.Float64() < s.aggression*0.3 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 2}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if draws.HasDraw() {
		if draws.Outs >= 8 && s.rng.Float64() < s.aggression*0.5 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 2}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if ctx.PlayersInHand <= 2 && s.rng.Float64() < s.aggression*0.25 {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}
