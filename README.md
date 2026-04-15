# logdrift

> A CLI tool that tails and diffs log streams across multiple services in real time.

---

## Installation

```bash
go install github.com/yourusername/logdrift@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/logdrift.git
cd logdrift
go build -o logdrift .
```

---

## Usage

Tail and diff log streams from multiple services simultaneously:

```bash
logdrift tail --services api,worker,scheduler --diff
```

Compare logs between two services and highlight divergence:

```bash
logdrift diff --left api --right worker --since 10m
```

Follow live output with timestamps and colored diffs:

```bash
logdrift tail --services api,db --timestamps --color --follow
```

### Flags

| Flag | Description |
|------|-------------|
| `--services` | Comma-separated list of service names to tail |
| `--diff` | Enable real-time diffing between streams |
| `--since` | Show logs from the last N minutes/hours (e.g. `5m`, `1h`) |
| `--follow` | Continuously stream new log lines |
| `--timestamps` | Prefix each line with a timestamp |
| `--color` | Enable colored output for diffs |

---

## Requirements

- Go 1.21 or higher

---

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

---

## License

[MIT](LICENSE)