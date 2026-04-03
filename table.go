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
	LastRaise     int
	SmallBlind    int
	BigBlind      int
	ActivePlayers int
	RaisesThisStreet int
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
	numPlayers := len(t.Players)

	sbPos := (t.DealerPos + 1) % numPlayers
	for t.Players[sbPos].Status != Active && t.Players[sbPos].Status != AllIn {
		sbPos = (sbPos + 1) % numPlayers
		if sbPos == t.DealerPos {
			return
		}
	}

	bbPos := (sbPos + 1) % numPlayers
	for t.Players[bbPos].Status != Active && t.Players[bbPos].Status != AllIn {
		bbPos = (bbPos + 1) % numPlayers
		if bbPos == sbPos {
			return
		}
	}

	sbAmount := t.Players[sbPos].PlaceBet(t.SmallBlind)
	bbAmount := t.Players[bbPos].PlaceBet(t.BigBlind)

	t.Pot += sbAmount + bbAmount
	t.CurrentBet = t.BigBlind
	t.LastRaise = t.BigBlind
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
	t.LastRaise = t.BigBlind
	t.RaisesThisStreet = 0
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
		minRaise := t.CurrentBet + t.LastRaise
		if minRaise < t.BigBlind*2 {
			minRaise = t.BigBlind * 2
		}
		if amount < minRaise && player.Chips+player.Bet > amount {
			amount = minRaise
		}
		raiseAmount := amount - player.Bet
		if raiseAmount <= 0 {
			return false
		}
		prevBet := t.CurrentBet
		bet := player.PlaceBet(raiseAmount)
		t.Pot += bet
		if player.Bet > prevBet {
			t.LastRaise = player.Bet - prevBet
		}
		t.CurrentBet = player.Bet
		t.RaisesThisStreet++
		return true

	case AllInAction:
		bet := player.PlaceBet(player.Chips)
		t.Pot += bet
		if player.Bet > t.CurrentBet {
			t.LastRaise = player.Bet - t.CurrentBet
			t.CurrentBet = player.Bet
			t.RaisesThisStreet++
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
	// Build side pots based on all-in amounts
	type sidePot struct {
		amount   int
		eligible []*Player
	}

	var contenders []*Player
	for _, p := range t.Players {
		if p.Status == Active || p.Status == AllIn {
			contenders = append(contenders, p)
		}
	}

	if len(contenders) == 0 {
		t.Pot = 0
		return
	}

	if len(contenders) == 1 {
		contenders[0].Award(t.Pot)
		t.Pot = 0
		return
	}

	// Collect all distinct bet levels from all-in players
	betLevels := make(map[int]bool)
	for _, p := range t.Players {
		if p.TotalBet > 0 {
			betLevels[p.TotalBet] = true
		}
	}
	sorted := make([]int, 0, len(betLevels))
	for b := range betLevels {
		sorted = append(sorted, b)
	}
	sortInts(sorted)

	var pots []sidePot
	prevLevel := 0
	for _, level := range sorted {
		pot := sidePot{}
		perPlayer := level - prevLevel
		if perPlayer <= 0 {
			continue
		}
		for _, p := range t.Players {
			contribution := p.TotalBet - prevLevel
			if contribution > perPlayer {
				contribution = perPlayer
			}
			if contribution > 0 {
				pot.amount += contribution
			}
			if (p.Status == Active || p.Status == AllIn) && p.TotalBet >= level {
				pot.eligible = append(pot.eligible, p)
			}
		}
		if pot.amount > 0 {
			pots = append(pots, pot)
		}
		prevLevel = level
	}

	// Award each side pot
	for _, pot := range pots {
		if len(pot.eligible) == 0 {
			continue
		}

		var bestHand HandResult
		var winners []*Player
		first := true

		for _, p := range pot.eligible {
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

		share := pot.amount / len(winners)
		remainder := pot.amount % len(winners)

		for i, w := range winners {
			award := share
			if i < remainder {
				award++
			}
			w.Award(award)
		}
	}

	t.Pot = 0
}

func sortInts(a []int) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

func (t *Table) NextDealer() {
	numPlayers := len(t.Players)
	start := t.DealerPos
	t.DealerPos = (t.DealerPos + 1) % numPlayers
	for t.Players[t.DealerPos].Chips == 0 {
		t.DealerPos = (t.DealerPos + 1) % numPlayers
		if t.DealerPos == start {
			break
		}
	}
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
