# myCalendar

myCalendar is a small service written in Go that reads events and tasks from a Google Calendar and sends daily and weekly summaries through a Telegram bot.

The goal of the project is to automate calendar notifications by delivering upcoming events directly to a Telegram chat.

## Features

- Connects to Google Calendar using OAuth2 authentication
- Reads calendar events and tasks
- Sends **daily summaries** of today's events and tasks
- Sends **weekly summaries** of upcoming events and tasks
- Delivers notifications through a **Telegram bot**

## How it works

1. The service authenticates with **Google Calendar API** using OAuth2.
2. Access and refresh tokens are stored locally in a file.
3. The application periodically queries Google Calendar for events and tasks.
4. Events are formatted and sent to a **Telegram chat** via a bot.

## Google Authentication

This application requires a **Google Cloud OAuth2 client**.

You must create OAuth credentials in the Google Cloud Console and obtain:

- `CLIENT_ID`
- `CLIENT_SECRET`

During the first execution, the service will start a temporary callback server to complete the OAuth2 authorization flow and obtain the access and refresh tokens.

These tokens are stored locally in a JSON file defined by `TOKENS_FILE`.

## Environment Variables

The service requires the following environment variables:

```bash
CLIENT_ID=XXX
CLIENT_SECRET=XXX

TELEGRAM_TOKEN=XXX
TELEGRAM_CHAT_ID=XXX

SERVER_HOST=:8080
CALLBACK_HOST=:8081
CALLBACK_URL=http://localhost:8081/callback

TOKENS_FILE=tokens.json