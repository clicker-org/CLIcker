# Setting up CLIcker

Go 1.24+ is required. That's the only real dependency you need to install yourself — everything else gets pulled in by `go mod tidy`.

## Quick start

```
git clone https://github.com/clicker-org/clicker
cd clicker
go mod tidy
make run
```

That builds and launches the game. Your save file ends up at `~/.config/clicker/save.json` (or `$XDG_CONFIG_HOME/clicker/save.json` if you have that set).

## Building

```bash
# build for your current platform → bin/clicker
make build

# cross-compile for linux and macOS (amd64 + arm64) → dist/
make build-all

# same as build-all but prints a summary line at the end
make release
```

If you don't want to use make, the plain go command works fine:

```bash
go build -o bin/clicker ./cmd/clicker
```

## Running

```bash
make run        # builds first, then runs
./bin/clicker   # if you've already built
```

The game uses the alternate screen buffer, so your terminal history stays clean. Minimum recommended terminal size is 80×24.

## Testing

```bash
make test
# or
go test ./...
```

Tests live next to the code they cover (`_test.go` files in the same package). The packages with meaningful test coverage are `internal/player`, `internal/economy`, `internal/offline`, `internal/config`, and `internal/save`. UI code isn't unit tested — that gets covered by just playing the game.

To run tests for a specific package:

```bash
go test ./internal/economy/...
go test ./internal/save/... -v    # -v shows individual test names
```

## Linting

You'll need [golangci-lint](https://golangci-lint.run/usage/install/) installed for this:

```bash
make lint
```

## Cleaning up

```bash
make clean    # removes bin/ and dist/
make purge    # same as clean, also deletes the save file
```

## Dev notes

- The save file is written on quit (`Q`) and every 30 seconds while running. To start fresh during dev, use `make purge`.
- World configs live in `configs/worlds/` as TOML files. You can edit balance values there without recompiling — the game reads them at startup.
- Adding a new world means creating a `.go` file in `internal/world/worlds/` and a `.toml` in `configs/worlds/`. Nothing else needs to change.
- The `internal/` packages have no Bubble Tea imports. Keep it that way.
