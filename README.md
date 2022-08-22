# lunohodobot telegram bot

## A Telegram bot for my friends community

Current primary functionality is providing random photo on demand and once a day.

## Bot commands:

- `/help`  - show help
- `/photo` - random photo of community
- `/beer`  - implement me

## How to build locally:

```bash
  go mod download
  go build
```

## How to run locally from source code:

```bash
  export TELEGRAM_TOKEN="<bot token>"
  export TELEGRAM_CHATID="<chat id>"
  export CRON_SPEC="0 9 * * *"
  export CRON_JITTER_SEC=1800
  go run ./main.go
```

## How to build docker container:

```bash
  docker build -t lunohodobot .
```

## How to run in docker container:

```bash
docker run -d \
  --name=lunohodobot \
  -e CRON_SPEC="0 9 * * *" \
  -e CRON_JITTER_SEC=1800 \
  -e TELEGRAM_TOKEN="<bot token>" \
  -e TELEGRAM_CHATID="<chat id>" \
  -e TZ=Europe/London \
  -v <folder with photos>:/data \
  -w /data \
  --restart unless-stopped \
  lunohodobot
```

Where:

- `<bot token>` - bot token from @BotFather. See [instructions](https://core.telegram.org/bots#6-botfather);
- `<chat id>` - chat id of community. Other chats will be restricted. Use `curl -X GET https://api.telegram.org/bot<YOUR_API_TOKEN>/getUpdates` after sending a message get chat id;
- `<folder with photos>` - a tree of folders with photos. Bot will use it to get random photo;
- `CRON_SPEC` - set a time for photo of the day in cron format;
- `CRON_JITTER_SEC` - set a jitter for photo of the day in seconds. Bot will send a photo at random time between `CRON_SPEC` and `CRON_SPEC + CRON_JITTER_SEC`. Can be zero.
