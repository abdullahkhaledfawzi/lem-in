/* =============================================================================
   lem-in visualizer — plain vanilla JS, no libraries.

   How it works, top to bottom:
     1. parseOutput()  : turns the raw lem-in text into data structures.
     2. layout()       : maps room (x y) coordinates onto the SVG viewBox.
     3. positionsFor() : precomputes where every ant stands at every turn.
     4. buildGraph()   : draws tunnels, rooms and ants once.
     5. renderTurn()   : moves the ants of one turn (CSS does the animation).
     6. Controls       : prev / next / reset / autoplay with a speed slider.
   ============================================================================= */

// ----------------------------- 1. parsing -----------------------------------

/**
 * Parse the raw output of lem-in (farm + blank line + turn lines).
 * Example input:
 *   4
 *   ##start
 *   0 0 3
 *   ...
 *   0-2
 *
 *   L1-2
 *   L1-3 L2-2
 * Returns { ants, rooms, tunnels, start, end, turns } or throws on bad data.
 */
function parseOutput(text) {
  const lines = text.split(/\r?\n/);

  // The farm ends at the first empty line; everything after it is moves.
  const blank = lines.findIndex((l) => l.trim() === "");
  const farm = blank === -1 ? lines : lines.slice(0, blank);
  const moveLines = blank === -1 ? [] : lines.slice(blank + 1);

  let ants = 0;
  let start = null;
  let end = null;
  let pending = null; // "start" | "end" — which room the next line defines
  const rooms = new Map(); // name -> { x, y }
  const tunnels = []; // [nameA, nameB]

  for (const raw of farm) {
    const line = raw.trim();
    if (!line || line.startsWith("#")) {
      if (line === "##start") pending = "start";
      if (line === "##end") pending = "end";
      continue; // comments and commands carry no geometry
    }
    if (line.includes("-")) {
      // tunnel: "a-b"
      const [a, b] = line.split("-");
      tunnels.push([a, b]);
      continue;
    }
    const parts = line.split(/\s+/);
    if (parts.length === 1 && /^\d+$/.test(parts[0])) {
      ants = parseInt(parts[0], 10); // first plain number = ant count
      continue;
    }
    if (parts.length >= 3) {
      // room: "name x y"
      const name = parts[0];
      rooms.set(name, { x: parseFloat(parts[1]), y: parseFloat(parts[2]) });
      if (pending === "start") start = name;
      if (pending === "end") end = name;
      pending = null;
    }
  }

  if (!ants) throw new Error("could not find the ant count");
  if (!rooms.has(start) || !rooms.has(end))
    throw new Error("could not find ##start / ##end rooms");

  // One turn per non-empty move line: "L1-3 L2-2" -> [{ant:1, room:"3"}, ...]
  const turns = moveLines
    .filter((l) => l.trim() !== "")
    .map((l) =>
      l
        .trim()
        .split(/\s+/)
        .map((tok) => {
          const m = tok.match(/^L(\d+)-(.+)$/);
          if (!m) throw new Error(`bad move token: "${tok}"`);
          return { ant: parseInt(m[1], 10), room: m[2] };
        })
    );

  return { ants, rooms, tunnels, start, end, turns };
}

// ----------------------------- 2. layout ------------------------------------

/**
 * Scale the farm's raw coordinates to fit the SVG viewBox (y is flipped so
 * bigger y is drawn higher). Rooms get a position { x, y } added to them.
 */
function layout(data, width = 1000, height = 640, pad = 70) {
  const xs = [...data.rooms.values()].map((r) => r.x);
  const ys = [...data.rooms.values()].map((r) => r.y);
  const minX = Math.min(...xs), maxX = Math.max(...xs);
  const minY = Math.min(...ys), maxY = Math.max(...ys);
  // avoid division by zero when all rooms share one coordinate
  const spanX = maxX - minX || 1;
  const spanY = maxY - minY || 1;

  for (const room of data.rooms.values()) {
    room.pos = {
      x: pad + ((room.x - minX) / spanX) * (width - 2 * pad),
      y: height - (pad + ((room.y - minY) / spanY) * (height - 2 * pad)),
    };
  }
  return { width, height };
}

// ------------------- 3. ant positions ---------------------------------------

