package pokerlib

func ParseCard(s string) (Card, bool) {
	if len(s) < 2 || len(s) > 3 {
		return Card{}, false
	}

	rankStr := s[:len(s)-1]
	suitChar := s[len(s)-1]

	var rank Rank
	switch rankStr {
	case "2":
		rank = Two
	case "3":
		rank = Three
	case "4":
		rank = Four
	case "5":
		rank = Five
	case "6":
		rank = Six
	case "7":
		rank = Seven
	case "8":
		rank = Eight
	case "9":
		rank = Nine
	case "10", "T":
		rank = Ten
	case "J":
		rank = Jack
	case "Q":
		rank = Queen
	case "K":
		rank = King
	case "A":
		rank = Ace
	default:
		return Card{}, false
	}

	var suit Suit
	switch suitChar {
	case 'h', 'H':
		suit = Hearts
	case 'd', 'D':
		suit = Diamonds
	case 'c', 'C':
		suit = Clubs
	case 's', 'S':
		suit = Spades
	default:
		return Card{}, false
	}

	return Card{Suit: suit, Rank: rank}, true
}

func ParseCards(cards ...string) ([]Card, bool) {
	result := make([]Card, len(cards))
	for i, s := range cards {
		card, ok := ParseCard(s)
		if !ok {
			return nil, false
		}
		result[i] = card
	}
	return result, true
}

func ParseHand(card1, card2 string) ([2]Card, bool) {
	c1, ok := ParseCard(card1)
	if !ok {
		return [2]Card{}, false
	}
	c2, ok := ParseCard(card2)
	if !ok {
		return [2]Card{}, false
	}
	return [2]Card{c1, c2}, true
}

func CardToShortString(c Card) string {
	var rankStr string
	switch c.Rank {
	case Ten:
		rankStr = "T"
	default:
		rankStr = c.Rank.String()
	}

	var suitChar byte
	switch c.Suit {
	case Hearts:
		suitChar = 'h'
	case Diamonds:
		suitChar = 'd'
	case Clubs:
		suitChar = 'c'
	case Spades:
		suitChar = 's'
	}

	return rankStr + string(suitChar)
}

func IsPocketPair(hand [2]Card) bool {
	return hand[0].Rank == hand[1].Rank
}

func IsSuited(hand [2]Card) bool {
	return hand[0].Suit == hand[1].Suit
}

func IsConnector(hand [2]Card) bool {
	diff := int(hand[0].Rank) - int(hand[1].Rank)
	if diff < 0 {
		diff = -diff
	}
	return diff == 1 || (hand[0].Rank == Ace && hand[1].Rank == Two) || (hand[0].Rank == Two && hand[1].Rank == Ace)
}

func HandCategory(hand [2]Card) string {
	high, low := hand[0].Rank, hand[1].Rank
	if low > high {
		high, low = low, high
	}

	suffix := ""
	if IsPocketPair(hand) {
		return high.String() + high.String()
	}
	if IsSuited(hand) {
		suffix = "s"
	} else {
		suffix = "o"
	}
	return high.String() + low.String() + suffix
}
