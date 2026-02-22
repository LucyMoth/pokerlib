package main

import (
	"github.com/LucyMoth/pokerlib"
)

type Session struct {
	Table        *pokerlib.Table
	Players      []*PlayerConfig
	SmallBlind   int
	BigBlind     int
	InitialChips map[string]int
	HandsPlayed  int
	HandHistory  []HandRecord
	PlayerStats  map[string]*PlayerStats
}

type PlayerConfig struct {
	Name         string
	InitialChips int
	Strategy     pokerlib.Strategy
	Player       *pokerlib.Player
}

type PlayerStats struct {
	Name          string
	Strategy      string
	HandsPlayed   int
	HandsWon      int
	HandsTied     int
	TotalWinnings int
	TotalLosses   int
	BiggestWin    int
	BiggestLoss   int
	FinalChips    int
	InitialChips  int
	ShowdownsWon  int
	ShowdownsSeen int
	FoldsPreflop  int
	FoldsPostflop int
}

type HandRecord struct {
	HandNumber  int
	Winners     []string
	WinningHand string
	PotSize     int
	Community   []pokerlib.Card
	PlayerHands map[string][2]pokerlib.Card
	PlayerChips map[string]int
	ChipChanges map[string]int
	Actions     []ActionRecord
}

type ActionRecord struct {
	Street pokerlib.Street
	Player string
	Action pokerlib.Action
	Amount int
}

func NewSession() *Session {
	return &Session{
		SmallBlind:   25,
		BigBlind:     50,
		InitialChips: make(map[string]int),
		HandHistory:  make([]HandRecord, 0),
		PlayerStats:  make(map[string]*PlayerStats),
	}
}

func (s *Session) Reset() {
	s.Table = nil
	s.Players = nil
	s.InitialChips = make(map[string]int)
	s.HandsPlayed = 0
	s.HandHistory = make([]HandRecord, 0)
	s.PlayerStats = make(map[string]*PlayerStats)
}

func (s *Session) ResetGame() {
	s.Table = nil
	s.HandsPlayed = 0
	s.HandHistory = make([]HandRecord, 0)

	for _, pc := range s.Players {
		pc.Player = nil
		s.InitialChips[pc.Name] = pc.InitialChips
		s.PlayerStats[pc.Name] = &PlayerStats{
			Name:         pc.Name,
			Strategy:     pc.Strategy.Name(),
			InitialChips: pc.InitialChips,
			FinalChips:   pc.InitialChips,
		}
	}
}

func (s *Session) RemovePlayer(name string) bool {
	for i, pc := range s.Players {
		if pc.Name == name {
			s.Players = append(s.Players[:i], s.Players[i+1:]...)
			delete(s.InitialChips, name)
			delete(s.PlayerStats, name)
			return true
		}
	}
	return false
}

func (s *Session) ModifyPlayer(name string, chips int, strategy pokerlib.Strategy) bool {
	for _, pc := range s.Players {
		if pc.Name == name {
			if chips > 0 {
				pc.InitialChips = chips
				s.InitialChips[name] = chips
			}
			if strategy != nil {
				pc.Strategy = strategy
			}
			s.PlayerStats[name] = &PlayerStats{
				Name:         name,
				Strategy:     pc.Strategy.Name(),
				InitialChips: pc.InitialChips,
				FinalChips:   pc.InitialChips,
			}
			return true
		}
	}
	return false
}

func (s *Session) GetPlayer(name string) *PlayerConfig {
	for _, pc := range s.Players {
		if pc.Name == name {
			return pc
		}
	}
	return nil
}

func (s *Session) SetBlinds(small, big int) {
	s.SmallBlind = small
	s.BigBlind = big
	if s.Table != nil {
		s.Table.SmallBlind = small
		s.Table.BigBlind = big
	}
}

func (s *Session) AddPlayer(name string, chips int, strategy pokerlib.Strategy) {
	if s.Players == nil {
		s.Players = make([]*PlayerConfig, 0)
	}

	config := &PlayerConfig{
		Name:         name,
		InitialChips: chips,
		Strategy:     strategy,
	}
	s.Players = append(s.Players, config)
	s.InitialChips[name] = chips

	s.PlayerStats[name] = &PlayerStats{
		Name:         name,
		Strategy:     strategy.Name(),
		InitialChips: chips,
		FinalChips:   chips,
	}
}

func (s *Session) InitializeTable() {
	s.Table = pokerlib.NewTable(s.SmallBlind, s.BigBlind)

	for _, pc := range s.Players {
		player := pokerlib.NewAIPlayer(pc.Name, pc.InitialChips, pc.Strategy)
		pc.Player = player
		s.Table.AddPlayer(player)
	}
}

func (s *Session) CanRun() bool {
	return len(s.Players) >= 2
}

func (s *Session) RunSimulation(numHands int, onProgress func(current, total int)) []HandRecord {
	if s.Table == nil {
		s.InitializeTable()
	}

	records := make([]HandRecord, 0, numHands)

	for i := 0; i < numHands; i++ {
		if onProgress != nil {
			onProgress(i+1, numHands)
		}

		record := s.playOneHand()
		records = append(records, record)
		s.HandHistory = append(s.HandHistory, record)
		s.HandsPlayed++

		if s.countActivePlayers() < 2 {
			break
		}
	}

	s.updateFinalStats()
	return records
}

