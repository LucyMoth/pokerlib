# pokerlib

Texas Hold'em engine and AI simulation library for Go. No external dependencies.

## Install

```bash
go get github.com/LucyMoth/pokerlib
```

## Usage

### Equity simulation

```go
hand := [2]pokerlib.Card{
    {Suit: pokerlib.Spades, Rank: pokerlib.Ace},
    {Suit: pokerlib.Spades, Rank: pokerlib.King},
}

// equity vs 2 random opponents, 10k iterations
result := pokerlib.SimulateHand(hand, nil, 2, 10000)
fmt.Printf("Equity: %.1f%%\n", result.Equity()*100)
```

### Heads-up comparison

```go
hand1, _ := pokerlib.ParseHand("Ah", "Kh")
hand2, _ := pokerlib.ParseHand("Jc", "Jd")

r1, r2 := pokerlib.SimulateHeadsUp(hand1, hand2, nil, 10000)
fmt.Printf("AKs: %.1f%% vs JJ: %.1f%%\n", r1.Equity()*100, r2.Equity()*100)
```

### Running a game with AI players

```go
config := pokerlib.DefaultConfig() // 25/50 blinds, 1000 chips
game := pokerlib.NewGame(config)

game.AddPlayer("Alice")
game.AddPlayer("Bob")

// assign strategies
for _, p := range game.Table.Players {
    p.SetStrategy(pokerlib.NewGTOStrategy())
}

game.PlayHand()
winners := game.Table.DetermineWinners()
game.Table.AwardPot()
```

### Draw analysis

```go
hand, _ := pokerlib.ParseHand("Ah", "Kh")
community, _ := pokerlib.ParseCards("7h", "Th", "2c")

draws := pokerlib.AnalyzeDraws(hand, community)
fmt.Println(draws.FlushDraw)           // true
fmt.Println(draws.Outs)                // 9
fmt.Printf("Equity: %.0f%%\n", draws.DrawEquity()*100)
```

## Strategies

| Name | Style | Description |
|------|-------|-------------|
| GTO  | Balanced | Mixed frequencies, position-aware opens/3-bets, board texture reads, semi-bluffs |
| TAG  | Tight-Aggressive | Selective hand range, aggressive when in, c-bets draws |
| LAG  | Loose-Aggressive | Wide ranges, frequent bluffs, pressures opponents |
| Fish | Calling Station | Calls too much, chases draws regardless of odds |
| Nit  | Ultra-Tight | Only plays premium hands, folds everything else |

All strategies implement the `Strategy` interface:

```go
type Strategy interface {
    Decide(ctx GameContext) Decision
    Name() string
}
```

## CLI

There's an interactive CLI tool in `cmd/pokercli` for running simulations.

```bash
cd cmd/pokercli && go build && ./pokercli
```

Commands: `newgame`, `addplayer`, `randplayers`, `run`, `slowsim`, `results`, `details`, `players`, `strategies`, `help`

The `slowsim` command steps through hands showing hole cards, community cards, and actions with a delay — useful for watching strategy behavior.

## License

MIT — see [LICENSE](LICENSE) for the full (and somewhat unconventional) terms.
