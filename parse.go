package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Room is one node of the colony.
type Room struct {
	Name    string
	X, Y    int
	IsStart bool
	IsEnd   bool
}

// Link is one tunnel between two rooms.
type Link struct{ From, To string }

// Colony is the parsed farm: rooms, tunnels (both directions in Adj),
// plus the raw file lines so main can echo them.
type Colony struct {
	AntCount int
	Rooms    map[string]*Room
	Start    *Room
	End      *Room
	Adj      map[string][]string
	Links    []Link
	RawInput []string
}

func fail(reason string) error { return &parseError{reason} }

type parseError struct{ reason string }

func (e *parseError) Error() string { return "ERROR: invalid data format, " + e.reason }

// ParseFile reads a farm file and parses it.
func ParseFile(path string) (*Colony, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fail("cannot open file")
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if sc.Err() != nil {
		return nil, fail("cannot read file")
	}
	return ParseLines(lines)
}

// ParseLines turns raw lines into a validated Colony.
// Phases: (1) ant count, (2) rooms then tunnels, (3) final checks.
func ParseLines(lines []string) (*Colony, error) {
	// --- Phase 1: ant count = first non-empty, non-comment line. ---
	i := 0
	ants := 0
	found := false
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		i++
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil || n <= 0 {
			return nil, fail("invalid number of Ants")
		}
		ants, found = n, true
		break
	}
	if !found {
		return nil, fail("invalid number of Ants")
	}
	c := &Colony{AntCount: ants, Rooms: map[string]*Room{}, Adj: map[string][]string{}}
	c.RawInput = lines

	// --- Phase 2: rooms, then tunnels. ---
	var wantStart, wantEnd bool
	startSeen, endSeen := 0, 0
	inTunnels := false
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "##") {
			switch line {
			case "##start":
				startSeen++
				if startSeen > 1 {
					return nil, fail("duplicate start/end room")
				}
				wantStart = true
			case "##end":
				endSeen++
				if endSeen > 1 {
					return nil, fail("duplicate start/end room")
				}
				wantEnd = true
			}
			continue // unknown ##command: ignore
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Tunnel = no spaces, has a dash.
		if !strings.Contains(line, " ") && strings.Contains(line, "-") {
			inTunnels = true
			wantStart, wantEnd = false, false // command before tunnel is ignored
			if err := addLink(c, line); err != nil {
				return nil, err
			}
			continue
		}
		if inTunnels {
			return nil, fail("room after tunnel") // room after tunnel
		}
		if err := addRoom(c, line, wantStart, wantEnd); err != nil {
			return nil, err
		}
		wantStart, wantEnd = false, false
	}

	// --- Phase 3: must have distinct start and end. ---
	if c.Start == nil {
		return nil, fail("no start room found")
	}
	if c.End == nil {
		return nil, fail("no end room found")
	}
	if c.Start.Name == c.End.Name {
		return nil, fail("start and end are the same room")
	}
	return c, nil
}

// addRoom parses "name x y" and registers it.
func addRoom(c *Colony, line string, isStart, isEnd bool) error {
	p := strings.Fields(line)
	if len(p) != 3 {
		return fail("invalid room format")
	}
	name := p[0]
	if strings.HasPrefix(name, "L") || strings.HasPrefix(name, "#") {
		return fail("invalid room name")
	}
	x, err1 := strconv.Atoi(p[1])
	y, err2 := strconv.Atoi(p[2])
	if err1 != nil || err2 != nil {
		return fail("invalid coordinates")
	}
	if _, dup := c.Rooms[name]; dup {
		return fail("duplicate room")
	}
	r := &Room{Name: name, X: x, Y: y, IsStart: isStart, IsEnd: isEnd}
	c.Rooms[name] = r
	if isStart {
		c.Start = r
	}
	if isEnd {
		c.End = r
	}
	return nil
}

// addLink parses "a-b" and registers a tunnel both ways.
func addLink(c *Colony, line string) error {
	k := strings.Index(line, "-")
	if k <= 0 || k >= len(line)-1 {
		return fail("invalid link format")
	}
	a, b := line[:k], line[k+1:]
	if a == "" || b == "" {
		return fail("invalid link format")
	}
	if a == b {
		return fail("self-linking room")
	}
	if _, ok := c.Rooms[a]; !ok {
		return fail("unknown room in link")
	}
	if _, ok := c.Rooms[b]; !ok {
		return fail("unknown room in link")
	}
	for _, n := range c.Adj[a] {
		if n == b {
			return fail("duplicate link")
		}
	}
	c.Adj[a] = append(c.Adj[a], b)
	c.Adj[b] = append(c.Adj[b], a)
	c.Links = append(c.Links, Link{From: a, To: b})
	return nil
}