/** Precompute every ant's room at every turn. Index 0 = all ants at start. */
function positionsFor(data) {
  const positions = [new Map()];
  for (let ant = 1; ant <= data.ants; ant++)
    positions[0].set(ant, data.start);

  data.turns.forEach((turn, i) => {
    const next = new Map(positions[i]); // ants that don't move stay put
    for (const mv of turn) next.set(mv.ant, mv.room);
    positions.push(next);
  });
  return positions;
}

// ==== PART 2 (graph building, rendering, controls) is appended below ====

// ------------------- 4. graph building --------------------------------------

/** Draw tunnels, rooms and ants into the SVG. Called exactly once. */
function buildGraph(svg, data) {
  const ns = "http://www.w3.org/2000/svg";
  const make = (tag) => document.createElementNS(ns, tag);

  svg.innerHTML = `
    <defs>
      <radialGradient id="ant-gradient" cx="35%" cy="30%">
        <stop offset="0%" stop-color="#a7f3d0"/>
        <stop offset="100%" stop-color="#14b8a6"/>
      </radialGradient>
    </defs>`;

  // tunnels first, so rooms are drawn on top of them
  const edgeEls = new Map(); // "a|b" -> line element (for turn highlighting)
  for (const [a, b] of data.tunnels) {
    const line = make("line");
    line.setAttribute("x1", data.rooms.get(a).pos.x);
    line.setAttribute("y1", data.rooms.get(a).pos.y);
    line.setAttribute("x2", data.rooms.get(b).pos.x);
    line.setAttribute("y2", data.rooms.get(b).pos.y);
    line.setAttribute("class", "tunnel");
    svg.appendChild(line);
    edgeEls.set([a, b].sort().join("|"), line);
  }

  // rooms: a circle with the name underneath
  for (const [name, room] of data.rooms) {
    const g = make("g");
    const role =
      name === data.start ? " start" : name === data.end ? " end" : "";
    g.setAttribute("class", "room" + role);
    const circle = make("circle");
    circle.setAttribute("cx", room.pos.x);
    circle.setAttribute("cy", room.pos.y);
    circle.setAttribute("r", 22);
    const label = make("text");
    label.setAttribute("x", room.pos.x);
    label.setAttribute("y", room.pos.y + 38);
    label.textContent = name;
    g.appendChild(circle);
    g.appendChild(label);
    svg.appendChild(g);
    room.el = g;
  }

  // ants: one group per ant, moved later via a transform
  const antEls = new Map();
  for (let ant = 1; ant <= data.ants; ant++) {
    const g = make("g");
    g.setAttribute("class", "ant");
    const circle = make("circle");
    circle.setAttribute("r", 14);
    const label = make("text");
    label.textContent = ant;
    g.appendChild(circle);
    g.appendChild(label);
    svg.appendChild(g);
    antEls.set(ant, g);
  }
  return { edgeEls, antEls };
}

// ==== PART 3 (rendering + controls) continues below ====

// ------------------- 5. rendering -------------------------------------------

let state = null; // filled by load(): { data, positions, graph, turn }

function renderTurn(turn) {
  const { data, positions, graph } = state;
  state.turn = turn;

  // move every ant to its room for this turn (CSS animates the change)
  for (const [ant, roomName] of positions[turn]) {
    const pos = data.rooms.get(roomName).pos;
    graph.antEls
      .get(ant)
      .setAttribute("transform", `translate(${pos.x} ${pos.y})`);
  }

  // highlight the tunnels used during this turn (from turn-1 -> here)
  const prev = positions[turn - 1] || positions[0];
  for (const line of graph.edgeEls.values()) line.classList.remove("active");
  if (turn > 0) {
    for (const mv of data.turns[turn - 1]) {
      const from = prev.get(mv.ant);
      const key = [from, mv.room].sort().join("|");
      const line = graph.edgeEls.get(key);
      if (line) line.classList.add("active");
    }
  }

  // rooms holding ants get a subtle fill change
  const crowded = new Set(positions[turn].values());
  for (const [name, room] of data.rooms)
    room.el.classList.toggle("crowded", crowded.has(name));

  // text panels
  document.getElementById("turn-now").textContent = turn;
  document.getElementById("turn-total").textContent = positions.length - 1;
  document.getElementById("log-title").textContent = `Turn ${turn}`;
  const moves = turn === 0 ? [] : data.turns[turn - 1];
  document.getElementById("log-moves").textContent = moves.length
    ? moves.map((m) => `L${m.ant}-${m.room}`).join("   ")
    : "— no moves this turn —";
  const arrived = [...positions[turn].values()].filter(
    (r) => r === data.end
  ).length;
  document.getElementById("arrived").textContent = arrived;
  document.getElementById("arrived-total").textContent = data.ants;

  // button availability
  document.getElementById("btn-prev").disabled = turn === 0;
  document.getElementById("btn-next").disabled = turn === positions.length - 1;
}

