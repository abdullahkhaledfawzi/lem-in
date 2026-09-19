package main

import (
	"sort"
	"strings"
)

// Move is one ant arriving in a room: printed as L<Ant>-<Room>.
type Move struct {
	Ant  int
	Room string
}

// HOW THE SOLVER WORKS (read this first):
//
// Goal: move all ants from Start to End in the fewest turns.
// Rules: 1 ant per middle room, 1 tunnel use per turn.
//
// Idea: find several Start->End paths that share NO middle rooms
// (vertex-disjoint). Then ants never block each other except at
// Start/End, so we space ants 1 turn apart on each path.
//
// Method: Edmonds-Karp max-flow with node splitting, in plain maps.
//   - Split every middle room v into v_in -> v_out (capacity 1).
//     Start/End are NOT split (they hold many ants).
//   - Every tunnel a-b becomes two directed edges
//     a_out -> b_in and b_out -> a_in (capacity 1 each).
//   - Repeatedly BFS the shortest Start->End path in the residual
//     graph, push 1 flow, read back the flow paths, keep the set
//     with the fewest turns.
// No hardcoding: everything derives from Adj + AntCount.

// inNode / outNode map a room to its flow-graph node.
func inNode(name, start, end string) string {
	if name == start || name == end {
		return name
	}
	return name + " in"
}

func outNode(name, start, end string) string {
	if name == start || name == end {
		return name
	}
	return name + " out"
}

// base strips " in"/" out" back to the room name.
func base(node string) string {
	if strings.HasSuffix(node, " in") {
		return node[:len(node)-3]
	}
	if strings.HasSuffix(node, " out") {
		return node[:len(node)-4]
	}
	return node
}

// hasPath reports whether End is reachable from Start.
func hasPath(c *Colony) bool {
	s, t := c.Start.Name, c.End.Name
	seen := map[string]bool{s: true}
	q := []string{s}
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		if u == t {
			return true
		}
		for _, v := range c.Adj[u] {
			if !seen[v] {
				seen[v] = true
				q = append(q, v)
			}
		}
	}
	return false
}

// buildNet creates the flow network.
// Returns res (residual caps, mutated by solver) and cap0 (original).
func buildNet(c *Colony) (res, cap0 map[string]map[string]int) {
	res = map[string]map[string]int{}
	cap0 = map[string]map[string]int{}
	add := func(u, v string) {
		if res[u] == nil {
			res[u] = map[string]int{}
		}
		if cap0[u] == nil {
			cap0[u] = map[string]int{}
		}
		if res[v] == nil {
			res[v] = map[string]int{}
		}
		res[u][v]++
		cap0[u][v]++
	}
	s, t := c.Start.Name, c.End.Name
	for name := range c.Rooms {
		if name == s || name == t {
			continue
		}
		add(name+" in", name+" out") // room holds 1 ant
	}
	for _, l := range c.Links {
		add(outNode(l.From, s, t), inNode(l.To, s, t))
		add(outNode(l.To, s, t), inNode(l.From, s, t))
	}
	return res, cap0
}

// bfsEdge finds the shortest path from src to dst using only edges
// with capacity > 0. Neighbours are sorted so runs are deterministic
// and easy to trace on paper.
func bfsEdge(adj map[string]map[string]int, src, dst string) []string {
	parent := map[string]string{src: ""}
	seen := map[string]bool{src: true}
	q := []string{src}
	for len(q) > 0 {
		u := q[0]
		q = q[1:]
		var nbs []string
		for v, left := range adj[u] {
			if left > 0 && !seen[v] {
				nbs = append(nbs, v)
			}
		}
		sort.Strings(nbs)
		for _, v := range nbs {
			seen[v] = true
			parent[v] = u
			if v == dst {
				p := []string{dst}
				for cur := dst; cur != src; {
					cur = parent[cur]
					p = append([]string{cur}, p...)
				}
				return p
			}
			q = append(q, v)
		}
	}
	return nil
}

// flowToRoom converts a flow-node path to a room path,
// collapsing "v in" -> "v out" pairs into one "v".
func flowToRoom(fp []string) []string {
	var rooms []string
	last := ""
	for _, nd := range fp {
		r := base(nd)
		if r != last {
			rooms = append(rooms, r)
			last = r
		}
	}
	return rooms
}

