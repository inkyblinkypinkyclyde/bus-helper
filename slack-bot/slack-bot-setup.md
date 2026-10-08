# Slack Bot Setup

This application integrates with Slack using a Slack App and Socket Mode. It listens for commands in Slack channels, calls the configured API, and posts the result back to the same channel.

## Prerequisites

You will need:

* A Slack workspace where you can install Slack apps
* A Slack API application
* A Slack bot token (`xoxb-...`)
* A Slack app-level token (`xapp-...`)
* Access to the API used by the bot

---

## 1. Create the Slack App

Go to:

https://api.slack.com/apps

Select **Create New App**.

Choose **From an app manifest**.

Select the Slack workspace where you want to install the bot.

If the repository contains a `manifest.yaml`, paste its contents into the manifest editor.

Alternatively, the following is a minimal example:

```yaml
display_information:
  name: API Bot
  description: Slack bot for calling an external API

features:
  bot_user:
    display_name: API Bot
    always_online: false

oauth_config:
  scopes:
    bot:
      - chat:write
      - channels:history

settings:
  event_subscriptions:
    bot_events:
      - message.channels
  interactivity:
    is_enabled: false
  org_deploy_enabled: false
  socket_mode_enabled: true
  token_rotation_enabled: false
```

Review the configuration and select **Create**.

> The exact scopes and events required may differ depending on which commands the bot implements.

---

## 2. Create the App-Level Token

Socket Mode requires an app-level token.

In the Slack app configuration, go to:

**Basic Information → App-Level Tokens**

Select **Generate Token and Scopes**.

Give the token a name, for example:

```text
socket-mode
```

Add the following scope:

```text
connections:write
```

Generate the token.

The resulting token starts with:

```text
xapp-
```

Save it as:

```text
SLACK_APP_TOKEN
```

---

## 3. Install the App

Go to:

**OAuth & Permissions**

Select **Install to Workspace**.

Review the requested permissions and approve the installation.

Slack will provide a **Bot User OAuth Token** beginning with:

```text
xoxb-
```

Save this as:

```text
SLACK_BOT_TOKEN
```

### Important

Do not commit either token to Git.

Do not put tokens directly into the source code.

---

## 4. Invite the Bot to a Channel

The bot needs to be a member of any channel in which it should receive messages.

In Slack, open the desired channel and run:

```text
/invite @API Bot
```

Alternatively, add the bot through the channel's member management UI.

The bot will then be able to receive messages from that channel, subject to the permissions configured for the app.

---

## 5. Configure the Application

The application requires at least these environment variables:

```bash
SLACK_BOT_TOKEN=xoxb-...
SLACK_APP_TOKEN=xapp-...
```

If the bot calls an external API, configure its API URL and credentials as appropriate for the application.

For example:

```bash
API_URL=https://api.example.com
API_TOKEN=...
```

A `.env` file can be used for local development:

```dotenv
SLACK_BOT_TOKEN=xoxb-...
SLACK_APP_TOKEN=xapp-...

API_URL=https://api.example.com
API_TOKEN=...
```

Do not commit `.env` to source control.

Add it to `.gitignore`:

```gitignore
.env
```

---

## 6. Run the Bot

### Local development

Set the required environment variables and run the application:

```bash
go run .
```

The bot should establish a Socket Mode connection to Slack and begin listening for messages.

---

## 7. Test the Bot

Invite the bot to a channel and issue one of its supported commands.

For example:

```text
@API Bot status server1
```

The bot should:

1. Receive the Slack message.
2. Identify the command.
3. Extract its arguments.
4. Call the configured API.
5. Process the API response.
6. Post the result back to the same Slack channel.

---

# Creating Another Instance

The Slack app can be recreated in another Slack workspace using the app manifest.

The manifest contains the application's configuration, including its bot settings, OAuth scopes, event subscriptions and Socket Mode configuration.

It **does not contain authentication tokens**.

Therefore, each workspace needs its own:

```text
xoxb-...  Bot User OAuth Token
xapp-...  App-Level Token
```

The process is:

1. Open https://api.slack.com/apps
2. Select **Create New App**.
3. Select **From an app manifest**.
4. Select the target Slack workspace.
5. Paste the contents of `manifest.yaml`.
6. Create the app.
7. Generate an app-level token with `connections:write`.
8. Install the app into the workspace.
9. Obtain the Bot User OAuth Token.
10. Configure `SLACK_BOT_TOKEN` and `SLACK_APP_TOKEN`.
11. Invite the bot to the required Slack channels.
12. Start the application.

The same application binary can then be used with the new Slack credentials.

---

# Security

Treat the following as secrets:

```text
SLACK_BOT_TOKEN
SLACK_APP_TOKEN
API_TOKEN
```

Never commit them to Git or include them in the Slack app manifest.

For production deployments, environment variables, Docker secrets, Kubernetes secrets, or another dedicated secret-management system should be used.

---

# Troubleshooting

### Bot does not receive messages

Check that:

* Socket Mode is enabled.
* The app-level token has `connections:write`.
* The bot is a member of the channel.
* `message.channels` is included in the app's event subscriptions.
* `SLACK_APP_TOKEN` and `SLACK_BOT_TOKEN` are correct.
* The application is successfully connected to Slack.

### Bot receives messages but cannot respond

Check that the bot has:

```text
chat:write
```

and that the bot has been installed after changing its OAuth scopes.

If scopes have been changed, Slack may require the app to be **reinstalled to the workspace**.

### Bot responds to its own messages

The application should ignore messages originating from the bot itself. Otherwise, the bot can accidentally create a response loop.

### Commands work in one channel but not another

Make sure the bot has been invited to the other channel. For private channels, the bot must explicitly be added as a member.
