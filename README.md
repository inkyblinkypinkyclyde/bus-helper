# bus-panel

Live bus info from the Bus Open Data Service (BODS), in three parts:

| Directory | What it is |
| --- | --- |
| [bods-helper](bods-helper/) | Python (Flask) API that fetches and processes live BODS data |
| [frontend](frontend/) | React + Vite map UI that shows live buses via bods-helper |
| [slack-bot](slack-bot/) | Go Slack bot that queries the API from Slack channels |

## bods-helper

```sh
cd bods-helper
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python app.py
```

Listens on http://127.0.0.1:5000. Run tests with `make test`. See [bods-helper/README.md](bods-helper/README.md) for endpoints and CLI helpers.

## frontend

Start bods-helper first, then either:

```sh
cd frontend
docker compose up        # no Node needed
# or
npm install && npm run dev
```

Open http://localhost:5173. See [frontend/README.md](frontend/README.md).

## slack-bot

Requires Go and a Slack app with Socket Mode (see [slack-bot/slack-bot-setup.md](slack-bot/slack-bot-setup.md)).

```sh
cd slack-bot
export SLACK_BOT_TOKEN=xoxb-...
export SLACK_APP_TOKEN=xapp-...
export API_BASE_URL=http://127.0.0.1:5000
go run .
```

Run tests with `make test`, or build binaries with `make build` (current platform) or `make build-all`.
