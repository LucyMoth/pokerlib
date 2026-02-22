package pokerlib

import (
	"math/rand"
)

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

	if pair {
		if high >= Queen {
			return PremiumHand
		}
		if high >= Nine {
			return StrongHand
		}
		if high >= Six {
			return PlayableHand
		}
		return MarginalHand
	}

	if high == Ace {
		if low >= King {
			return PremiumHand
		}
		if low >= Ten || suited {
			return StrongHand
		}
		return PlayableHand
	}

	if high == King {
		if low >= Queen {
			if suited {
				return StrongHand
			}
			return PlayableHand
		}
		if low >= Ten && suited {
			return PlayableHand
		}
		return MarginalHand
	}

	gap := int(high) - int(low)
	if gap <= 2 && high >= Ten && suited {
		return PlayableHand
	}
	if gap == 1 && high >= Nine && suited {
		return PlayableHand
	}

	if suited && high >= Queen && low >= Ten {
		return PlayableHand
	}

	return WeakHand
}

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

func (s *GTOStrategy) Decide(ctx GameContext) Decision {
	if ctx.Street == Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *GTOStrategy) decidePreflop(ctx GameContext) Decision {
	strength := ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()
	isHeadsUp := ctx.PlayersInHand <= 2

	if toCall == 0 {
		switch strength {
		case PremiumHand:
			return Decision{Action: Raise, Amount: ctx.BigBlind * 3}
		case StrongHand:
			return Decision{Action: Raise, Amount: ctx.BigBlind * 3}
		case PlayableHand:
			if ctx.Position >= MiddlePosition || isHeadsUp {
				return Decision{Action: Raise, Amount: ctx.BigBlind * 3}
			}
			if s.rng.Float64() < 0.4 {
				return Decision{Action: Raise, Amount: ctx.BigBlind * 3}
			}
			return Decision{Action: Check}
		case MarginalHand:
			if ctx.Position == LatePosition || isHeadsUp {
				if s.rng.Float64() < 0.6 {
					return Decision{Action: Raise, Amount: ctx.BigBlind*2 + ctx.BigBlind/2}
				}
			}
			return Decision{Action: Check}
		default:
			return Decision{Action: Check}
		}
	}

	switch strength {
	case PremiumHand:
		if ctx.RaisesThisRound < 3 {
			return Decision{Action: Raise, Amount: toCall * 3}
		}
		return Decision{Action: Call}
	case StrongHand:
		if ctx.RaisesThisRound == 0 {
			if s.rng.Float64() < 0.4 {
				return Decision{Action: Raise, Amount: toCall * 3}
			}
		}
		if ctx.PotOdds() < 0.3 {
			return Decision{Action: Call}
		}
		return Decision{Action: Fold}
	case PlayableHand:
		if ctx.RaisesThisRound == 0 && ctx.PotOdds() < 0.2 {
			return Decision{Action: Call}
		}
		return Decision{Action: Fold}
	default:
		return Decision{Action: Fold}
	}
}

