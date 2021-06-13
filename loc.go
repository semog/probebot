package main

const (
	locGotQuestion                  = "Now, send answer options to the poll."
	locStartCommand                 = "/start"
	locEditCommand                  = "/edit"
	locCreateNewPoll                = "create new poll"
	locInlineInsertPoll             = "insert poll into chat"
	locSharePoll                    = "share poll"
	locNewQuestion                  = "Let's get started!\nSend the question for the new poll."
	locEditQuestion                 = "Send the new question."
	locEditQuestionButton           = "change question"
	locEditOption                   = "Alright, send the options for the poll."
	locEditOptionsButton            = "edit options"
	locDeletePollButton             = "delete poll"
	locResetPollButton              = "reset poll"
	locResetPollMessage             = "Poll responses were cleared"
	locClosePollBeforeDeleteMessage = "Close poll before deleting"
	locGotEditQuestion              = "The question was changed to \"%s\"."
	locNoMessageToEdit              = "Sorry, I could not find a poll to edit."
	locCurrentlySelectedPoll        = "Currently selected poll (%s):\n"
	locMainMenu                     = "I can help you create, send and manage polls.\n\nWhat would you like to do?"
	locAboutCommand                 = "/about"
	locAboutMessage                 = "You can find me on github:\nhttps://github.com/semog/probebot"
	locPollDoneButton               = "done"
	locToggleInactive               = "open poll"
	locToggleOpen                   = "close poll"
	locToggleSingleChoice           = "set multiple choice"
	locToggleMultipleChoice         = "set ranked voting"
	locToggleRankedVoting           = "set single choice"
	locToggleShowVotePct            = "show vote pct"
	locToggleHideVotePct            = "hide vote pct"
	locAddedOption                  = "You can add more options by sending a new message.\nDelete options by sending just the number (i.e., '1.' to delete option one). Edit options by prefixing them with the option number (i.e., '2. New option two').\nIf you are done, please push the 'done' button.\n\nPreview:\n"
	locPollIsInactive               = "This poll is closed"
	locPollDeletedMessage           = "Poll deleted: %s"
	locErrUpdatingPollMessage       = "Error updating poll"
	locErrDeletingPollMessage       = "Error deleting poll"
	locSelectionRemoved             = "Selection removed"
	locYouSelected                  = `You selected "%s"`
)

/*
Following is the command menu for constructing the bot with @BotFather.
Use the /setcommands command and reply with the following list of commands.
---------------------
start - Start the bot.
edit - Edit poll questions.
about - About this bot.
*/

/*
Description text:
-------------------------
This bot creates inline polls in Telegram chats without spamming multiple messages.
*/
