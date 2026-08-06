# RamenBot

## Hosting

#### Compatible platforms
RamenBot has to be self hosted, but can run on virtually any server and Operating System. There is no released executable of the bot yet, so you will need to install [golang](https://go.dev/doc/install) to be able to build the executable yourself.

#### Specs requirements
RamenBot is lightweight and fast, if you only intend to use it on your own servers, any specs will do. Keep in mind however that no testing has been done for large scale use, anything above a few hundreds of servers has no guarantee to work well on low end servers.

## Setup: Discord Developer Portal
 
Before installing the bot, you need to create a Discord application and configure it properly.
 
1. Go to [Discord Developer Portal](https://discord.com/developers/applications) and create a **New Application**
2. In the **bot** section under **Privileged Gateway Intents**, enable:
   - **Server Members Intent**
   - **Message Content Intent**
3. Copy your bot token (**Reset Token** → copy). You will need it later.
4. Go to *OAuth2 → URL Generator*:
   - Check **bot** and **applications.commands**
   - Under Bot Permissions, check: **Ban Members**, **Send Messages**, **View Channels**, **Embed Links**
   - Copy the generated URL, open it in your browser, and invite the bot to your server.
5. Your `Application ID` is found under **General Information** in the developer portal.

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
* `DISCORD_TOKEN`is the identification token of your discord bot, it should be a 3 part alphanumerical string.
* `APP_ID` is a number associated to your application ID, found under **General Information** in the developer portal.

5. Lauching the application

If everything goes well, you can simply run the executable and the bot will be online and running. 
```bash
autoban-bot
``` 
The application will create `config.json` if no file is found when the first config command is sent, this file stores every server configuration. If everything goes well, you will see `RamenBot is now running.  Press CTRL-C to exit.`.

## Running as a persistent service with systemd
 
To keep the bot running after closing your terminal and have it restart automatically on crash or reboot, set it up as a systemd service.
 
1. Create the service file
 
```bash
sudo nano /etc/systemd/system/autoban-bot.service
```
 
Paste the following content (adjust `WorkingDirectory`, `User` and `ExecStart` if you cloned the repo to a different path):
 
```ini
[Unit]
Description=AutoBan Discord Bot (Go)
After=network.target
 
[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/autoban-bot
ExecStart=/home/ubuntu/autoban-bot/autoban-bot
Restart=always
RestartSec=10
 
[Install]
WantedBy=multi-user.target
```
 
2. Enable and start the service
 
```bash
sudo systemctl daemon-reload
sudo systemctl enable autoban-bot   # start automatically on reboot
sudo systemctl start autoban-bot
sudo systemctl status autoban-bot   # should show: Active: active (running)
```
 
# Useful service commands
 
| Command | Description |
|---|---|
| `sudo systemctl status autoban-bot` | Check if the bot is running |
| `sudo systemctl restart autoban-bot` | Restart the bot |
| `sudo systemctl stop autoban-bot` | Stop the bot |
| `sudo journalctl -u autoban-bot -f` | Watch live logs |
| `sudo journalctl -u autoban-bot -n 50` | Show last 50 log lines |
 
# Updating the bot
 
When updates happen on GitHub, pull and rebuild on your server:
 
```bash
cd ~/autoban-bot
git pull
go build -o autoban-bot
sudo systemctl restart autoban-bot
sudo systemctl status autoban-bot
```
---
## Functionnalities
* Banning members if they select a specified role. 

## Commands
* `/ping` pings the bot, use it if you are unsure whether the bot is online or not.
* `/showbannedrole` shows the role which will trigger a ban upon being picked.
* `/setrole` sets or changes the role to trigger a banned.
* `/disableautoban` disables the bot until turned back on (off by default).
* `/deletedata` deletes all server data on the server it's used on.
* `/setlogchannel` sets the channel where the logs for bans, mutes, etc. will be outputed.
