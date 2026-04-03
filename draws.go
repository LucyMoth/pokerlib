package pokerlib

import "sort"

type DrawInfo struct {
	FlushDraw         bool
	FlushDrawSuit     Suit
	FlushDrawCards    int
	OpenEndedStraight bool
	GutShotStraight   bool
	Outs              int
}

func AnalyzeDraws(hand [2]Card, community []Card) DrawInfo {
	if len(community) < 3 || len(community) > 4 {
		return DrawInfo{}
	}

	allCards := make([]Card, 0, 2+len(community))
	allCards = append(allCards, hand[:]...)
	allCards = append(allCards, community...)

	info := DrawInfo{}

	suitCounts := make(map[Suit]int)
	for _, c := range allCards {
		suitCounts[c.Suit]++
	}
	for suit, count := range suitCounts {
		if count == 4 {
			info.FlushDraw = true
			info.FlushDrawSuit = suit
			info.FlushDrawCards = 4
			info.Outs += 9
		}
	}

	uniqueRanks := make(map[int]bool)
	for _, c := range allCards {
		uniqueRanks[int(c.Rank)] = true
	}
	if uniqueRanks[int(Ace)] {
		uniqueRanks[1] = true
	}

	sorted := make([]int, 0, len(uniqueRanks))
	for r := range uniqueRanks {
		sorted = append(sorted, r)
	}
	sort.Ints(sorted)

	info.OpenEndedStraight, info.GutShotStraight = detectStraightDraw(sorted)
	if info.OpenEndedStraight {
		outs := 8
		if info.FlushDraw {
			outs = 6
		}
		info.Outs += outs
	} else if info.GutShotStraight {
		outs := 4
		if info.FlushDraw {
			outs = 3
		}
		info.Outs += outs
	}

	return info
}

func detectStraightDraw(sorted []int) (openEnded bool, gutshot bool) {
	n := len(sorted)
	for i := 0; i < n; i++ {
		for windowSize := 5; windowSize >= 4; windowSize-- {
			end := sorted[i] + windowSize - 1
			if end > int(Ace) && sorted[i] != 1 {
				continue
			}

			count := 0
			for _, r := range sorted {
				if r >= sorted[i] && r <= end {
					count++
				}
			}

			if count == 4 && windowSize == 5 {
				low := sorted[i]
				high := end
				allPresent := true
				missing := -1
				for v := low; v <= high; v++ {
					if !contains(sorted, v) {
						if missing == -1 {
							missing = v
						} else {
							allPresent = false
							break
						}
					}
				}
				if allPresent && missing != -1 {
					if missing == low || missing == high {
						openEnded = true
					} else {
						gutshot = true
					}
				}
			}
		}

		if sorted[i]+3 <= int(Ace) || sorted[i] == 1 {
			var window []int
			for _, r := range sorted {
				if r >= sorted[i] && r <= sorted[i]+4 {
					window = append(window, r)
				}
			}
			if len(window) == 4 {
				consec := isConsecutive(window)
				if consec {
					low := window[0]
					high := window[len(window)-1]
					if low > 1 && high < int(Ace) {
						openEnded = true
					} else if !openEnded {
						gutshot = true
					}
				}
			}
		}
	}
	return
}

func isConsecutive(sorted []int) bool {
	for i := 1; i < len(sorted); i++ {
		if sorted[i]-sorted[i-1] != 1 {
			return false
		}
	}
	return true
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func (d DrawInfo) DrawEquity() float64 {
	if d.Outs <= 0 {
		return 0
	}
	outs := d.Outs
	if outs > 20 {
		outs = 20
	}
	return float64(outs) * 0.02
}

func (d DrawInfo) HasDraw() bool {
	return d.FlushDraw || d.OpenEndedStraight || d.GutShotStraight
}