func (s *Session) playOneHand() HandRecord {
	record := HandRecord{
		HandNumber:  s.HandsPlayed + 1,
		PlayerHands: make(map[string][2]pokerlib.Card),
		PlayerChips: make(map[string]int),
		ChipChanges: make(map[string]int),
		Actions:     make([]ActionRecord, 0),
	}

	startingChips := make(map[string]int)
	for _, p := range s.Table.Players {
		startingChips[p.Name] = p.Chips
	}

	s.Table.StartHand()

	for _, p := range s.Table.Players {
		if p.Status == pokerlib.Active || p.Status == pokerlib.AllIn {
			record.PlayerHands[p.Name] = p.Hand
		}
	}

	for s.Table.Street != pokerlib.Showdown && s.Table.ActivePlayers > 1 {
		if s.countCanActPlayers() > 1 {
			s.playBettingRound(&record)
		}

		if s.Table.ActivePlayers > 1 {
			s.Table.AdvanceStreet()
		}
	}

	record.Community = make([]pokerlib.Card, len(s.Table.Community))
	copy(record.Community, s.Table.Community)
	record.PotSize = s.Table.Pot

	winners := s.Table.DetermineWinners()
	for _, w := range winners {
		record.Winners = append(record.Winners, w.Name)
		s.PlayerStats[w.Name].HandsWon++
		if len(winners) > 1 {
			s.PlayerStats[w.Name].HandsTied++
		}
	}

	if len(winners) > 0 && len(s.Table.Community) == 5 {
		allCards := winners[0].AllCards(s.Table.Community)
		result := pokerlib.EvaluateHand(allCards)
		record.WinningHand = result.Rank.String()

		for _, w := range winners {
			s.PlayerStats[w.Name].ShowdownsWon++
		}
		for _, p := range s.Table.Players {
			if p.Status == pokerlib.Active || p.Status == pokerlib.AllIn {
				s.PlayerStats[p.Name].ShowdownsSeen++
			}
		}
	}

	s.Table.AwardPot()

	for _, p := range s.Table.Players {
		record.PlayerChips[p.Name] = p.Chips
		change := p.Chips - startingChips[p.Name]
		record.ChipChanges[p.Name] = change

		stats := s.PlayerStats[p.Name]
		stats.HandsPlayed++
		if change > 0 {
			stats.TotalWinnings += change
			if change > stats.BiggestWin {
				stats.BiggestWin = change
			}
		} else if change < 0 {
			stats.TotalLosses += (-change)
			if (-change) > stats.BiggestLoss {
				stats.BiggestLoss = -change
			}
		}
	}

	s.Table.NextDealer()

	return record
}

func (s *Session) countCanActPlayers() int {
	count := 0
	for _, p := range s.Table.Players {
		if p.Status == pokerlib.Active {
			count++
		}
	}
	return count
}

func (s *Session) playBettingRound(record *HandRecord) {
	if s.Table.ActivePlayers <= 1 {
		return
	}

	numPlayers := len(s.Table.Players)
	startIdx := (s.Table.DealerPos + 1) % numPlayers
	if s.Table.Street == pokerlib.Preflop {
		startIdx = (s.Table.DealerPos + 3) % numPlayers
	}

	lastAggressor := -1
	actionsThisRound := 0
	maxActions := numPlayers * 4

	currentIdx := startIdx
	canAct := s.countCanActPlayers()
	playersToAct := canAct

	for playersToAct > 0 && actionsThisRound < maxActions {
		player := s.Table.Players[currentIdx]

		if player.Status == pokerlib.Active {
			canAct = s.countCanActPlayers()
			if canAct <= 1 {
				return
			}

			needsToAct := player.Bet < s.Table.CurrentBet || actionsThisRound < canAct

			if lastAggressor == currentIdx {
				break
			}

			if needsToAct || lastAggressor == -1 {
				ctx := pokerlib.BuildGameContext(player, s.Table, currentIdx)
				decision := player.MakeDecision(ctx)

				record.Actions = append(record.Actions, ActionRecord{
					Street: s.Table.Street,
					Player: player.Name,
					Action: decision.Action,
					Amount: decision.Amount,
				})

				prevBet := s.Table.CurrentBet
				s.Table.ProcessAction(player, decision.Action, decision.Amount)
				actionsThisRound++

				canAct = s.countCanActPlayers()
				isRaise := (decision.Action == pokerlib.Raise || decision.Action == pokerlib.AllInAction) && s.Table.CurrentBet > prevBet
				if isRaise {
					lastAggressor = currentIdx
					playersToAct = canAct
					if playersToAct > 1 {
						playersToAct--
					}
				} else {
					playersToAct--
				}

				if decision.Action == pokerlib.Fold {
					stats := s.PlayerStats[player.Name]
					if s.Table.Street == pokerlib.Preflop {
						stats.FoldsPreflop++
					} else {
						stats.FoldsPostflop++
					}
				}

				if canAct <= 1 {
					return
				}
			}
		}

		currentIdx = (currentIdx + 1) % numPlayers

		if currentIdx == startIdx && lastAggressor == -1 {
			break
		}
	}
}

func (s *Session) countActivePlayers() int {
	count := 0
	for _, p := range s.Table.Players {
		if p.Chips > 0 {
			count++
		}
	}
	return count
}

func (s *Session) updateFinalStats() {
	for _, p := range s.Table.Players {
		if stats, ok := s.PlayerStats[p.Name]; ok {
			stats.FinalChips = p.Chips
		}
	}
}

func (s *Session) GetPlayerStats() []*PlayerStats {
	stats := make([]*PlayerStats, 0, len(s.PlayerStats))
	for _, pc := range s.Players {
		if st, ok := s.PlayerStats[pc.Name]; ok {
			stats = append(stats, st)
		}
	}
	return stats
}
