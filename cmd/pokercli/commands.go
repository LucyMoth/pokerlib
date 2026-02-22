package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LucyMoth/pokerlib"
)

func (c *CLI) cmdHelp(args []string) {
	if len(args) > 0 {
		cmdName := strings.ToLower(args[0])
		if cmd, exists := c.commands[cmdName]; exists {
			fmt.Println()
			PrintHeader("Command: %s", cmd.Name)
			PrintLabel("Description", cmd.Description)
			PrintLabel("Usage", cmd.Usage)
			fmt.Println()
			return
		}
		PrintError("Unknown command: %s", args[0])
		return
	}

	fmt.Println()
	PrintHeader("Available Commands")
	PrintDivider()

	categories := map[string][]string{
		"Game Setup": {"newgame", "addplayer", "modplayer", "removeplayer", "randplayers", "players", "strategies"},
		"Simulation": {"run", "slowsim", "results", "details"},
		"Session":    {"status", "clear", "help", "exit"},
	}

	for category, cmds := range categories {
		fmt.Printf("\n%s%s%s%s\n", Bold, Yellow, category, Reset)
		for _, name := range cmds {
			if cmd, exists := c.commands[name]; exists {
				fmt.Printf("  %s%-12s%s %s%s%s\n", Cyan, cmd.Name, Reset, Dim, cmd.Description, Reset)
			}
		}
	}
	fmt.Println()
}

func (c *CLI) cmdNewGame(args []string) {
	small := 25
	big := 50

	if len(args) > 0 {
		parts := strings.Split(args[0], "/")
		if len(parts) == 2 {
			if s, err := strconv.Atoi(parts[0]); err == nil {
				small = s
			}
			if b, err := strconv.Atoi(parts[1]); err == nil {
				big = b
			}
		}
	}

	c.session.ResetGame()
	c.session.SetBlinds(small, big)

	fmt.Println()
	PrintSuccess("New game created")
	PrintLabel("Small Blind", small)
	PrintLabel("Big Blind", big)
	PrintLabel("Players", len(c.session.Players))
	fmt.Println()
	if len(c.session.Players) == 0 {
		PrintInfo("Add players with: addplayer <name> <chips> <strategy>")
	} else {
		PrintInfo("Players retained from previous session. Use 'clear' to remove all.")
	}
	fmt.Println()
}

func (c *CLI) cmdAddPlayer(args []string) {
	if len(args) < 3 {
		PrintError("Usage: addplayer <name> <chips> <strategy>")
		PrintInfo("Example: addplayer Alice 1000 gto")
		return
	}

	name := args[0]
	chips, err := strconv.Atoi(args[1])
	if err != nil || chips <= 0 {
		PrintError("Invalid chip amount: %s", args[1])
		return
	}

	strategyName := strings.ToLower(args[2])
	strategy := getStrategy(strategyName)
	if strategy == nil {
		PrintError("Unknown strategy: %s", args[2])
		PrintInfo("Use 'strategies' to see available options")
		return
	}

	c.session.AddPlayer(name, chips, strategy)

	fmt.Println()
	PrintSuccess("Player added: %s", name)
	PrintLabel("Chips", chips)
	PrintLabel("Strategy", ColorizeStrategy(strategy.Name()))
	fmt.Println()
}

func (c *CLI) cmdModPlayer(args []string) {
	if len(args) < 1 {
		PrintError("Usage: modplayer <name> [chips] [strategy]")
		PrintInfo("Example: modplayer Alice 2000 tag")
		return
	}

	name := args[0]
	player := c.session.GetPlayer(name)
	if player == nil {
		PrintError("Player not found: %s", name)
		return
	}

	chips := 0
	var strategy pokerlib.Strategy

	if len(args) >= 2 {
		if c, err := strconv.Atoi(args[1]); err == nil && c > 0 {
			chips = c
		} else {
			strategy = getStrategy(args[1])
		}
	}

	if len(args) >= 3 {
		strategy = getStrategy(args[2])
	}

	if chips == 0 && strategy == nil {
		PrintWarning("No changes specified")
		PrintInfo("Usage: modplayer <name> [chips] [strategy]")
		return
	}

	c.session.ModifyPlayer(name, chips, strategy)

	fmt.Println()
	PrintSuccess("Player modified: %s", name)
	updated := c.session.GetPlayer(name)
	PrintLabel("Chips", updated.InitialChips)
	PrintLabel("Strategy", ColorizeStrategy(updated.Strategy.Name()))
	fmt.Println()
}

