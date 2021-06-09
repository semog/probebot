package main

// Command list
const (
	qryDummy         = "dummy"
	qryCreateNewPoll = "createNewPoll"
	qryCreatePoll    = "createPoll"
	qryPollDone      = "pollDone"
)

// Query command sub-operators
const (
	qryEditPayload          = 'e'
	qryPrevPoll             = "-"
	qryNextPoll             = "+"
	qryToggleActive         = "c"
	qryToggleMultipleChoice = "m"
	qryEditQuestion         = "q"
	qryAddOptions           = "o"
	qryToggleShowVotePct    = "v"
	qryDeletePoll           = "d"
	qryResetPoll            = "r"
)

// Poll editing states. Do not change the order of these constants.
// Their values are persisted to the database, and changing them could
// break the application.
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

const (
	maxNumberOfUsersListed = 100
	maxPollsInlineQuery    = 5
	lineSep                = "\u257C\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u2501\u257E"
)
