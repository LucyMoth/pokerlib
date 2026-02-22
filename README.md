# pokerlib

A Go library for poker hand evaluation and game simulation.

## Install

```bash
go get github.com/LucyMoth/pokerlib
```

## Usage

```go
import "github.com/LucyMoth/pokerlib"

deck := pokerlib.NewDeck()
deck.Shuffle()

hand := [2]pokerlib.Card{
    {Suit: pokerlib.Spades, Rank: pokerlib.Ace},
    {Suit: pokerlib.Spades, Rank: pokerlib.King},
}

result := pokerlib.SimulateHand(hand, nil, 1, 10000)
fmt.Printf("Equity: %.2f%%\n", result.Equity()*100)
```
