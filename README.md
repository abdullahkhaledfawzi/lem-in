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
