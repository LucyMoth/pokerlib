package pokerlib

import (
	"math/rand"
	"sort"
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
	allCards := make([]Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := AnalyzeDraws(ctx.Hand, ctx.Community)

	handValue := s.assessHandValue(result, ctx, draws)
	isHeadsUp := ctx.PlayersInHand <= 2

	if toCall == 0 {
		switch {
		case handValue >= 0.85:
			// Slow-play some monsters
			if s.rng.Float64() < 0.2 {
				return Decision{Action: Check}
			}
			betSize := ctx.Pot * 3 / 4
			return Decision{Action: Raise, Amount: betSize}
		case handValue >= 0.6:
			betSize := ctx.Pot * 2 / 3
			if s.rng.Float64() < 0.85 {
				return Decision{Action: Raise, Amount: betSize}
			}
			return Decision{Action: Check}
		case handValue >= 0.35:
			freq := 0.55
			if isHeadsUp {
				freq = 0.7
			}
			if s.rng.Float64() < freq {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
			return Decision{Action: Check}
		case draws.HasDraw():
			// Semi-bluff with draws
			freq := 0.45
			if draws.OpenEndedStraight || draws.FlushDraw {
				freq = 0.55
			}
			if isHeadsUp {
				freq += 0.15
			}
			if s.rng.Float64() < freq {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
			return Decision{Action: Check}
		default:
			// Pure bluff frequency
			bluffFreq := 0.15
			if isHeadsUp {
				bluffFreq = 0.30
			}
			if ctx.Position == LatePosition {
				bluffFreq += 0.10
			}
			if s.rng.Float64() < bluffFreq {
				return Decision{Action: Raise, Amount: ctx.Pot / 3}
			}
			return Decision{Action: Check}
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
			return Decision{Action: Raise, Amount: toCall*2 + ctx.Pot/2}
		}
		if handValue >= 0.5 && s.rng.Float64() < 0.25 {
			return Decision{Action: Raise, Amount: toCall * 2}
		}
		return Decision{Action: Call}
	}

	// Semi-bluff raise with strong draws
	if draws.HasDraw() && draws.Outs >= 8 && s.rng.Float64() < 0.3 {
		return Decision{Action: Raise, Amount: toCall * 2}
	}

	if effectiveValue > potOdds*0.7 && s.rng.Float64() < 0.25 {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
}

func (s *GTOStrategy) assessHandValue(result HandResult, ctx GameContext, draws DrawInfo) float64 {
	baseValue := 0.0
	switch result.Rank {
	case RoyalFlush, StraightFlush:
		return 1.0
	case FourOfAKind:
		return 0.97
	case FullHouse:
		baseValue = 0.90
	case Flush:
		baseValue = 0.80
		if len(result.HighCards) > 0 && result.HighCards[0] >= Queen {
			baseValue = 0.85
		}
	case Straight:
		baseValue = 0.72
	case ThreeOfAKind:
		baseValue = 0.62
		if s.isSetOnBoard(result, ctx) {
			baseValue = 0.55
		}
	case TwoPair:
		baseValue = 0.48
		if s.isTopTwoPair(result, ctx) {
			baseValue = 0.55
		}
	case OnePair:
		baseValue = s.assessPairValue(result, ctx)
	default:
		if len(result.HighCards) > 0 && result.HighCards[0] >= Queen {
			baseValue = 0.12
		} else {
			baseValue = 0.05
		}
	}

	// Adjust for board texture
	if len(ctx.Community) >= 3 {
		paired := s.boardPaired(ctx.Community)
		monotone := s.boardMonotone(ctx.Community)
		connected := s.boardConnected(ctx.Community)

		if paired && result.Rank <= OnePair {
			baseValue *= 0.85
		}
		if monotone && result.Rank < Flush {
			baseValue *= 0.80
		}
		if connected && result.Rank < Straight {
			baseValue *= 0.90
		}
	}

	// Adjust for number of opponents
	if ctx.PlayersInHand > 2 && result.Rank <= OnePair {
		baseValue *= 0.85
	}
	if ctx.PlayersInHand > 3 && result.Rank <= TwoPair {
		baseValue *= 0.90
	}

	return baseValue
}

func (s *GTOStrategy) assessPairValue(result HandResult, ctx GameContext) float64 {
	if len(result.HighCards) == 0 || len(ctx.Community) == 0 {
		return 0.25
	}
	pairRank := result.HighCards[0]

	// Check if it's an overpair (both hole cards form pair above board)
	maxBoard := Rank(0)
	for _, c := range ctx.Community {
		if c.Rank > maxBoard {
			maxBoard = c.Rank
		}
	}

	if pairRank > maxBoard {
		// Overpair
		if pairRank >= Queen {
			return 0.45
		}
		return 0.38
	}

	// Top pair
	if pairRank == maxBoard {
		if len(result.HighCards) > 1 && result.HighCards[1] >= Ten {
			return 0.35
		}
		return 0.28
	}

	// Middle/bottom pair
	return 0.18
}

func (s *GTOStrategy) isSetOnBoard(result HandResult, ctx GameContext) bool {
	if result.Rank != ThreeOfAKind || len(result.HighCards) == 0 {
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

func (s *GTOStrategy) isTopTwoPair(result HandResult, ctx GameContext) bool {
	if result.Rank != TwoPair || len(result.HighCards) < 2 {
		return false
	}
	maxBoard := Rank(0)
	for _, c := range ctx.Community {
		if c.Rank > maxBoard {
			maxBoard = c.Rank
		}
	}
	return result.HighCards[0] >= maxBoard
}

func (s *GTOStrategy) boardPaired(community []Card) bool {
	seen := make(map[Rank]bool)
	for _, c := range community {
		if seen[c.Rank] {
			return true
		}
		seen[c.Rank] = true
	}
	return false
}

func (s *GTOStrategy) boardMonotone(community []Card) bool {
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

func (s *GTOStrategy) boardConnected(community []Card) bool {
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
	allCards := make([]Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := AnalyzeDraws(ctx.Hand, ctx.Community)

	hasPair := result.Rank >= OnePair
	hasStrongHand := result.Rank >= TwoPair

	if toCall == 0 {
		if hasStrongHand {
			// Fish min-bet or half-pot randomly
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

	// Fish chase draws regardless of pot odds
	if s.chaseDraws && draws.HasDraw() {
		return Decision{Action: Call}
	}

	// Fish also chase backdoor draws and overcards
	if s.chaseDraws && s.hasOvercards(ctx) {
		if s.rng.Float64() < 0.5 {
			return Decision{Action: Call}
		}
	}

	if s.rng.Float64() < s.callFreq*0.4 {
		return Decision{Action: Call}
	}

	return Decision{Action: Fold}
}

func (s *FishStrategy) hasOvercards(ctx GameContext) bool {
	if len(ctx.Community) == 0 {
		return false
	}
	maxBoard := Rank(0)
	for _, c := range ctx.Community {
		if c.Rank > maxBoard {
			maxBoard = c.Rank
		}
	}
	return ctx.Hand[0].Rank > maxBoard || ctx.Hand[1].Rank > maxBoard
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
	allCards := make([]Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := AnalyzeDraws(ctx.Hand, ctx.Community)

	if toCall == 0 {
		if result.Rank >= TwoPair {
			betSize := int(float64(ctx.Pot) * (0.5 + s.aggression*0.3))
			return Decision{Action: Raise, Amount: betSize}
		}
		if result.Rank >= OnePair {
			if s.rng.Float64() < s.aggression*0.7 {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
			return Decision{Action: Check}
		}
		// Semi-bluff with draws
		if draws.HasDraw() && draws.Outs >= 8 {
			if s.rng.Float64() < s.aggression*0.6 {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
		}
		// C-bet bluff on dry boards
		if ctx.Street == Flop && ctx.PlayersInHand <= 3 {
			if s.rng.Float64() < s.aggression*0.4 {
				return Decision{Action: Raise, Amount: ctx.Pot / 3}
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
		return Decision{Action: Fold}
	}

	// Call with good draws if pot odds warrant it
	if draws.HasDraw() {
		equity := draws.DrawEquity()
		if equity > ctx.PotOdds()*0.8 {
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
	allCards := make([]Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()
	draws := AnalyzeDraws(ctx.Hand, ctx.Community)

	if toCall == 0 {
		// Bet strong hands
		if result.Rank >= TwoPair {
			betSize := ctx.Pot*2/3 + ctx.Pot/4
			return Decision{Action: Raise, Amount: betSize}
		}
		// Bet pairs aggressively
		if result.Rank >= OnePair {
			betSize := ctx.Pot/2 + ctx.Pot/4
			return Decision{Action: Raise, Amount: betSize}
		}
		// Semi-bluff with any draw
		if draws.HasDraw() {
			if s.rng.Float64() < s.aggression*0.7 {
				return Decision{Action: Raise, Amount: ctx.Pot * 2 / 3}
			}
			return Decision{Action: Check}
		}
		// Bluff frequently, especially in position
		bluffFreq := s.aggression * 0.5
		if ctx.Position == LatePosition {
			bluffFreq += 0.15
		}
		if ctx.PlayersInHand <= 2 {
			bluffFreq += 0.10
		}
		if s.rng.Float64() < bluffFreq {
			return Decision{Action: Raise, Amount: ctx.Pot / 2}
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
		// Float with pairs, occasionally raise
		if s.rng.Float64() < s.aggression*0.3 {
			return Decision{Action: Raise, Amount: toCall * 2}
		}
		return Decision{Action: Call}
	}

	// Call or raise with draws
	if draws.HasDraw() {
		if draws.Outs >= 8 && s.rng.Float64() < s.aggression*0.5 {
			return Decision{Action: Raise, Amount: toCall * 2}
		}
		return Decision{Action: Call}
	}

	// Float bluff
	if ctx.PlayersInHand <= 2 && s.rng.Float64() < s.aggression*0.25 {
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
	allCards := make([]Card, 0, 2+len(ctx.Community))
	allCards = append(allCards, ctx.Hand[:]...)
	allCards = append(allCards, ctx.Community...)
	result := EvaluateHand(allCards)
	toCall := ctx.ToCall()

	if toCall == 0 {
		if result.Rank >= ThreeOfAKind {
			return Decision{Action: Raise, Amount: ctx.Pot * 2 / 3}
		}
		if result.Rank >= TwoPair {
			if s.rng.Float64() < 0.6 {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
			return Decision{Action: Check}
		}
		// Bet overpairs occasionally
		if result.Rank == OnePair && len(result.HighCards) > 0 {
			maxBoard := Rank(0)
			for _, c := range ctx.Community {
				if c.Rank > maxBoard {
					maxBoard = c.Rank
				}
			}
			if result.HighCards[0] > maxBoard && s.rng.Float64() < 0.4 {
				return Decision{Action: Raise, Amount: ctx.Pot / 2}
			}
		}
		return Decision{Action: Check}
	}

	if result.Rank >= ThreeOfAKind {
		if s.rng.Float64() < 0.5 {
			return Decision{Action: Raise, Amount: toCall * 2}
		}
		return Decision{Action: Call}
	}

	if result.Rank >= TwoPair {
		return Decision{Action: Call}
	}

	if result.Rank == OnePair && len(result.HighCards) > 0 && result.HighCards[0] >= Queen {
		if ctx.PotOdds() < 0.25 {
			return Decision{Action: Call}
		}
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
