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
	qryEditOptions          = "o"
	qryToggleShowVotePct    = "v"
	qryDeletePoll           = "d"
	qryResetPoll            = "r"
	qryReturnToPoll         = "b"
	qryAdvancedOpts         = "a"
	qrySetCloseAt           = "l"
	qryClearCloseAt         = "L"
	qrySetResetAt           = "t"
	qryClearResetAt         = "T"
	qrySetOpenAt            = "p"
	qryClearOpenAt          = "P"
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
	editAdvancedOptions
	waitingForCloseAt
	waitingForResetAt
	waitingForOpenAt
)

const (
	open = iota
	inactive
)

const (
	standard = iota
	multipleChoice
	rankedVoting
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

const (
	dateTimeFormat = "2006-01-02 3:04PM"
)
