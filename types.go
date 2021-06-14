package main

import (
	"fmt"
)

type answer struct {
	ID        int
	PollID    int
	UserID    int
	OptionID  int
	LastSaved int
}

type option struct {
	ID     int
	PollID int
	Text   string
}

type poll struct {
	ID             int
	MessageID      int
	UserID         int
	Question       string
	Inactive       int
	Type           int
	DisplayPercent int
	Options        []option
	Answers        []answer
}

func (poll *poll) fmtQuery(query string) string {
	return fmt.Sprintf("%c:%d:%s", qryEditPayload, poll.ID, query)
}

func (poll *poll) isInactive() bool {
	return poll.Inactive == inactive
}

func (poll *poll) isSingleChoice() bool {
	return poll.Type == standard
}

func (poll *poll) isMultipleChoice() bool {
	return poll.Type == multipleChoice
}

func (poll *poll) isRankedVoting() bool {
	return poll.Type == rankedVoting
}

func (poll *poll) isDisplayVotePercent() bool {
	return poll.DisplayPercent == displayVotePercent
}
