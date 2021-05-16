## Cowin Alerts

A telegram bot for crawling the cowin api and returning
alerts back to the user in case vaccines are available.

### Environment variables

```
BOT_TOKEN=          # Api token for telegram bot
DATA_DIR=           # Directory to store bolt db
PORT=               # Port to run server on
CHECK_INTERVAL=     # Interval (in seconds) to query cowin api
ALERT_DURATION=     # Time between successful alerts for a pin
```