func (c *CLI) cmdRemovePlayer(args []string) {
	if len(args) < 1 {
		PrintError("Usage: removeplayer <name>")
		return
	}

	name := args[0]
	if c.session.RemovePlayer(name) {
		PrintSuccess("Player removed: %s", name)
	} else {
		PrintError("Player not found: %s", name)
	}
}

func (c *CLI) cmdRandPlayers(args []string) {
	count := 4
	chips := 1000

	if len(args) >= 1 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 && n <= 9 {
			count = n
		}
	}
	if len(args) >= 2 {
		if c, err := strconv.Atoi(args[1]); err == nil && c > 0 {
			chips = c
		}
	}

	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry", "Ivy"}
	strategies := []string{"gto", "tag", "lag", "fish", "nit"}

	existing := make(map[string]bool)
	for _, p := range c.session.Players {
		existing[p.Name] = true
	}

	added := 0
	for _, name := range names {
		if added >= count {
			break
		}
		if existing[name] {
			continue
		}

		stratName := strategies[added%len(strategies)]
		strategy := getStrategy(stratName)
		c.session.AddPlayer(name, chips, strategy)
		added++
	}

	fmt.Println()
	PrintSuccess("Added %d random players", added)
	PrintLabel("Chips each", chips)
	fmt.Println()

	c.cmdPlayers(nil)
}

func (c *CLI) cmdPlayers(args []string) {
	if len(c.session.Players) == 0 {
		PrintWarning("No players added yet")
		PrintInfo("Use 'addplayer <name> <chips> <strategy>' to add players")
		return
	}

	fmt.Println()
	PrintHeader("Players (%d)", len(c.session.Players))
	PrintDivider()

	columns := []string{"#", "Name", "Chips", "Strategy"}
	widths := []int{3, 15, 10, 12}
	PrintTableHeader(columns, widths)

	for i, pc := range c.session.Players {
		chips := pc.InitialChips
		if pc.Player != nil {
			chips = pc.Player.Chips
		}
		values := []string{
			fmt.Sprintf("%d", i+1),
			pc.Name,
			fmt.Sprintf("%d", chips),
			pc.Strategy.Name(),
		}
		PrintTableRow(values, widths, false)
	}
	fmt.Println()
}

func (c *CLI) cmdStrategies(args []string) {
	fmt.Println()
	PrintHeader("Available Strategies")
	PrintDivider()

	strategies := []struct {
		name string
		desc string
	}{
		{"gto", "Game Theory Optimal - balanced, unexploitable play"},
		{"tag", "Tight-Aggressive - solid, selective, aggressive"},
		{"lag", "Loose-Aggressive - wide ranges, high aggression"},
		{"fish", "Calling station - passive, calls too much"},
		{"nit", "Ultra-tight - only premium hands"},
	}

	for _, s := range strategies {
		fmt.Printf("  %s%-8s%s %s%s%s\n",
			Bold+Cyan, s.name, Reset,
			Dim, s.desc, Reset)
	}
	fmt.Println()
}

func (c *CLI) cmdRun(args []string) {
	if !c.session.CanRun() {
		PrintError("Need at least 2 players to run simulation")
		PrintInfo("Use 'addplayer' to add more players")
		return
	}

	numHands := 100
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			numHands = n
		}
	}

	fmt.Println()
	PrintHeader("Running simulation: %d hands", numHands)
	fmt.Println()

	c.session.RunSimulation(numHands, func(current, total int) {
		PrintProgressBar(current, total, 40)
	})

	fmt.Println()
	PrintSuccess("Simulation complete!")
	fmt.Println()

	c.printResultsSummary()
}

