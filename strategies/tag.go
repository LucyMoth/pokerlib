package strategies

import (
	"math/rand"

	"github.com/LucyMoth/pokerlib"
)

type TAGStrategy struct {
	rng        *rand.Rand
	aggression float64
	tightness  float64
}

func NewTAGStrategy() *TAGStrategy {
	return &TAGStrategy{
		rng:        rand.New(rand.NewSource(rand.Int63())),
		aggression: 0.6,
		tightness:  0.7,
	}
}

func NewTAGStrategyCustom(aggression, tightness float64) *TAGStrategy {
	return &TAGStrategy{
		rng:        rand.New(rand.NewSource(rand.Int63())),
		aggression: aggression,
		tightness:  tightness,
	}
}

func (s *TAGStrategy) Name() string {
	return "TAG"
}

func (s *TAGStrategy) Decide(ctx pokerlib.GameContext) pokerlib.Decision {
	if ctx.Street == pokerlib.Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *TAGStrategy) decidePreflop(ctx pokerlib.GameContext) pokerlib.Decision {
	strength := pokerlib.ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	playThreshold := pokerlib.MarginalHand
	if s.tightness > 0.7 {
		playThreshold = pokerlib.PlayableHand
	}
	if s.tightness > 0.85 {
		playThreshold = pokerlib.StrongHand
	}

	if toCall == 0 {
		if strength <= playThreshold {
			raiseAmount := ctx.BigBlind * 3
			if ctx.Position == pokerlib.LatePosition {
				raiseAmount = ctx.BigBlind*2 + ctx.BigBlind/2
			}
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: raiseAmount}
		}
		if ctx.Position == pokerlib.LatePosition && strength == pokerlib.MarginalHand {
			if s.rng.Float64() < 0.4 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind*2 + ctx.BigBlind/2}
			}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	switch strength {
	case pokerlib.PremiumHand:
		if s.rng.Float64() < s.aggression {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 3}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	case pokerlib.StrongHand:
		if ctx.RaisesThisRound == 0 && s.rng.Float64() < s.aggression*0.5 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 3}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	case pokerlib.PlayableHand:
		if ctx.Position >= pokerlib.MiddlePosition && ctx.PotOdds() < 0.25 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
		return pokerlib.Decision{Action: pokerlib.Fold}
	default:
		return pokerlib.Decision{Action: pokerlib.Fold}
	}
}

func (s *TAGStrategy) decidePostflop(ctx pokerlib.GameContext) pokerlib.Decision {
	allCards := make([]pokerlib.Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := pokerlib.EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := pokerlib.AnalyzeDraws(ctx.Hand, ctx.Community)

	if toCall == 0 {
		if result.Rank >= pokerlib.TwoPair {
			betSize := int(float64(ctx.Pot) * (0.5 + s.aggression*0.3))
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: betSize}
		}
		if result.Rank >= pokerlib.OnePair {
			if s.rng.Float64() < s.aggression*0.7 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		}
		if draws.HasDraw() && draws.Outs >= 8 {
			if s.rng.Float64() < s.aggression*0.6 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
			}
		}
		if ctx.Street == pokerlib.Flop && ctx.PlayersInHand <= 3 {
			if s.rng.Float64() < s.aggression*0.4 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 3}
			}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	if result.Rank >= pokerlib.ThreeOfAKind {
		if s.rng.Float64() < s.aggression {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: toCall * 2}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if result.Rank >= pokerlib.OnePair {
		if ctx.PotOdds() < 0.35 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
		return pokerlib.Decision{Action: pokerlib.Fold}
	}

	if draws.HasDraw() {
		equity := draws.DrawEquity()
		if equity > ctx.PotOdds()*0.8 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}
