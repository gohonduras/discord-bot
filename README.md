# Geeks-Honduras-GO Discord Bot

[![Discord](https://user-images.githubusercontent.com/7288322/34471967-1df7808a-efbb-11e7-9088-ed0b04151291.png)](https://discord.gg/bdfFKS6)

This repository hosts a Discord bot for the Geeks Honduras GO Discord written in Go as a way to showcase how to write a Go project and use it for real-life tasks. The bot is extensible and allows for several cool features including:

- Searching a term on hackernews (https://news.ycombinator.com)
- TODO: Searching godoc.org documentation for a package function by scraping the website (bonus for using goroutine workers and maintaining some scraped results in memory with a cache for faster results)
- TODO: Converting a Go snippet in the chat into a playground on https://play.golang.org
- TODO: Add more ideas!

## Installation Guidelines

### Running With Go

Install the latest version of [Golang](https://golang.org/dl/) for your operating system, then do:

```
$ go get github.com/gohonduras/discord-bot
```

If you want to try the bot yourself, you can create a new bot in the discord developers portal by following the instructions [here](https://discordpy.readthedocs.io/en/latest/discord.html). Then, run:

```
$ DISCORD_TOKEN=<YOUR_DISCORD_BOT_TOKEN> go run main.go
```

### Docker

```
$ docker build -t gohonduras/discord-bot .
```

Next, run the built docker image:

```
$ docker run -e DISCORD_TOKEN=<YOUR_DISCORD_BOT_TOKEN> gohonduras/discord-bot
```

Output:
```
time="19190-05-05 07:519:193" level=info msg="Bot is now running, press ctrl-c to exit" prefix=main
```

## Running Tests

You can run all tests for the bot with:
```
$ go test ./... -v
```