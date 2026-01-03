# SockRacer

SOCKS5 parallel racing aggregator

## Usage

### Config file mode

Create a config file (default: `config.json`):

```json
{
  "listeners": [
    {
      "listen": "127.0.0.1:1080",
      "socks": [
        {
          "name": "US-West-1",
          "address": "upstream1:1081"
        },
        {
          "name": "US-East-1",
          "address": "upstream2:1082"
        }
      ]
    }
  ]
}
```

Run with default config file:
```bash
sockracer
```

Run with custom config file:
```bash
sockracer --config /path/to/config.json
sockracer -c /path/to/config.json
```

### Command line mode

```bash
sockracer --listen-address 127.0.0.1 --listen-port 1080 --socks upstream1:1081 --socks upstream2:1082
```

## Command Line Options

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--config` | `-c` | `config.json` | Path to config file |
| `--listen-address` | | `127.0.0.1` | Listen address for SOCKS5 server |
| `--listen-port` | | | Listen port (required for command line mode) |
| `--socks` | | | Upstream SOCKS5 proxy (can be specified multiple times) |
| `--help` | `-h` | | Show help message |

### Notes

- Config file mode: If `--listen-port` is not specified, the program will load configuration from the config file
- Command line mode: If `--listen-port` is specified, at least one `--socks` upstream must be provided
- The `--socks` option can be used multiple times to specify multiple upstream proxies

## Build

```bash
go build -o sockracer
```

## Testing

Test the SOCKS5 proxy with curl:

```bash
curl --socks5 127.0.0.1:1080 http://ipinfo.io
```

Check your IP address:

```bash
curl --socks5 127.0.0.1:1080 https://api.ipify.org
```

Test connection speed:

```bash
curl --socks5 127.0.0.1:1080 -w "\nTime: %{time_total}s\n" -o /dev/null -s http://www.google.com
```

## Example Output

```
2026/01/03 17:25:36 listening on 127.0.0.1:1080 with 6 upstreams
2026/01/03 17:26:39 → new connection from 127.0.0.1:58992
2026/01/03 17:26:39 request from 127.0.0.1:58992 to 104.26.13.205:443
2026/01/03 17:26:39 racing 6 upstreams for 104.26.13.205:443
2026/01/03 17:26:39 ✓ winner: US-West-1 (198.18.169.1:1080) (114ms)
2026/01/03 17:26:40 relaying data for 127.0.0.1:58992 -> 104.26.13.205:443
2026/01/03 17:26:40 ← connection closed from 127.0.0.1:58992
```
