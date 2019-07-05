package main

const qryDummy = "dummy"
const qryCreateNewPoll = "createNewPoll"
const qryCreatePoll = "createpoll"
const qryPollDone = "polldone"

const qryEditPayload = 'e'
const qryPrevPoll = "-"
const qryNextPoll = "+"
const qryToggleActive = "c"
const qryToggleMultipleChoice = "m"
const qryEditQuestion = "q"
const qryAddOptions = "o"

const (
	ohHi = iota
	waitingForQuestion
	waitingForOption
	pollDone
	editPoll
	editQuestion
	addOption
)

const (
	open = iota
	inactive
)

const (
	standard = iota
	multipleChoice
)

const (
	displayVotePercent = iota
	hideDisplayVotePercent
)

var maxNumberOfUsersListed = 100
var maxPollsInlineQuery = 5

const lineSep = "\u257C\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u257E"
