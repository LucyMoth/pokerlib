package strategies

import (
	"math/rand"

	"github.com/LucyMoth/pokerlib"
)

type FishStrategy struct {
	rng        *rand.Rand
	callFreq   float64
	bluffFreq  float64
	chaseDraws bool
}

func NewFishStrategy() *FishStrategy {
	return &FishStrategy{
		rng:        rand.New(rand.NewSource(rand.Int63())),
		callFreq:   0.7,
		bluffFreq:  0.1,
		chaseDraws: true,
	}
}

func NewFishStrategyCustom(callFreq, bluffFreq float64, chaseDraws bool) *FishStrategy {
	return &FishStrategy{
		rng:        rand.New(rand.NewSource(rand.Int63())),
		callFreq:   callFreq,
		bluffFreq:  bluffFreq,
		chaseDraws: chaseDraws,
	}
}

func (s *FishStrategy) Name() string {
	return "Fish"
}

func (s *FishStrategy) Decide(ctx pokerlib.GameContext) pokerlib.Decision {
	if ctx.Street == pokerlib.Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *FishStrategy) decidePreflop(ctx pokerlib.GameContext) pokerlib.Decision {
	strength := pokerlib.ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if strength <= pokerlib.StrongHand {
			if s.rng.Float64() < 0.3 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind * 2}
			}
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind * 3}
		}
		if s.rng.Float64() < 0.4 {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind * 2}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	if strength <= pokerlib.PlayableHand {
		if toCall > ctx.Pot && s.rng.Float64() < 0.2 {
			return pokerlib.Decision{Action: pokerlib.Fold}
		}
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if s.rng.Float64() < s.callFreq {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}

func (s *FishStrategy) decidePostflop(ctx pokerlib.GameContext) pokerlib.Decision {
	allCards := make([]pokerlib.Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := pokerlib.EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := pokerlib.AnalyzeDraws(ctx.Hand, ctx.Community)

	hasPair := result.Rank >= pokerlib.OnePair
	hasStrongHand := result.Rank >= pokerlib.TwoPair

	if toCall == 0 {
		if hasStrongHand {
			if s.rng.Float64() < 0.4 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind}
			}
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.Pot / 2}
		}
		if hasPair {
			if s.rng.Float64() < 0.5 {
				return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind}
			}
			return pokerlib.Decision{Action: pokerlib.Check}
		}
		if s.rng.Float64() < s.bluffFreq {
			return pokerlib.Decision{Action: pokerlib.Raise, Amount: ctx.BigBlind}
		}
		return pokerlib.Decision{Action: pokerlib.Check}
	}

	if hasStrongHand {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if hasPair {
		if toCall <= ctx.Pot/2 || s.rng.Float64() < 0.6 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
		return pokerlib.Decision{Action: pokerlib.Fold}
	}

	if s.chaseDraws && draws.HasDraw() {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	if s.chaseDraws && s.hasOvercards(ctx) {
		if s.rng.Float64() < 0.5 {
			return pokerlib.Decision{Action: pokerlib.Call}
		}
	}

	if s.rng.Float64() < s.callFreq*0.4 {
		return pokerlib.Decision{Action: pokerlib.Call}
	}

	return pokerlib.Decision{Action: pokerlib.Fold}
}

func (s *FishStrategy) hasOvercards(ctx pokerlib.GameContext) bool {
	if len(ctx.Community) == 0 {
		return false
	}
	maxBoard := pokerlib.Rank(0)
	for _, c := range ctx.Community {
		if c.Rank > maxBoard {
			maxBoard = c.Rank
		}
	}
	return ctx.Hand[0].Rank > maxBoard || ctx.Hand[1].Rank > maxBoard
}
