package pokerlib

type PlayerStatus int

const (
	Active PlayerStatus = iota
	Folded
	AllIn
	Out
)

type Player struct {
	Name   string
	Hand   [2]Card
	Chips  int
	Bet    int
	Status PlayerStatus
}

func NewPlayer(name string, chips int) *Player {
	return &Player{
		Name:   name,
		Chips:  chips,
		Status: Active,
	}
}

func (p *Player) ReceiveCards(cards []Card) {
	if len(cards) >= 2 {
		p.Hand[0] = cards[0]
		p.Hand[1] = cards[1]
	}
}

func (p *Player) ClearHand() {
	p.Hand = [2]Card{}
}

func (p *Player) PlaceBet(amount int) int {
	if amount >= p.Chips {
		bet := p.Chips
		p.Bet += bet
		p.Chips = 0
		p.Status = AllIn
		return bet
	}
	p.Chips -= amount
	p.Bet += amount
	return amount
}

func (p *Player) Fold() {
	p.Status = Folded
}

func (p *Player) ResetForNewHand() {
	p.ClearHand()
	p.Bet = 0
	if p.Chips > 0 {
		p.Status = Active
	} else {
		p.Status = Out
	}
}

func (p *Player) Award(amount int) {
	p.Chips += amount
}

func (p *Player) AllCards(community []Card) []Card {
	cards := make([]Card, 0, 7)
	cards = append(cards, p.Hand[:]...)
	cards = append(cards, community...)
	return cards
}
