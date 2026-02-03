package pokerlib

type Action int

const (
	Fold Action = iota
	Check
	Call
	Raise
	AllInAction
)

func (a Action) String() string {
	return [...]string{"Fold", "Check", "Call", "Raise", "All-In"}[a]
}

type ActionResult struct {
	Player *Player
	Action Action
	Amount int
}
