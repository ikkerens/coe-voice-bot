# Chronicles of Elyria voice text channel bot [![Build Status](https://travis-ci.org/ikkerens/coe-voice-bot.svg?branch=master)](https://travis-ci.org/ikkerens/coe-voice-bot) [![Go Report Card](https://goreportcard.com/badge/github.com/ikkerens/coe-voice-bot)](https://goreportcard.com/report/github.com/ikkerens/coe-voice-bot) 

This bot is originally made for use in the [Chronicles of Elyria discord server](https://discord.gg/chroniclesofelyria).  
It allows anyone with the `MANAGE_CHANNELS` permission to link voice and text channels to allow people in voice to
communicate with those that do not have a microphone available.

This bot is not hosted anywhere (so no invite link), it is expected for you to host & run it yourself.

## Building
This application is made using [the Go Language](https://golang.org/), version 1.9+, which has to be installed to your PATH in order for these commands to work.
Most linux distributions have at least version 1.6 installed in their package repositories under the names `golang` or `golang-go`.  
*If you do not have the GOPATH environment variable configured, it will default to `$HOME/go` on unix-like.*

The two commands below will download all dependencies and then install the bot binary to `$GOPATH/bin`
```sh
# Clone the project and get all dependencies without building
go get -d github.com/ikkerens/coe-voice-bot
# Install the binary to $GOPATH/bin
go install github.com/ikkerens/coe-voice-bot
```

## Running
Unix-like:
```sh
export TOKEN=PASTEYOURTOKENHERE
$GOPATH/bin/coe-voice-bot
```

Systemd-service:
```
[Unit]
Description=CoE Voice Bot
Wants=network.target
After=multi-user.target network.target

[Service]
# It is recommended to change the user to something less privileged.
# The user will need writing permission to the WorkingDirectory to write config.json containing all active links.
User=root
Type=simple
Environment=TOKEN=PASTEYOURTOKENHERE
WorkingDirectory=/etc/coe-voice-bot
ExecStart=/path/to/binary/coe-voice-bot
Restart=always
RestartSec=15

[Install]
WantedBy=multi-user.target
```

## Usage
The bot needs the following permissions to operate:
* `MANAGE_ROLES`
* The ability to see the channels it needs to manage.
* The ability to send messages in the channel commands are executed in, to provide meaningful error messages.

The bot currently knows two commands, both require the user to have the `MANAGE_CHANNELS` permission serverwide:

##### !voicelink \<voiceChannelID> <textChannelID|textChannelMention>
This command will make a link between the specified voice chat channel and the specified text channel.  
Example: `!voicelink 118109806723727364 #voice-chat`

##### !voiceunlink \<voiceChannelID>
This command will remove link for the specified voice channel.  
Example: `!voiceunlink 118109806723727364`

##### !voicelinklist
This command will list all currently known and active channel links.
