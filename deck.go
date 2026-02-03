package pokerlib

import (
	"math/rand"
)

type Deck struct {
	cards []Card
}

func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for _, suit := range AllSuits() {
		for _, rank := range AllRanks() {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	return &Deck{cards: cards}
}

func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
}

func (d *Deck) Deal() Card {
	if len(d.cards) == 0 {
		panic("deck is empty")
	}
	card := d.cards[0]
	d.cards = d.cards[1:]
	return card
}

func (d *Deck) DealN(n int) []Card {
	if len(d.cards) < n {
		panic("not enough cards in deck")
	}
	cards := make([]Card, n)
	copy(cards, d.cards[:n])
	d.cards = d.cards[n:]
	return cards
}

func (d *Deck) Remaining() int {
	return len(d.cards)
}

func (d *Deck) Reset() {
	*d = *NewDeck()
}

func (d *Deck) Cards() []Card {
	return d.cards
}

func (d *Deck) RemoveCard(card Card) bool {
	for i, c := range d.cards {
		if c.Suit == card.Suit && c.Rank == card.Rank {
			d.cards = append(d.cards[:i], d.cards[i+1:]...)
			return true
		}
	}
	return false
}
