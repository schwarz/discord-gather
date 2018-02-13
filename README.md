# Discord Gather

Discord Gather is a gather/mix bot for Discord servers built with minimal abstraction, suited for simple use cases.

## Design

Users interact with the bot in two ways:

1. Sending messages to a channel such as `!add` or `!remove` 
1. Updating their presence, such as going offline

By sending `!add` in a bot's declared channel, a user is added to a map. Once `GATHER_NEEDED` users have done so, two random, equal-sized teams are printed out using mentions to alert the users. The map is cleared and the process can start again.

Should a user go offline while added, the bot will automatically remove and mention the user.

## Getting Started

Assuming you have [installed Go](https://golang.org/doc/install), either clone the repository and build the binary yourself or run:

    $ go install github.com/schwarz/discord-gather

Which will create an executable `$GOPATH/bin/discord-gather` ready for use.

The bot is configured using environment variables. For an overview of all the used variables please see `.env.example`.

_Important:_ The `GATHER_DISCORD_TOKEN` environment variable must be prefixed with `Bot ` for bot users.

## License

Discord Gather is released under the [MIT License](https://opensource.org/licenses/MIT).
