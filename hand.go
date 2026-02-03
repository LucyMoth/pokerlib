package pokerlib

import "sort"

type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

func (h HandRank) String() string {
	return [...]string{
		"High Card", "One Pair", "Two Pair", "Three of a Kind",
		"Straight", "Flush", "Full House", "Four of a Kind",
		"Straight Flush", "Royal Flush",
	}[h]
}

type HandResult struct {
	Rank      HandRank
	HighCards []Rank
	Cards     []Card
}

func (h HandResult) Compare(other HandResult) int {
	if h.Rank != other.Rank {
		if h.Rank > other.Rank {
			return 1
		}
		return -1
	}
	for i := 0; i < len(h.HighCards) && i < len(other.HighCards); i++ {
		if h.HighCards[i] > other.HighCards[i] {
			return 1
		}
		if h.HighCards[i] < other.HighCards[i] {
			return -1
		}
	}
	return 0
}

func EvaluateHand(cards []Card) HandResult {
	if len(cards) < 5 {
		return HandResult{Rank: HighCard}
	}

	allCombinations := combinations(cards, 5)
	var best HandResult
	first := true

	for _, combo := range allCombinations {
		result := evaluateFiveCards(combo)
		if first || result.Compare(best) > 0 {
			best = result
			first = false
		}
	}

	return best
}

func evaluateFiveCards(cards []Card) HandResult {
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank > sorted[j].Rank
	})

	isFlush := checkFlush(sorted)
	isStraight, highCard := checkStraight(sorted)

	rankCounts := make(map[Rank]int)
	for _, c := range sorted {
		rankCounts[c.Rank]++
	}

	var quads, trips, pairs []Rank
	var singles []Rank

	for rank, count := range rankCounts {
		switch count {
		case 4:
			quads = append(quads, rank)
		case 3:
			trips = append(trips, rank)
		case 2:
			pairs = append(pairs, rank)
		case 1:
			singles = append(singles, rank)
		}
	}

	sort.Slice(quads, func(i, j int) bool { return quads[i] > quads[j] })
	sort.Slice(trips, func(i, j int) bool { return trips[i] > trips[j] })
	sort.Slice(pairs, func(i, j int) bool { return pairs[i] > pairs[j] })
	sort.Slice(singles, func(i, j int) bool { return singles[i] > singles[j] })

	if isFlush && isStraight {
		if highCard == Ace {
			return HandResult{Rank: RoyalFlush, HighCards: []Rank{Ace}, Cards: sorted}
		}
		return HandResult{Rank: StraightFlush, HighCards: []Rank{highCard}, Cards: sorted}
	}

	if len(quads) > 0 {
		kicker := singles[0]
		if len(pairs) > 0 && pairs[0] > kicker {
			kicker = pairs[0]
		}
		if len(trips) > 0 && trips[0] > kicker {
			kicker = trips[0]
		}
		return HandResult{Rank: FourOfAKind, HighCards: []Rank{quads[0], kicker}, Cards: sorted}
	}

	if len(trips) > 0 && (len(pairs) > 0 || len(trips) > 1) {
		pairRank := Rank(0)
		if len(pairs) > 0 {
			pairRank = pairs[0]
		}
		if len(trips) > 1 && trips[1] > pairRank {
			pairRank = trips[1]
		}
		return HandResult{Rank: FullHouse, HighCards: []Rank{trips[0], pairRank}, Cards: sorted}
	}

	if isFlush {
		highCards := make([]Rank, len(sorted))
		for i, c := range sorted {
			highCards[i] = c.Rank
		}
		return HandResult{Rank: Flush, HighCards: highCards, Cards: sorted}
	}

	if isStraight {
		return HandResult{Rank: Straight, HighCards: []Rank{highCard}, Cards: sorted}
	}

	if len(trips) > 0 {
		highCards := []Rank{trips[0]}
		highCards = append(highCards, singles[:2]...)
		return HandResult{Rank: ThreeOfAKind, HighCards: highCards, Cards: sorted}
	}

	if len(pairs) >= 2 {
		highCards := []Rank{pairs[0], pairs[1]}
		if len(singles) > 0 {
			highCards = append(highCards, singles[0])
		}
		return HandResult{Rank: TwoPair, HighCards: highCards, Cards: sorted}
	}

	if len(pairs) == 1 {
		highCards := []Rank{pairs[0]}
		highCards = append(highCards, singles[:3]...)
		return HandResult{Rank: OnePair, HighCards: highCards, Cards: sorted}
	}

	highCards := make([]Rank, len(sorted))
	for i, c := range sorted {
		highCards[i] = c.Rank
	}
	return HandResult{Rank: HighCard, HighCards: highCards, Cards: sorted}
}

func checkFlush(cards []Card) bool {
	suit := cards[0].Suit
	for _, c := range cards[1:] {
		if c.Suit != suit {
			return false
		}
	}
	return true
}

func checkStraight(cards []Card) (bool, Rank) {
	ranks := make([]int, len(cards))
	for i, c := range cards {
		ranks[i] = int(c.Rank)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ranks)))

	for i := 0; i < len(ranks)-1; i++ {
		if ranks[i]-ranks[i+1] != 1 {
			if ranks[0] == int(Ace) && ranks[1] == int(Five) {
				isWheelStraight := true
				expected := int(Five)
				for j := 1; j < len(ranks); j++ {
					if ranks[j] != expected {
						isWheelStraight = false
						break
					}
					expected--
				}
				if isWheelStraight {
					return true, Five
				}
			}
			return false, 0
		}
	}
	return true, Rank(ranks[0])
}

func combinations(cards []Card, k int) [][]Card {
	var result [][]Card
	n := len(cards)
	if k > n {
		return result
	}

	indices := make([]int, k)
	for i := range indices {
		indices[i] = i
	}

	for {
		combo := make([]Card, k)
		for i, idx := range indices {
			combo[i] = cards[idx]
		}
		result = append(result, combo)

		i := k - 1
		for i >= 0 && indices[i] == n-k+i {
			i--
		}
		if i < 0 {
			break
		}
		indices[i]++
		for j := i + 1; j < k; j++ {
			indices[j] = indices[j-1] + 1
		}
	}

	return result
}
