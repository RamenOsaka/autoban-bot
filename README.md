# RamenBot

## Hosting

#### Compatible platforms
RamenBot has to be self hosted, but can run on virtually any server and Operating System. There is no released executable of the bot yet, so you will need to install [golang](https://go.dev/doc/install) to be able to build the executable yourself.

#### Specs requirements
RamenBot is lightweight and fast, if you only intend to use it on your own servers, any specs will do. Keep in mind however that no testing has been done for large scale use, anything above a few hundreds of servers has no guarantee to work well on low end servers.

## Installation (for Debian based-distros)

1. Install Go (Download the package directly from Go.dev instead of using `apt` as some older repositories won't download a recent enough version of Go) :
```bash
sudo apt remove golang-go -y # in case go is already installed
wget https://go.dev/dl/go1.26.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc # Adding Go to the environnement PATH
source ~/.bashrc
go version # should output something like "go version go1.26.4 linux/amd64"
```

2. Clone this repository to your local directory, and enter the folder :
```bash
cd ~
git clone https://github.com/RamenOsaka/autoban-bot autoban-bot
cd autoban-bot
```

3. Install dependancies and compile the app :
```bash
go mod download
go build -o autoban-bot # should create an executable called "autoban-bot"
```

4. Set up the environnement variables `DISCORD_TOKEN` and `APP_ID`.

You can create a `.env` file in the root folder of the bot and add one environnement variable by line with the synthax `VARIABLE=VALUE` if you do not want to add global environnenement variables.
    1. `DISCORD_TOKEN`is the identification token of your discord bot, it should be a 3 part alphanumerical string.
    2. `APP_ID` is a number associated to your application ID, found under **General Information** in the developer portal.

6. Lauching the application

If everything goes well, you can simply run the executable and the bot will be online and running. 
```bash
autoban-bot
``` 
The application will create `config.json` if no file is found, this file stores every server configuration. If everything goes well, you will see `RamenBot is now running.  Press CTRL-C to exit.`.

## Functionnalities
* Banning members if they select a role from a customizable *banned role* list. 

## Commands
* `/ping` pings the bot, use it if you are unsure whether the bot is online or not.
* `/listbannedroles` lists every role which will trigger a ban upon being picked.
* `/addrole` adds a role to the banned role list.
* `/removerole` removes a role to the banned role list.
* `/setlogchannel` sets the channel where the logs for bans, mutes, etc. will be outputed.