function step(delta) {
  const last = state.positions.length - 1;
  renderTurn(Math.min(last, Math.max(0, state.turn + delta)));
}

// ==== PART 4 (controls + loading) continues below ====

// ------------------- 6. controls --------------------------------------------

let playTimer = null;

function msPerTurn() {
  // The slider reads 100 (left) .. 2000 (right). Right should be FAST, so
  // invert it into a delay: left = 2000 ms/turn (slow 🐢), right = 100 ms/turn (fast 🐇).
  return 2100 - parseInt(document.getElementById("speed").value, 10);
}

function play() {
  if (state.turn >= state.positions.length - 1) renderTurn(0); // restart
  document.getElementById("btn-play").textContent = "⏸";
  const tick = () => {
    if (state.turn >= state.positions.length - 1) return pause();
    step(1);
    playTimer = setTimeout(tick, msPerTurn());
  };
  playTimer = setTimeout(tick, msPerTurn() / 2);
}

function pause() {
  clearTimeout(playTimer);
  playTimer = null;
  document.getElementById("btn-play").textContent = "▶";
}

function setupControls() {
  document.getElementById("btn-prev").onclick = () =>
    playTimer ? pause() : step(-1);
  document.getElementById("btn-next").onclick = () =>
    playTimer ? pause() : step(1);
  document.getElementById("btn-reset").onclick = () => {
    pause();
    renderTurn(0);
  };
  document.getElementById("btn-play").onclick = () =>
    playTimer ? pause() : play();
  // changing the speed mid-play simply affects the next scheduled tick
  document.getElementById("speed").oninput = () => {};
}

// ------------------- loading -------------------------------------------------

function load(text) {
  pause();
  const data = parseOutput(text);
  layout(data);
  const positions = positionsFor(data);
  const graph = buildGraph(document.getElementById("svg"), data);
  state = { data, positions, graph, turn: 0 };

  // fill the header stats
  document.getElementById("stat-ants").textContent = data.ants;
  document.getElementById("stat-rooms").textContent = data.rooms.size;
  document.getElementById("stat-tunnels").textContent = data.tunnels.length;
  document.getElementById("stat-turns").textContent = data.turns.length;
  document.getElementById("arrived-total").textContent = data.ants;

  document.getElementById("drop-hint").classList.add("hidden");
  renderTurn(0);
}

function showError(message) {
  const hint = document.getElementById("drop-hint");
  hint.classList.remove("hidden");
  hint.querySelector("p").textContent = `⚠ ${message}`;
}

/** Try to fetch the piped output file written by serve.sh. */
function tryLoadPipedOutput() {
  fetch("output.txt")
    .then((res) => (res.ok ? res.text() : Promise.reject(res.status)))
    .then(load)
    .catch(() => {
      /* no output.txt — stay on the drop hint, that's fine */
    });
}

// Only touch the DOM when running in a browser (keeps app.js importable
// for quick parser tests with node).
if (typeof document !== "undefined") {
  setupControls();
  tryLoadPipedOutput();

  // fallback without a server: pick or drop a file
  const hint = document.getElementById("drop-hint");
  document.getElementById("file-input").onchange = (e) => {
    const file = e.target.files[0];
    if (file) file.text().then(load).catch((err) => showError(err.message));
  };
  ["dragover", "dragleave", "drop"].forEach((evt) =>
    hint.addEventListener(evt, (e) => {
      e.preventDefault();
      hint.classList.toggle("dragover", evt === "dragover");
      const file = evt === "drop" ? e.dataTransfer.files[0] : null;
      if (file) file.text().then(load).catch((err) => showError(err.message));
    })
  );
  document.getElementById("board").addEventListener("dragover", (e) => {
    e.preventDefault();
    hint.classList.remove("hidden");
  });
}
