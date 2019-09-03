package main

const (
	locGotQuestion           = "OK now that we got a question, please send answer options to your poll."
	locStartCommand          = "/start"
	locEditCommand           = "/edit"
	locCreateNewPoll         = "create new poll"
	locInlineInsertPoll      = "insert poll into chat"
	locSharePoll             = "share poll"
	locNewQuestion           = "Great! Send a question for the new poll, please."
	locEditQuestion          = "Okay, just send the new question, please."
	locEditQuestionButton    = "change question"
	locAddOption             = "Alright, send options that you want to add to the poll, please."
	locAddOptionButton       = "add options"
	locDeletePollButton      = "delete poll"
	locGotEditQuestion       = "Thanks, the question was changed to \"%s\"."
	locNoMessageToEdit       = "Sorry, I could not find a poll to edit."
	locCurrentlySelectedPoll = "Currently selected poll:\n"
	locMainMenu              = "I can help you create, send and manage polls.\n\nWhat do you want to do?"
	locAboutCommand          = "/about"
	locAboutMessage          = "You can find me on github:\nhttps://github.com/semog/probebot"
	locPollDoneButton        = "done"
	locToggleInactive        = "open poll"
	locToggleOpen            = "close poll"
	locToggleSingleChoice    = "set multiple choice"
	locToggleMultipleChoice  = "set single choice"
	locToggleShowVotePct     = "show vote pct"
	locToggleHideVotePct     = "hide vote pct"
	locAddedOption           = "You can add more options by sending messages each containing one option. If you are done, please push the 'done' button.\n\nPreview:\n"
	locPollIsInactive        = "This poll is inactive."
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
