# Make Your Game — Score Handling

A browser-based Pac-Man clone paired with a Go backend that receives, ranks,
stores, and live-broadcasts player scores. The project focuses on the
score-handling side of a "make your game" exercise: submitting a score,
merging repeat submissions from the same player, ranking all scores, and
pushing live leaderboard updates to connected clients over WebSockets.

## Overview

The client is a vanilla HTML/CSS/JavaScript Pac-Man game (`index.html` +
`static/index.js`) with three selectable maps. When a game ends (win or loss),
the client posts the player's name, score, and elapsed time to a Go server.
The server (`main.go`) keeps all scores in memory, merges repeat submissions
from the same player, re-ranks the full list, and broadcasts the updated
leaderboard to every browser connected over WebSocket. A separate page
(`home.html`, served at `/scores`) renders the leaderboard and updates it live
without a page reload.

## Features

- Playable Pac-Man clone with three maps (`map1`, `map2`, `map3`), unlocked
  progressively in `localStorage` after winning the previous one
  (`static/index.js`, `story.html`).
- Ghost movement/AI, a "souls" (lives) system starting at 3, pause/resume,
  and a mute toggle for sound effects (`Chomp.mp3`, `Death.mp3`).
- Score handling on submit (`POST /receive` in `main.go`):
  - A brand-new player name is appended to the score list.
  - A repeat submission from an existing player name has its score and time
    **averaged** with the previous entry (`(old + new) / 2` in
    `updateExistingScore`), rather than overwritten or duplicated.
  - After every submission, all scores are re-ranked using a manual
    selection-sort (`manualSortAndUpdateRanks`), descending by score, and
    each entry's `Rank` field is recomputed.
- Real-time leaderboard: any score update is broadcast to all clients
  connected to `/ws` (via `github.com/gorilla/websocket`), and `home.html`
  rewrites its table in place when a message arrives.
- Custom error page (`error.html`) rendered for 404/405/500 responses via
  `errorHandler` in `main.go`.

## Technologies

- Go 1.23 (module `pacman`) — standard library `net/http`, `html/template`,
  `encoding/json`, `sync`.
- `github.com/gorilla/websocket` v1.5.3 — WebSocket upgrade/broadcast.
- Vanilla JavaScript, HTML5, CSS3 on the client (no framework or build tool).

## Project Structure

- `main.go` — HTTP server: routes for the landing page (`/`), scoreboard
  (`/scores`), game page (`/index`), score submission (`/receive`), and the
  WebSocket endpoint (`/ws`).
- `story.html` — landing page: player name entry and map selection.
- `index.html` — the game page.
- `home.html` — scoreboard page template, live-updated over WebSocket.
- `error.html` — error page template.
- `static/` — assets served under `/static/`:
  - `index.js` — game logic (movement, ghosts, scoring, sound, score submission)
  - `style.css` — game styling
  - `Images/`, `sound/` — game art and sound effects
  - `main.go` — an earlier variant of the server with slightly different
    routes (e.g. `/story` instead of `/`, no method/error-code checks); it is
    not the entry point used by the module build.
- `text.txt` — an earlier draft of the game/scoring logic kept in the repo,
  not wired into the site.
- `go.mod` / `go.sum` — Go module definition and dependency checksums.
- `COPYRIGHT.md` / `LICENSE` — copyright terms.

## Requirements

- Go 1.23 or later.

## Installation

```
git clone https://github.com/3xoob/make-your-game-score-handling.git
cd make-your-game-score-handling
go build ./...
```

## Usage

Run the server from the repository root:

```
go run main.go
```

This starts an HTTP server bound to `0.0.0.0:8080`. Available routes:

- `GET /` — landing page (`story.html`)
- `GET /index` — the game (`index.html`)
- `GET /scores` — scoreboard (`home.html`)
- `POST /receive` — submit a score as JSON: `{"name": "...", "score": 0, "time": 0}`
- `GET /ws` — WebSocket endpoint that pushes the ranked score list on every update
- `GET /static/...` — static assets (JS, CSS, images, sounds)

## Configuration

- The listen address and port (`0.0.0.0:8080`) are hardcoded in `main.go`;
  there are no environment variables or config files.
- `static/index.js` posts finished-game scores to a hardcoded address,
  `http://10.1.204.167:8080/receive`, and only falls back to
  `http://localhost:8080/receive` if that request fails — see the `a()`
  function in `static/index.js`.

## Limitations

- Scores are held only in the `allScores` in-memory slice; there is no
  database, so all scores are lost on server restart.
- The score-submission and WebSocket endpoints have no authentication, and
  `Access-Control-Allow-Origin` is set to `*`, so any client can submit
  arbitrary scores.
- `static/main.go` duplicates most of the root `main.go` logic but with
  different routing (e.g. `/story` instead of `/`) and without the
  method/error-code checks present in the root version; it appears to be a
  leftover alternate copy rather than an active part of the build.
- The startup log line in `main.go` prints
  `https://localhost:8080/`, but the server is actually started with plain
  HTTP (`http.ListenAndServe`, no TLS).
- `story.html` references map "story" preview videos (`.mp4` files) that are
  not present in the repository.

## License

See [LICENSE](LICENSE) and [COPYRIGHT.md](COPYRIGHT.md). The source is made
available for portfolio/viewing purposes only; copying, modifying,
distributing, or commercial use requires prior written permission from the
copyright holder.
