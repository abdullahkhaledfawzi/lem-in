# lem-in

A Go program that moves ants across a colony from `##start` to `##end`
in the fewest turns. Standard library only.

## What it does

Reads a farm file (`ants + rooms + tunnels`), finds the set of
vertex-disjoint shortest paths that minimizes total turns
(Edmonds-Karp max-flow with node splitting), spreads ants greedily
over those paths, and prints the file followed by one
`L<ant>-<room> ...` line per turn.

## Usage

Run from source:

```bash
go run . tests/example00.txt
```

Or use a prebuilt executable (Linux `lem-in`, Windows `lem-in.exe`):

```bash
./lem-in tests/example00.txt
lem-in.exe tests/example00.txt
```

Build them yourself with:

```bash
go build -o lem-in .                         # Linux
GOOS=windows GOARCH=amd64 go build -o lem-in.exe .   # Windows
```

Output = original file + blank line + moves:

```
4
##start
0 0 3
...

L1-2
L1-3 L2-2
...
```

## Program structure (single package, trace top to bottom)

1. `main.go`: args -> `ParseFile` -> `Solve` -> print.
2. `parse.go`: ant count, then rooms, then tunnels,
   then start/end checks.
3. `solve.go`: the path search and ant scheduling in small funcs:
   `hasPath -> buildNet -> bfsEdge -> currentPaths/flowToRoom ->`
   `findBestPaths -> assign/turnsFor -> simulate -> Solve`.

Example: run `go run . tests/example00.txt` and follow
`findBestPaths`: one BFS finds `0-2-3-1`, no second disjoint path
exists, so all 4 ants queue on it = 6 turns.

## Rules enforced

- First non-comment line must be ants > 0.
- Room `name x y`: name has no spaces, does not start with `L`/`#`,
  `x y` are ints, no duplicate names.
- Tunnel `a-b`: both rooms exist, no `a-a`, no duplicate `a-b`/`b-a`.
- Exactly one `##start` and one `##end`, distinct, rooms only
  before the first tunnel. Unknown `##cmd` ignored, `#` ignored.
- A path must exist or it prints `ERROR: invalid data format, ...`.
- Simulation keeps 1 ant per middle room and 1 tunnel use per turn.

## Tests

```bash
bash run.sh        # vet + unit tests + all examples + perf
go test ./... -v   # unit tests only
```

---

## 🎨 Visualizer

Watch the colony solve itself! The [`visualizer/`](visualizer/) folder contains
a tiny, dependency-free visualizer (plain **HTML + CSS + JS**) that renders the
farm — rooms 🏛, tunnels 🔗, `##start` 🟢 and `##end` 🔴 — and replays the
solution turn by turn with animated ants 🐜.

### 🚀 Quick start (pipe the output straight in)

```bash
./lem-in tests/example00.txt | visualizer/serve.sh
```

`serve.sh` saves the piped output and serves the page at
**http://localhost:8000** (it opens your browser automatically).
Change the port with `PORT=9000 visualizer/serve.sh`.

You can also let it run lem-in for you:

```bash
visualizer/serve.sh tests/example01.txt
```

### 📂 No server? No problem

Just open `visualizer/index.html` in your browser and **drag & drop** any
lem-in output file onto the page (or use the file picker).

### 🎮 Controls

| Button | Action |
|:------:|--------|
| ⏮ / ⏭ | Step one turn backward / forward |
| ▶ / ⏸ | Auto play / pause — replays the whole solution |
| 🐢–🐇 | Speed slider: how long each turn takes (100 ms – 2 s) |
| ⟲ | Reset back to turn 0 |
| 🏁 | "Arrived" counter shows how many ants reached `##end` |

### 🧠 How to read it

- 🟢 **Green room** = `##start`, 🔴 **red room** = `##end`.
- 🐜 **Numbered discs** are the ants; they glide along tunnels each turn.
- 🟡 **Amber tunnels** are the ones used during the current turn.
- The **turn log** below the board lists the exact `L<ant>-<room>` moves of
  the current turn — same lines lem-in prints.

### 🗂 Files

```
visualizer/
├── index.html   ← page structure (board, controls, log)
├── style.css    ← dark theme, room/ant/tunnel styling + animation
├── app.js       ← parsing, layout, rendering, playback (~350 commented lines)
└── serve.sh     ← takes lem-in's output via pipe and serves the page
```

---

Created by Abdullah Khaled from Egypt with ❤️
