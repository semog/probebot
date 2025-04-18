package main

import (
	"fmt"
)

type answer struct {
	ID        int
	PollID    int
	UserID    int64
	OptionID  int
	LastSaved int64
}

type option struct {
	ID     int
	PollID int
	Text   string
}

type poll struct {
	ID             int
	MessageID      int
	UserID         int64
	Question       string
	Inactive       int
	Type           int
	DisplayPercent int
	CloseAt        int64
	CloseEvery     string
	ResetAt        int64
	ResetEvery     string
	OpenAt         int64
	OpenEvery      string
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
