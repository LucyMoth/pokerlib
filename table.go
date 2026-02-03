package pokerlib

type Street int

const (
	Preflop Street = iota
	Flop
	Turn
	River
	Showdown
)

func (s Street) String() string {
	return [...]string{"Preflop", "Flop", "Turn", "River", "Showdown"}[s]
}

type Table struct {
	Players       []*Player
	Community     []Card
	Pot           int
	Deck          *Deck
	Street        Street
	DealerPos     int
	CurrentBet    int
	SmallBlind    int
	BigBlind      int
	ActivePlayers int
}

func NewTable(smallBlind, bigBlind int) *Table {
	return &Table{
		Players:    make([]*Player, 0),
		Community:  make([]Card, 0, 5),
		SmallBlind: smallBlind,
		BigBlind:   bigBlind,
	}
}

func (t *Table) AddPlayer(p *Player) {
	t.Players = append(t.Players, p)
}

func (t *Table) RemovePlayer(name string) {
	for i, p := range t.Players {
		if p.Name == name {
			t.Players = append(t.Players[:i], t.Players[i+1:]...)
			return
		}
	}
}

func (t *Table) StartHand() {
	t.Deck = NewDeck()
	t.Deck.Shuffle()
	t.Community = t.Community[:0]
	t.Pot = 0
	t.Street = Preflop
	t.CurrentBet = 0
	t.ActivePlayers = 0

	for _, p := range t.Players {
		p.ResetForNewHand()
		if p.Status == Active {
			t.ActivePlayers++
		}
	}

	t.DealHoleCards()
	t.PostBlinds()
}

func (t *Table) DealHoleCards() {
	for _, p := range t.Players {
		if p.Status == Active || p.Status == AllIn {
			cards := t.Deck.DealN(2)
			p.ReceiveCards(cards)
		}
	}
}

func (t *Table) PostBlinds() {
	sbPos := (t.DealerPos + 1) % len(t.Players)
	bbPos := (t.DealerPos + 2) % len(t.Players)

	sbAmount := t.Players[sbPos].PlaceBet(t.SmallBlind)
	bbAmount := t.Players[bbPos].PlaceBet(t.BigBlind)

	t.Pot += sbAmount + bbAmount
	t.CurrentBet = t.BigBlind
}

func (t *Table) DealFlop() {
	t.Deck.Deal()
	t.Community = append(t.Community, t.Deck.DealN(3)...)
	t.Street = Flop
	t.ResetBetsForStreet()
}

func (t *Table) DealTurn() {
	t.Deck.Deal()
	t.Community = append(t.Community, t.Deck.Deal())
	t.Street = Turn
	t.ResetBetsForStreet()
}

func (t *Table) DealRiver() {
	t.Deck.Deal()
	t.Community = append(t.Community, t.Deck.Deal())
	t.Street = River
	t.ResetBetsForStreet()
}

func (t *Table) ResetBetsForStreet() {
	t.CurrentBet = 0
	for _, p := range t.Players {
		p.Bet = 0
	}
}

func (t *Table) CollectBets() {
	for _, p := range t.Players {
		t.Pot += p.Bet
		p.Bet = 0
	}
}

func (t *Table) ProcessAction(player *Player, action Action, amount int) bool {
	switch action {
	case Fold:
		player.Fold()
		t.ActivePlayers--
		return true

	case Check:
		if t.CurrentBet > player.Bet {
			return false
		}
		return true

	case Call:
		callAmount := t.CurrentBet - player.Bet
		bet := player.PlaceBet(callAmount)
		t.Pot += bet
		return true

	case Raise:
		if amount < t.CurrentBet*2 && player.Chips > amount {
			return false
		}
		raiseAmount := amount - player.Bet
		bet := player.PlaceBet(raiseAmount)
		t.Pot += bet
		t.CurrentBet = player.Bet
		return true

	case AllInAction:
		bet := player.PlaceBet(player.Chips)
		t.Pot += bet
		if player.Bet > t.CurrentBet {
			t.CurrentBet = player.Bet
		}
		return true
	}
	return false
}

func (t *Table) DetermineWinners() []*Player {
	var activePlayers []*Player
	for _, p := range t.Players {
		if p.Status == Active || p.Status == AllIn {
			activePlayers = append(activePlayers, p)
		}
	}

	if len(activePlayers) == 1 {
		return activePlayers
	}

	var bestHand HandResult
	var winners []*Player
	first := true

	for _, p := range activePlayers {
		allCards := p.AllCards(t.Community)
		hand := EvaluateHand(allCards)

		if first {
			bestHand = hand
			winners = []*Player{p}
			first = false
		} else {
			cmp := hand.Compare(bestHand)
			if cmp > 0 {
				bestHand = hand
				winners = []*Player{p}
			} else if cmp == 0 {
				winners = append(winners, p)
			}
		}
	}

	return winners
}

func (t *Table) AwardPot() {
	winners := t.DetermineWinners()
	if len(winners) == 0 {
		return
	}

	share := t.Pot / len(winners)
	remainder := t.Pot % len(winners)

	for i, w := range winners {
		award := share
		if i < remainder {
			award++
		}
		w.Award(award)
	}

	t.Pot = 0
}

func (t *Table) NextDealer() {
	t.DealerPos = (t.DealerPos + 1) % len(t.Players)
}

func (t *Table) GetActivePlayers() []*Player {
	var active []*Player
	for _, p := range t.Players {
		if p.Status == Active {
			active = append(active, p)
		}
	}
	return active
}

func (t *Table) IsHandComplete() bool {
	return t.ActivePlayers <= 1 || t.Street == Showdown
}

func (t *Table) AdvanceStreet() {
	switch t.Street {
	case Preflop:
		t.DealFlop()
	case Flop:
		t.DealTurn()
	case Turn:
		t.DealRiver()
	case River:
		t.Street = Showdown
	}
}
