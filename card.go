package pokerlib

import "fmt"

type Suit int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
)

func (s Suit) String() string {
	return [...]string{"Hearts", "Diamonds", "Clubs", "Spades"}[s]
}

type Rank int

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

func (r Rank) String() string {
	names := map[Rank]string{
		Two: "2", Three: "3", Four: "4", Five: "5",
		Six: "6", Seven: "7", Eight: "8", Nine: "9",
		Ten: "10", Jack: "J", Queen: "Q", King: "K", Ace: "A",
	}
	return names[r]
}

type Card struct {
	Suit Suit
	Rank Rank
}

func (c Card) String() string {
	return fmt.Sprintf("%s of %s", c.Rank, c.Suit)
}

func AllSuits() []Suit {
	return []Suit{Hearts, Diamonds, Clubs, Spades}
}

func AllRanks() []Rank {
	return []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}
}
