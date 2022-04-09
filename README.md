# DevNull's Chatbot ( aka Pointbot )

[![DeepSource](https://deepsource.io/gh/devnull-twitch/go-pointbot.svg/?label=active+issues&show_trend=true&token=cmCOoaNHTyyiLkBWL6m_HGF0)](https://deepsource.io/gh/devnull-twitch/go-pointbot/?ref=repository-badge)

This is a chat bot that provides mainly a alternative point system. In comparison to twitch channel points the broadcaster has commands to give away points.

In addition to the chat bot a REST like API is accessible so stream games or whatever can give away points as well. 

All data is stored in postgres database. To setup the initial database you may use [tern](https://github.com/JackC/tern)  
Adjust the `ops/tern.local.conf` to match your database.

`tern migrate --config ops/tern.local.conf --migrations migration`

Create a `.env.yaml` in the repository root 

```yaml
WEBADDRESS: ":<Port for webserver to listen on>"
USERNAME: "<Your bot account username>"
TOKEN: "oauth:<you bots user access token>"
COMMAND_MARK: "!"
DATABASE_URL: "postgres://user:pass@localhost:5432/database"
TW_CLIENTID: "<Twitch App client ID>"
TW_APP_ACCESS: "<Twitch App access token>"
```