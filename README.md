# probebot
This is a Telegram bot that helps by creating inline polls in Telegram chats
without spamming multiple messages.  This is based on [@pollrBot](https://github.com/jheuel/pollrBot).

The bot uses inline queries and feedback to inline queries, which have to be
enabled with the Telegram [@BotFather](https://telegram.me/BotFather).

## Usage
The bot can be installed with
```
go get github.com/semog/probebot
```
if you have a working Go environment.


After that you can run the bot with
```
DB=database.db APITOKEN=uiaeouiaouiao probebot
```
