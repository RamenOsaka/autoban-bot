package main

import "github.com/bwmarrin/discordgo"

var configFilePath = "config.json"
var defaultPerms int64 = discordgo.PermissionAdministrator
var encryptionKey []byte

// test server for devs
var testGuildID string = ""
