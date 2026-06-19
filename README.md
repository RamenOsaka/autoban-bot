# RamenBot

## Hosting

#### Compatible platforms
RamenBot has to be self hosted, but can run on virtually any server and Operating System. There is no released executable of the bot yet, so you will need to install [golang](https://go.dev/doc/install) to be able to run it or build the executable yourself.

#### Material requirements
RamenBot is lightweight and fast, if you only intend to use it on your own servers, any specs will do. Keep in mind however that no testing has been done for large scale use, anything above a few hundreds of servers has no guarantee to work well on low end servers.

## Installation

1. Clone this repository to your local directory, and enter the folder 
```bash
git clone https://github.com/RamenOsaka/autoban-bot & cd autoban-bot
```
2. Set up the environnement variables `DISCORD_TOKEN` and `APP_ID`. If you do not know how to set up environnement variables, you can create a `.env` file in the root folder of the bot and add one environnement variable by line with the synthax `VARIABLE=VALUE`.
    1. `DISCORD_TOKEN`is the identification token of your discord bot, it should be a 3 part alphanumerical string.
    2. `APP_ID` is a number associated to your application ID, found under **General Information** in the developer portal.
3. Run the application in the terminal :
```bash
go run .
``` 
The application will create `config.json` upon starting for the first time, this file stores every server configuration. If everything goes well, you will see `RamenBot is now running.  Press CTRL-C to exit.`. You simply have to invite the bot to your server to use it.

## Functionnalities
* Banning members if they select a role from a customizable *banned role* list. 

## Commands
* `/ping` pings the bot, use it if you are unsure whether the bot is online or not.
* `/listbannedroles` lists every role which will trigger a ban upon being picked.
* `/addrole` adds a role to the banned role list.
* `/removerole` removes a role to the banned role list.
* `/setlogchannel` sets the channel where the logs for bans, mutes, etc. will be outputed.