func (c *CLI) cmdSlowSim(args []string) {
	if !c.session.CanRun() {
		PrintError("Need at least 2 players to run simulation")
		PrintInfo("Use 'addplayer' to add more players")
		return
	}

	numHands := 5
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			numHands = n
		}
	}

	delay := time.Second

	fmt.Println()
	PrintHeader("Slow Simulation: %d hands (1s delay)", numHands)
	fmt.Println()

	if c.session.Table == nil {
		c.session.InitializeTable()
	}

	for hand := 0; hand < numHands; hand++ {
		if c.session.countActivePlayers() < 2 {
			PrintWarning("Not enough players with chips to continue")
			break
		}

		c.playSlowHand(hand+1, delay)
		c.session.HandsPlayed++
		c.session.Table.NextDealer()

		if hand < numHands-1 {
			PrintDivider()
			time.Sleep(delay)
		}
	}

	c.session.updateFinalStats()
	fmt.Println()
	PrintSuccess("Slow simulation complete!")
	fmt.Println()
	c.printResultsSummary()
}

func (c *CLI) playSlowHand(handNum int, delay time.Duration) {
	table := c.session.Table

	startingChips := make(map[string]int)
	for _, p := range table.Players {
		startingChips[p.Name] = p.Chips
	}

	table.Deck = pokerlib.NewDeck()
	table.Deck.Shuffle()
	table.Community = table.Community[:0]
	table.Pot = 0
	table.Street = pokerlib.Preflop
	table.CurrentBet = 0
	table.ActivePlayers = 0

	for _, p := range table.Players {
		p.ResetForNewHand()
		if p.Status == pokerlib.Active {
			table.ActivePlayers++
		}
	}

	fmt.Printf("\n%s%s═══ HAND #%d ═══%s\n\n", Bold, Yellow, handNum, Reset)
	time.Sleep(delay)

	fmt.Printf("%s%sDealing hole cards...%s\n", Dim, Cyan, Reset)
	time.Sleep(delay / 2)

	for _, p := range table.Players {
		if p.Status == pokerlib.Active || p.Status == pokerlib.AllIn {
			cards := table.Deck.DealN(2)
			p.ReceiveCards(cards)
			card1 := ColorizeCard(pokerlib.CardToShortString(p.Hand[0]))
			card2 := ColorizeCard(pokerlib.CardToShortString(p.Hand[1]))
			stratName := "Human"
			if p.Strategy != nil {
				stratName = p.Strategy.Name()
			}
			fmt.Printf("  %s%-12s%s [%s] %s %s\n", Bold, p.Name, Reset, stratName, card1, card2)
		}
	}
	fmt.Println()
	time.Sleep(delay)

	sbPos := (table.DealerPos + 1) % len(table.Players)
	bbPos := (table.DealerPos + 2) % len(table.Players)
	sbAmount := table.Players[sbPos].PlaceBet(table.SmallBlind)
	bbAmount := table.Players[bbPos].PlaceBet(table.BigBlind)
	table.Pot += sbAmount + bbAmount
	table.CurrentBet = table.BigBlind

	fmt.Printf("%s%sBlinds posted:%s SB %s%s%s (%d), BB %s%s%s (%d)\n",
		Dim, White, Reset,
		Cyan, table.Players[sbPos].Name, Reset, sbAmount,
		Cyan, table.Players[bbPos].Name, Reset, bbAmount)
	fmt.Printf("%s%sPot: %d%s\n\n", Bold, Green, table.Pot, Reset)
	time.Sleep(delay)

	for table.Street != pokerlib.Showdown && table.ActivePlayers > 1 {
		c.playSlowBettingRound(delay)

		if table.ActivePlayers > 1 {
			prevStreet := table.Street
			table.AdvanceStreet()
			if table.Street == pokerlib.Flop && prevStreet == pokerlib.Preflop {
				c.showCommunityCards("FLOP", delay)
			} else if table.Street == pokerlib.Turn && prevStreet == pokerlib.Flop {
				c.showCommunityCards("TURN", delay)
			} else if table.Street == pokerlib.River && prevStreet == pokerlib.Turn {
				c.showCommunityCards("RIVER", delay)
			}
		}
	}

	winners := table.DetermineWinners()

	fmt.Printf("\n%s%s━━━ SHOWDOWN ━━━%s\n", Bold, Magenta, Reset)
	time.Sleep(delay / 2)

	if len(table.Community) == 5 {
		fmt.Printf("%sBoard:%s ", Bold, Reset)
		for _, card := range table.Community {
			fmt.Printf("%s ", ColorizeCard(pokerlib.CardToShortString(card)))
		}
		fmt.Println()
		time.Sleep(delay / 2)
	}

	fmt.Printf("%s%sPot: %d%s\n", Bold, Green, table.Pot, Reset)
	time.Sleep(delay / 2)

	if len(winners) > 0 {
		winnerNames := make([]string, len(winners))
		for i, w := range winners {
			winnerNames[i] = w.Name
			c.session.PlayerStats[w.Name].HandsWon++
		}

		if len(table.Community) == 5 {
			allCards := winners[0].AllCards(table.Community)
			result := pokerlib.EvaluateHand(allCards)
			fmt.Printf("%s%s🏆 Winner: %s with %s%s\n",
				Bold, BrightYellow,
				strings.Join(winnerNames, ", "),
				result.Rank.String(),
				Reset)

			for _, w := range winners {
				c.session.PlayerStats[w.Name].ShowdownsWon++
			}
			for _, p := range table.Players {
				if p.Status == pokerlib.Active || p.Status == pokerlib.AllIn {
					c.session.PlayerStats[p.Name].ShowdownsSeen++
				}
			}
		} else {
			fmt.Printf("%s%s🏆 Winner: %s (others folded)%s\n",
				Bold, BrightYellow,
				strings.Join(winnerNames, ", "),
				Reset)
		}
	}

	table.AwardPot()
	time.Sleep(delay / 2)

	fmt.Printf("\n%sChip changes:%s\n", Dim, Reset)
	for _, p := range table.Players {
		change := p.Chips - startingChips[p.Name]
		c.session.PlayerStats[p.Name].HandsPlayed++

		changeStr := ColorizeNumber(change, true)
		fmt.Printf("  %-12s %s → %s%d%s chips\n", p.Name, changeStr, Bold, p.Chips, Reset)

		stats := c.session.PlayerStats[p.Name]
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
	fmt.Println()
}

func (c *CLI) showCommunityCards(street string, delay time.Duration) {
	fmt.Printf("\n%s%s━━━ %s ━━━%s\n", Bold, Cyan, street, Reset)
	fmt.Printf("%sBoard:%s ", Bold, Reset)
	for _, card := range c.session.Table.Community {
		fmt.Printf("%s ", ColorizeCard(pokerlib.CardToShortString(card)))
	}
	fmt.Printf("%s%s(Pot: %d)%s\n\n", Dim, White, c.session.Table.Pot, Reset)
	time.Sleep(delay)
}

func (c *CLI) playSlowBettingRound(delay time.Duration) {
	table := c.session.Table
	if table.ActivePlayers <= 1 {
		return
	}

	numPlayers := len(table.Players)
	startIdx := (table.DealerPos + 1) % numPlayers
	if table.Street == pokerlib.Preflop {
		startIdx = (table.DealerPos + 3) % numPlayers
	}

	lastAggressor := -1
	actionsThisRound := 0
	maxActions := numPlayers * 4

	currentIdx := startIdx
	canAct := c.session.countCanActPlayers()
	playersToAct := canAct

	for playersToAct > 0 && actionsThisRound < maxActions {
		player := table.Players[currentIdx]

		if player.Status == pokerlib.Active {
			if table.ActivePlayers <= 1 {
				return
			}

			needsToAct := player.Bet < table.CurrentBet || actionsThisRound < table.ActivePlayers

			if lastAggressor == currentIdx {
				break
			}

			if needsToAct || lastAggressor == -1 {
				ctx := pokerlib.BuildGameContext(player, table, currentIdx)
				decision := player.MakeDecision(ctx)

				actionColor := White
				switch decision.Action {
				case pokerlib.Fold:
					actionColor = Red
				case pokerlib.Call, pokerlib.Check:
					actionColor = Yellow
				case pokerlib.Raise, pokerlib.AllInAction:
					actionColor = Green
				}

				amountStr := ""
				if decision.Amount > 0 {
					amountStr = fmt.Sprintf(" %d", decision.Amount)
				}
				fmt.Printf("  %s%-12s%s %s%s%s%s\n",
					Cyan, player.Name, Reset,
					Bold, actionColor, decision.Action.String(), Reset)
				if amountStr != "" {
					fmt.Printf("    %s→ %d chips%s\n", Dim, decision.Amount, Reset)
				}
				time.Sleep(delay)

				prevBet := table.CurrentBet
				table.ProcessAction(player, decision.Action, decision.Amount)
				actionsThisRound++

				canAct = c.session.countCanActPlayers()
				isRaise := (decision.Action == pokerlib.Raise || decision.Action == pokerlib.AllInAction) && table.CurrentBet > prevBet
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
					stats := c.session.PlayerStats[player.Name]
					if table.Street == pokerlib.Preflop {
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

func (c *CLI) cmdResults(args []string) {
	if c.session.HandsPlayed == 0 {
		PrintWarning("No hands played yet")
		PrintInfo("Use 'run <count>' to simulate hands")
		return
	}

	c.printResultsSummary()
}

func (c *CLI) printResultsSummary() {
	stats := c.session.GetPlayerStats()
	if len(stats) == 0 {
		return
	}

	fmt.Println()
	PrintHeader("Results Summary - %d Hands", c.session.HandsPlayed)
	PrintDivider()

	columns := []string{"Player", "Strategy", "Chips", "Net", "Win%", "VPIP", "SD Won"}
	widths := []int{12, 8, 10, 10, 8, 8, 8}
	PrintTableHeader(columns, widths)

	for _, s := range stats {
		net := s.FinalChips - s.InitialChips
		winRate := float64(0)
		if s.HandsPlayed > 0 {
			winRate = float64(s.HandsWon) / float64(s.HandsPlayed) * 100
		}
		vpip := float64(0)
		if s.HandsPlayed > 0 {
			vpip = float64(s.HandsPlayed-s.FoldsPreflop) / float64(s.HandsPlayed) * 100
		}
		sdWon := float64(0)
		if s.ShowdownsSeen > 0 {
			sdWon = float64(s.ShowdownsWon) / float64(s.ShowdownsSeen) * 100
		}

		netStr := fmt.Sprintf("%+d", net)
		netColor := White
		if net > 0 {
			netColor = Green
		} else if net < 0 {
			netColor = Red
		}

		values := []string{
			s.Name,
			s.Strategy,
			fmt.Sprintf("%d", s.FinalChips),
			netStr,
			fmt.Sprintf("%.1f%%", winRate),
			fmt.Sprintf("%.1f%%", vpip),
			fmt.Sprintf("%.1f%%", sdWon),
		}
		colors := []string{"", "", "", netColor, "", "", ""}
		PrintTableRowColored(values, widths, colors)
	}

	fmt.Println()
	PrintDivider()

	fmt.Printf("\n%sLegend:%s\n", Bold, Reset)
	fmt.Printf("  %sWin%%%s  - Percentage of hands won\n", Dim, Reset)
	fmt.Printf("  %sVPIP%s  - Voluntarily Put In Pot (not folded preflop)\n", Dim, Reset)
	fmt.Printf("  %sSD Won%s - Showdown win percentage\n", Dim, Reset)
	fmt.Println()
}

func (c *CLI) cmdDetails(args []string) {
	if len(c.session.HandHistory) == 0 {
		PrintWarning("No hand history available")
		PrintInfo("Use 'run <count>' to simulate hands first")
		return
	}

	if len(args) == 0 {
		c.printRecentHands(10)
		return
	}

	if strings.ToLower(args[0]) == "all" {
		c.printRecentHands(len(c.session.HandHistory))
		return
	}

	handNum, err := strconv.Atoi(args[0])
	if err != nil || handNum < 1 || handNum > len(c.session.HandHistory) {
		PrintError("Invalid hand number: %s", args[0])
		PrintInfo("Valid range: 1-%d", len(c.session.HandHistory))
		return
	}

	c.printHandDetail(c.session.HandHistory[handNum-1])
}

func (c *CLI) printRecentHands(count int) {
	start := len(c.session.HandHistory) - count
	if start < 0 {
		start = 0
	}

	fmt.Println()
	PrintHeader("Hand History (showing %d hands)", len(c.session.HandHistory)-start)
	PrintDivider()

	columns := []string{"#", "Winners", "Hand", "Pot"}
	widths := []int{5, 20, 18, 8}
	PrintTableHeader(columns, widths)

	for i := start; i < len(c.session.HandHistory); i++ {
		h := c.session.HandHistory[i]
		winners := strings.Join(h.Winners, ", ")
		if len(winners) > 18 {
			winners = winners[:15] + "..."
		}
		values := []string{
			fmt.Sprintf("%d", h.HandNumber),
			winners,
			h.WinningHand,
			fmt.Sprintf("%d", h.PotSize),
		}
		PrintTableRow(values, widths, false)
	}
	fmt.Println()
	PrintInfo("Use 'details <hand_number>' for detailed view")
	fmt.Println()
}

func (c *CLI) printHandDetail(h HandRecord) {
	fmt.Println()
	PrintHeader("Hand #%d Details", h.HandNumber)
	PrintDivider()

	fmt.Printf("\n%sCommunity Cards:%s ", Bold, Reset)
	for _, card := range h.Community {
		fmt.Printf("%s ", ColorizeCard(pokerlib.CardToShortString(card)))
	}
	fmt.Println()

	fmt.Printf("\n%sPlayer Hands:%s\n", Bold, Reset)
	for name, hand := range h.PlayerHands {
		if hand[0].Rank == 0 && hand[1].Rank == 0 {
			continue
		}
		card1 := ColorizeCard(pokerlib.CardToShortString(hand[0]))
		card2 := ColorizeCard(pokerlib.CardToShortString(hand[1]))
		change := h.ChipChanges[name]
		changeStr := ColorizeNumber(change, true)
		fmt.Printf("  %-12s: %s %s  (%s)\n", name, card1, card2, changeStr)
	}

	fmt.Printf("\n%sPot:%s %d\n", Bold, Reset, h.PotSize)
	fmt.Printf("%sWinner(s):%s %s\n", Bold, Reset, strings.Join(h.Winners, ", "))
	if h.WinningHand != "" {
		fmt.Printf("%sWinning Hand:%s %s\n", Bold, Reset, ColorizeHandRank(h.WinningHand))
	}

	if len(h.Actions) > 0 {
		fmt.Printf("\n%sAction Summary:%s\n", Bold, Reset)
		currentStreet := pokerlib.Street(-1)
		for _, a := range h.Actions {
			if a.Street != currentStreet {
				currentStreet = a.Street
				fmt.Printf("  %s%s%s\n", Dim+Yellow, currentStreet.String(), Reset)
			}
			amountStr := ""
			if a.Amount > 0 {
				amountStr = fmt.Sprintf(" %d", a.Amount)
			}
			fmt.Printf("    %-12s %s%s\n", a.Player, a.Action.String(), amountStr)
		}
	}
	fmt.Println()
}

func (c *CLI) cmdClear(args []string) {
	c.session.Reset()
	PrintSuccess("Session cleared")
	PrintInfo("Use 'newgame' to start a new game")
}

func (c *CLI) cmdStatus(args []string) {
	fmt.Println()
	PrintHeader("Current Session Status")
	PrintDivider()

	PrintLabel("Blinds", fmt.Sprintf("%d/%d", c.session.SmallBlind, c.session.BigBlind))
	PrintLabel("Players", len(c.session.Players))
	PrintLabel("Hands Played", c.session.HandsPlayed)

	if c.session.Table != nil {
		PrintLabel("Table Active", "Yes")
	} else {
		PrintLabel("Table Active", "No")
	}
	fmt.Println()

	if len(c.session.Players) > 0 {
		c.cmdPlayers(nil)
	}
}

func getStrategy(name string) pokerlib.Strategy {
	switch strings.ToLower(name) {
	case "gto":
		return pokerlib.NewGTOStrategy()
	case "tag":
		return pokerlib.NewTAGStrategy()
	case "lag":
		return pokerlib.NewLAGStrategy()
	case "fish":
		return pokerlib.NewFishStrategy()
	case "nit":
		return pokerlib.NewNitStrategy()
	default:
		return nil
	}
}