func (s *GTOStrategy) decidePostflop(ctx GameContext) Decision {
	allCards := append(ctx.Hand[:], ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()

	handValue := s.assessHandValue(result, ctx)
	isHeadsUp := ctx.PlayersInHand <= 2

	if toCall == 0 {
		switch {
		case handValue >= 0.8:
			betSize := ctx.Pot * 3 / 4
			return Decision{Action: Raise, Amount: betSize}
		case handValue >= 0.5:
			betSize := ctx.Pot * 2 / 3
			if s.rng.Float64() < 0.8 {
				return Decision{Action: Raise, Amount: betSize}
			}
			return Decision{Action: Check}
		case handValue >= 0.3:
			if s.rng.Float64() < 0.6 || isHeadsUp {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
			return Decision{Action: Check}
		default:
			if isHeadsUp && s.rng.Float64() < 0.35 {
				return Decision{Action: Raise, Amount: ctx.Pot / 3}
			}
			return Decision{Action: Check}
		}
	}

	potOdds := ctx.PotOdds()
	if handValue > potOdds {
		if handValue >= 0.7 && s.rng.Float64() < 0.5 {
			return Decision{Action: Raise, Amount: toCall*2 + ctx.Pot/2}
		}
		return Decision{Action: Call}
	}

	if handValue > potOdds*0.7 && s.rng.Float64() < 0.3 {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
}

func (s *GTOStrategy) assessHandValue(result HandResult, ctx GameContext) float64 {
	switch result.Rank {
	case RoyalFlush, StraightFlush:
		return 1.0
	case FourOfAKind:
		return 0.95
	case FullHouse:
		return 0.85
	case Flush:
		return 0.75
	case Straight:
		return 0.70
	case ThreeOfAKind:
		return 0.60
	case TwoPair:
		return 0.50
	case OnePair:
		if len(result.HighCards) > 0 && result.HighCards[0] >= Ten {
			return 0.35
		}
		return 0.25
	default:
		if len(result.HighCards) > 0 && result.HighCards[0] >= Queen {
			return 0.15
		}
		return 0.05
	}
}

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

func (s *FishStrategy) Decide(ctx GameContext) Decision {
	if ctx.Street == Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *FishStrategy) decidePreflop(ctx GameContext) Decision {
	strength := ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if strength <= StrongHand {
			if s.rng.Float64() < 0.3 {
				return Decision{Action: Raise, Amount: ctx.BigBlind * 2}
			}
			return Decision{Action: Raise, Amount: ctx.BigBlind * 3}
		}
		if s.rng.Float64() < 0.4 {
			return Decision{Action: Raise, Amount: ctx.BigBlind * 2}
		}
		return Decision{Action: Check}
	}

	if strength <= PlayableHand {
		if toCall > ctx.Pot && s.rng.Float64() < 0.2 {
			return Decision{Action: Fold}
		}
		return Decision{Action: Call}
	}

	if s.rng.Float64() < s.callFreq {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
}

func (s *FishStrategy) decidePostflop(ctx GameContext) Decision {
	allCards := append(ctx.Hand[:], ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()

	hasPair := result.Rank >= OnePair
	hasStrongHand := result.Rank >= TwoPair

	if toCall == 0 {
		if hasStrongHand {
			if s.rng.Float64() < 0.4 {
				return Decision{Action: Raise, Amount: ctx.BigBlind}
			}
			return Decision{Action: Raise, Amount: ctx.Pot / 2}
		}
		if hasPair {
			if s.rng.Float64() < 0.5 {
				return Decision{Action: Raise, Amount: ctx.BigBlind}
			}
			return Decision{Action: Check}
		}
		if s.rng.Float64() < s.bluffFreq {
			return Decision{Action: Raise, Amount: ctx.BigBlind}
		}
		return Decision{Action: Check}
	}

	if hasStrongHand {
		return Decision{Action: Call}
	}

	if hasPair {
		if toCall <= ctx.Pot/2 || s.rng.Float64() < 0.6 {
			return Decision{Action: Call}
		}
		return Decision{Action: Fold}
	}

	if s.chaseDraws && s.hasDrawingHand(ctx) {
		return Decision{Action: Call}
	}

	if s.rng.Float64() < s.callFreq*0.4 {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
}

func (s *FishStrategy) hasDrawingHand(ctx GameContext) bool {
	allCards := append(ctx.Hand[:], ctx.Community...)

	suitCounts := make(map[Suit]int)
	for _, c := range allCards {
		suitCounts[c.Suit]++
	}
	for _, count := range suitCounts {
		if count >= 4 {
			return true
		}
	}

	ranks := make([]int, len(allCards))
	for i, c := range allCards {
		ranks[i] = int(c.Rank)
	}
	for i := 0; i < len(ranks); i++ {
		consecutive := 1
		for j := i + 1; j < len(ranks); j++ {
			diff := ranks[j] - ranks[i]
			if diff < 0 {
				diff = -diff
			}
			if diff <= 4 {
				consecutive++
			}
		}
		if consecutive >= 4 {
			return true
		}
	}

	return false
}

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

func (s *TAGStrategy) Decide(ctx GameContext) Decision {
	if ctx.Street == Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *TAGStrategy) decidePreflop(ctx GameContext) Decision {
	strength := ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	playThreshold := MarginalHand
	if s.tightness > 0.7 {
		playThreshold = PlayableHand
	}
	if s.tightness > 0.85 {
		playThreshold = StrongHand
	}

	if toCall == 0 {
		if strength <= playThreshold {
			raiseAmount := ctx.BigBlind * 3
			if ctx.Position == LatePosition {
				raiseAmount = ctx.BigBlind*2 + ctx.BigBlind/2
			}
			return Decision{Action: Raise, Amount: raiseAmount}
		}
		if ctx.Position == LatePosition && strength == MarginalHand {
			if s.rng.Float64() < 0.4 {
				return Decision{Action: Raise, Amount: ctx.BigBlind*2 + ctx.BigBlind/2}
			}
		}
		return Decision{Action: Check}
	}

	switch strength {
	case PremiumHand:
		if s.rng.Float64() < s.aggression {
			return Decision{Action: Raise, Amount: toCall * 3}
		}
		return Decision{Action: Call}
	case StrongHand:
		if ctx.RaisesThisRound == 0 && s.rng.Float64() < s.aggression*0.5 {
			return Decision{Action: Raise, Amount: toCall * 3}
		}
		return Decision{Action: Call}
	case PlayableHand:
		if ctx.Position >= MiddlePosition && ctx.PotOdds() < 0.25 {
			return Decision{Action: Call}
		}
		return Decision{Action: Fold}
	default:
		return Decision{Action: Fold}
	}
}

func (s *TAGStrategy) decidePostflop(ctx GameContext) Decision {
	allCards := append(ctx.Hand[:], ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if result.Rank >= TwoPair {
			betSize := int(float64(ctx.Pot) * (0.5 + s.aggression*0.3))
			return Decision{Action: Raise, Amount: betSize}
		}
		if result.Rank >= OnePair {
			if s.rng.Float64() < s.aggression*0.7 {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
		}
		return Decision{Action: Check}
	}

	if result.Rank >= ThreeOfAKind {
		if s.rng.Float64() < s.aggression {
			return Decision{Action: Raise, Amount: toCall * 2}
		}
		return Decision{Action: Call}
	}

	if result.Rank >= OnePair {
		if ctx.PotOdds() < 0.35 {
			return Decision{Action: Call}
		}
	}

	return Decision{Action: Fold}
}

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

func (s *LAGStrategy) Decide(ctx GameContext) Decision {
	if ctx.Street == Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *LAGStrategy) decidePreflop(ctx GameContext) Decision {
	strength := ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if strength <= MarginalHand || s.rng.Float64() < 0.4 {
			return Decision{Action: Raise, Amount: ctx.BigBlind * 3}
		}
		return Decision{Action: Check}
	}

	if strength <= PlayableHand {
		if s.rng.Float64() < s.aggression {
			return Decision{Action: Raise, Amount: toCall * 3}
		}
		return Decision{Action: Call}
	}

	if strength == MarginalHand && ctx.Position >= MiddlePosition {
		return Decision{Action: Call}
	}

	if s.rng.Float64() < 0.2 {
		return Decision{Action: Raise, Amount: toCall * 3}
	}

	return Decision{Action: Fold}
}

func (s *LAGStrategy) decidePostflop(ctx GameContext) Decision {
	allCards := append(ctx.Hand[:], ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if result.Rank >= OnePair || s.rng.Float64() < s.aggression*0.6 {
			betSize := ctx.Pot/2 + ctx.Pot/4
			return Decision{Action: Raise, Amount: betSize}
		}
		return Decision{Action: Check}
	}

	if result.Rank >= TwoPair {
		if s.rng.Float64() < s.aggression {
			return Decision{Action: Raise, Amount: toCall * 3}
		}
		return Decision{Action: Call}
	}

	if result.Rank >= OnePair {
		return Decision{Action: Call}
	}

	if s.rng.Float64() < s.aggression*0.3 {
		return Decision{Action: Raise, Amount: toCall * 2}
	}

	if ctx.PotOdds() < 0.25 {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
}

type NitStrategy struct {
	rng *rand.Rand
}

func NewNitStrategy() *NitStrategy {
	return &NitStrategy{rng: rand.New(rand.NewSource(rand.Int63()))}
}

func (s *NitStrategy) Name() string {
	return "Nit"
}

func (s *NitStrategy) Decide(ctx GameContext) Decision {
	if ctx.Street == Preflop {
		return s.decidePreflop(ctx)
	}
	return s.decidePostflop(ctx)
}

func (s *NitStrategy) decidePreflop(ctx GameContext) Decision {
	strength := ClassifyPreflopHand(ctx.Hand)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if strength <= StrongHand {
			return Decision{Action: Raise, Amount: ctx.BigBlind * 4}
		}
		return Decision{Action: Check}
	}

	if strength == PremiumHand {
		return Decision{Action: Raise, Amount: toCall * 3}
	}
	if strength == StrongHand && ctx.PotOdds() < 0.2 {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
}

func (s *NitStrategy) decidePostflop(ctx GameContext) Decision {
	allCards := append(ctx.Hand[:], ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if result.Rank >= ThreeOfAKind {
			return Decision{Action: Raise, Amount: ctx.Pot * 2 / 3}
		}
		return Decision{Action: Check}
	}

	if result.Rank >= TwoPair {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
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
		Hand:          player.Hand,
		Community:     table.Community,
		Street:        table.Street,
		Pot:           table.Pot,
		CurrentBet:    table.CurrentBet,
		PlayerBet:     player.Bet,
		PlayerChips:   player.Chips,
		Position:      GetPosition(playerIndex, table.DealerPos, len(table.Players)),
		PlayersInHand: table.ActivePlayers,
		BigBlind:      table.BigBlind,
	}
}