// currentPaths lists the flow's room paths right now by walking
// edges that carry flow (cap0 - res > 0), shortest first.
func currentPaths(res, cap0 map[string]map[string]int, src, dst string) [][]string {
	flow := map[string]map[string]int{}
	for u, m := range cap0 {
		for v, c0 := range m {
			left := 0
			if res[u] != nil {
				left = res[u][v]
			}
			if c0-left > 0 {
				if flow[u] == nil {
					flow[u] = map[string]int{}
				}
				flow[u][v] = c0 - left
			}
		}
	}
	var out [][]string
	for {
		fp := bfsEdge(flow, src, dst)
		if fp == nil {
			break
		}
		out = append(out, flowToRoom(fp))
		for i := 0; i+1 < len(fp); i++ {
			flow[fp[i]][fp[i+1]]--
		}
	}
	return out
}

// assign spreads ant IDs 1..n greedily: each ant goes to the path
// minimizing (length + already assigned). Optimal for fixed paths.
func assign(paths [][]string, n int) [][]int {
	lens := make([]int, len(paths))
	for i, p := range paths {
		lens[i] = len(p) - 1 // edges
	}
	counts := make([]int, len(paths))
	out := make([][]int, len(paths))
	for id := 1; id <= n; id++ {
		best, bestCost := 0, lens[0]+counts[0]
		for j := 1; j < len(paths); j++ {
			cost := lens[j] + counts[j]
			if cost < bestCost || (cost == bestCost && lens[j] < lens[best]) {
				best, bestCost = j, cost
			}
		}
		out[best] = append(out[best], id)
		counts[best]++
	}
	return out
}

// turnsFor predicts turns for a path set with optimal assignment:
// max over paths of (length + antsOnPath - 1).
func turnsFor(paths [][]string, n int) int {
	as := assign(paths, n)
	max := 0
	for i, p := range paths {
		if len(as[i]) > 0 {
			if t := len(p) - 1 + len(as[i]) - 1; t > max {
				max = t
			}
		}
	}
	return max
}

// findBestPaths augments one path at a time and remembers the
// path set with the fewest predicted turns.
func findBestPaths(c *Colony) [][]string {
	s, t := c.Start.Name, c.End.Name
	res, cap0 := buildNet(c)
	var best [][]string
	bestTurns := -1
	for {
		aug := bfsEdge(res, s, t)
		if aug == nil {
			break
		}
		for i := 0; i+1 < len(aug); i++ {
			u, v := aug[i], aug[i+1]
			res[u][v]--
			if res[v] == nil {
				res[v] = map[string]int{}
			}
			res[v][u]++
		}
		got := currentPaths(res, cap0, s, t)
		if len(got) == 0 {
			continue
		}
		sort.Slice(got, func(i, j int) bool {
			if len(got[i]) != len(got[j]) {
				return len(got[i]) < len(got[j])
			}
			return got[i][1] < got[j][1]
		})
		if turns := turnsFor(got, c.AntCount); bestTurns < 0 || turns < bestTurns {
			bestTurns = turns
			best = append([][]string(nil), got...)
		}
	}
	return best
}

// simulate walks the timetable analytically: ant j (0-based) on its
// path enters at turn j+1 and sits at path[turn-j] on that turn.
// Safe: paths share no middle rooms and ants leave 1 turn apart,
// so no room collision and no tunnel is reused in one turn.
func simulate(paths [][]string, as [][]int) [][]Move {
	max := 0
	for i, p := range paths {
		if len(as[i]) > 0 {
			if t := len(p) - 1 + len(as[i]) - 1; t > max {
				max = t
			}
		}
	}
	var turns [][]Move
	for turn := 1; turn <= max; turn++ {
		var mv []Move
		for i, p := range paths {
			for j, id := range as[i] {
				if step := turn - j; step >= 1 && step < len(p) {
					mv = append(mv, Move{Ant: id, Room: p[step]})
				}
			}
		}
		sort.Slice(mv, func(a, b int) bool { return mv[a].Ant < mv[b].Ant })
		if len(mv) > 0 {
			turns = append(turns, mv)
		}
	}
	return turns
}

// Solve returns per-turn moves plus the chosen room paths.
func Solve(c *Colony) ([][]Move, error) {
	if !hasPath(c) {
		return nil, fail("no path found")
	}
	paths := findBestPaths(c)
	if len(paths) == 0 {
		return nil, fail("no path found")
	}
	return simulate(paths, assign(paths, c.AntCount)), nil
}
