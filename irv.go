package main

import "math"

// Get Instant Runoff Voting option ID
// Based on https://courses.lumenlearning.com/waymakermath4libarts/chapter/instant-runoff-voting/
func getIRVOptionIDs(p *poll) []int {
	// scan the p.Answers to perform elimination.
	// Answers from user are ranked by LastSaved
	// return calculate voting by elimination and roll-up.
	// resolve ties by LastSaved. Oldest takes priority.

	numOptions := len(p.Options)
	// Options that have been eliminated
	eliminatedOptions := make([]int, 0)
	var voteCount map[int]int
	for {
		// Weighted count of votes for each option
		voteCount = make(map[int]int)
		voteMatrix := make(map[int][]int)
		// Build the evaluation matrix
		for _, a := range p.Answers {
			if !intrg_contains(eliminatedOptions, a.OptionID) {
				voteMatrix[a.UserID] = append(voteMatrix[a.UserID], a.OptionID)
			}
		}

		// Tally the weighted counts. More weight to first place choices.
		for _, v := range voteMatrix {
			for i, m := range v {
				voteCount[m] += numOptions - i
			}
		}

		if len(voteCount) == 1 || allValuesEqual(voteCount) {
			// We have a winner or a tie
			break
		}

		// Remove the least vote getter.
		lowVote := math.MaxInt32
		lowVoteID := 0
		for i, c := range voteCount {
			if lowVote > c {
				lowVote = c
				lowVoteID = i
			}
		}
		eliminatedOptions = append(eliminatedOptions, lowVoteID)
	}

	ids := make([]int, 0)
	for id := range voteCount {
		ids = append(ids, id)
	}
	return ids
}

func allValuesEqual(counts map[int]int) bool {
	lastVal := -1
	for _, vc := range counts {
		if lastVal == -1 {
			lastVal = vc
			continue
		}
		if lastVal != vc {
			return false
		}
	}

	return true
}

// scenario of ties:
// Option 1 - one vote
// Option 2 - no votes
// Option 3 - one vote
