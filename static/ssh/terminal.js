// killdeer.digital browser terminal.
//
// xterm.js draws the screen. killdeer-ssh.wasm (Go, see webssh/) does all of
// the SSH work in this tab: it opens a WebSocket to the relay at ./_ws/,
// checks the pinned host key, relays the gateway's password prompts to us and
// streams the session. Nothing typed here is sent anywhere except inside that
// encrypted SSH session.

import { Terminal } from './vendor/xterm.mjs';
import { WebglAddon } from './vendor/addon-webgl.mjs';

const COLS = 80;          // The Killdeer TUI renders at a fixed 80 columns.
const MIN_ROWS = 24;
const LINE_HEIGHT = 1.12;
const FONT = '"Martian Mono", ui-monospace, "SFMono-Regular", Menlo, Consolas, monospace';
const reducedMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;

const WORDMARK = [
  '    __   _ ____    __',
  '   / /__(_) / /___/ /__  ___  _____',
  '  / //_/ / / / __  / _ \\/ _ \\/ ___/',
  ' / ,< / / / / /_/ /  __/  __/ /',
  '/_/|_/_/_/_/\\__,_/\\___/\\___/_/',
  '                  .digital',
];

// ---------------------------------------------------------------- settings

const prefs = {
  get(key, fallback) {
    try { return localStorage.getItem('killdeer-term:' + key) ?? fallback; } catch { return fallback; }
  },
  set(key, value) {
    try { localStorage.setItem('killdeer-term:' + key, value); } catch { /* private mode */ }
  },
};

function mix(a, b, t) {
  const pa = [1, 3, 5].map((i) => parseInt(a.slice(i, i + 2), 16));
  const pb = [1, 3, 5].map((i) => parseInt(b.slice(i, i + 2), 16));
  return '#' + pa.map((v, i) => Math.round(v + (pb[i] - v) * t).toString(16).padStart(2, '0')).join('');
}

// Phosphor screens are monochrome: every ANSI colour becomes a brightness
// level of the one phosphor, so the TUI keeps its contrast but not its hues.
function phosphorTheme(fg, bg) {
  const level = (t) => mix(bg, fg, t);
  const bright = (t) => mix(level(t), '#ffffff', 0.22);
  return {
    foreground: level(0.92), background: bg, cursor: fg, cursorAccent: bg,
    selectionBackground: mix(bg, fg, 0.35), selectionForeground: bg,
    black: level(0.18), red: level(0.78), green: level(0.9), yellow: level(1),
    blue: level(0.62), magenta: level(0.7), cyan: level(0.84), white: level(0.92),
    brightBlack: level(0.45), brightRed: bright(0.85), brightGreen: bright(0.95), brightYellow: bright(1),
    brightBlue: bright(0.72), brightMagenta: bright(0.8), brightCyan: bright(0.9), brightWhite: '#fff8ea',
  };
}

const THEMES = {
  amber: phosphorTheme('#ffb000', '#120b02'),
  green: phosphorTheme('#4dff88', '#021208'),
  plain: {
    foreground: '#ededed', background: '#171717', cursor: '#ededed', cursorAccent: '#171717',
    selectionBackground: '#4a4a4a',
    black: '#232323', red: '#ff6b5e', green: '#7bd88f', yellow: '#f2c66d',
    blue: '#78a9ff', magenta: '#d79cff', cyan: '#6fd6d6', white: '#d4d4d4',
    brightBlack: '#6b6b6b', brightRed: '#ff8f85', brightGreen: '#a2e9b0', brightYellow: '#ffd98e',
    brightBlue: '#a1c4ff', brightMagenta: '#e6bfff', brightCyan: '#9ee8e8', brightWhite: '#ffffff',
  },
};

// ---------------------------------------------------------------- terminal

const screenEl = document.getElementById('screen');
const termEl = document.getElementById('terminal');
const announcer = document.getElementById('announcer');

let theme = THEMES[prefs.get('theme', 'amber')] ? prefs.get('theme', 'amber') : 'amber';
let fx = prefs.get('fx', reducedMotion ? 'off' : 'on') === 'on';

const term = new Terminal({
  cols: COLS,
  rows: MIN_ROWS,
  fontFamily: FONT,
  fontSize: 14,
  lineHeight: LINE_HEIGHT,
  letterSpacing: 0,
  cursorBlink: true,
  cursorStyle: 'block',
  scrollback: 2000,
  drawBoldTextInBrightColors: true,
  theme: THEMES[theme],
});

function applySettings() {
  document.documentElement.dataset.theme = theme;
  document.body.classList.toggle('fx', fx);
  term.options.theme = THEMES[theme];
  document.querySelectorAll('[data-theme-choice]').forEach((b) => {
    b.setAttribute('aria-pressed', String(b.dataset.themeChoice === theme));
  });
  const fxButton = document.getElementById('fx-toggle');
  fxButton.setAttribute('aria-pressed', String(fx));
  fxButton.textContent = fx ? 'fx on' : 'fx off';
}

document.querySelectorAll('[data-theme-choice]').forEach((b) => {
  b.addEventListener('click', () => {
    theme = b.dataset.themeChoice;
    prefs.set('theme', theme);
    applySettings();
    term.focus();
  });
});
document.getElementById('fx-toggle').addEventListener('click', () => {
  fx = !fx;
  prefs.set('fx', fx ? 'on' : 'off');
  applySettings();
  term.focus();
});

// Fit exactly 80 columns across the screen and as many rows as fit (at
// least 24). Narrow screens scroll sideways rather than break the TUI.
let session = null;
function layout() {
  const width = termEl.clientWidth;
  const height = termEl.clientHeight;
  if (!width || !height) return;
  const probe = document.createElement('canvas').getContext('2d');
  probe.font = `100px ${FONT}`;
  const advance = probe.measureText('M'.repeat(10)).width / 1000; // per px of font size
  const fontSize = Math.max(8, Math.min(22, Math.floor((width - 2) / (COLS * advance) * 4) / 4));
  term.options.fontSize = fontSize;
  const cellHeight = Math.ceil(fontSize * LINE_HEIGHT) || 1;
  const rows = Math.max(MIN_ROWS, Math.floor(height / cellHeight));
  if (rows !== term.rows) term.resize(COLS, rows);
  session?.resize(COLS, term.rows);
}
let layoutTimer = 0;
new ResizeObserver(() => {
  clearTimeout(layoutTimer);
  layoutTimer = setTimeout(layout, 60);
}).observe(termEl);

// ---------------------------------------------------------------- status LEDs

function led(id, state) {
  const el = document.getElementById('led-' + id);
  if (state) el.dataset.state = state; else delete el.dataset.state;
}
function resetLeds(keepCore = true) {
  ['relay', 'key', 'auth'].forEach((id) => led(id));
  if (!keepCore) led('core');
}
function announce(text) { announcer.textContent = text; }

// ---------------------------------------------------------------- output helpers

const ESC = '\x1b[';
const dim = (s) => `${ESC}2m${s}${ESC}22m`;
const bold = (s) => `${ESC}1m${s}${ESC}22m`;
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let skipTyping = reducedMotion;
async function type(text, cps = 140) {
  for (const ch of text) {
    term.write(ch);
    if (!skipTyping && ch !== ' ') await sleep(1000 / cps);
  }
}

function leader(label, width = 46) {
  return `  > ${label} ${dim('.'.repeat(Math.max(3, width - label.length)))} `;
}

// Everything that writes to the screen during a connection runs through one
// queue, so a status line finishes typing before the next output lands.
let chain = Promise.resolve();
const inOrder = (fn) => (chain = chain.then(fn, fn));

let pendingStep = null;
async function step(label) {
  pendingStep = label;
  await type(leader(label), 260);
}
function stepDone(result = 'ok') {
  if (!pendingStep) return;
  term.write(bold(result) + '\r\n');
  pendingStep = null;
}
function stepFailed(result = 'fail') {
  if (!pendingStep) return;
  term.write(`${ESC}7m ${result} ${ESC}27m\r\n`);
  pendingStep = null;
}

// ---------------------------------------------------------------- input routing

// Keystrokes go to whichever of these is active: the local line editor
// (username and password prompts), the live SSH session, or the
// "press enter to reconnect" wait.
let lineReader = null;
let onEnter = null;

function readLine(prompt, echo) {
  return new Promise((resolve, reject) => {
    term.write(prompt);
    lineReader = { buffer: '', echo, resolve, reject };
  });
}

function feedLine(data) {
  const r = lineReader;
  for (let i = 0; i < data.length; i++) {
    const ch = data[i];
    if (ch === '\x1b') return; // ignore arrows and other escape sequences
    if (ch === '\r' || ch === '\n') {
      lineReader = null;
      term.write('\r\n');
      r.resolve(r.buffer);
      return;
    }
    if (ch === '\x03') {
      lineReader = null;
      term.write('^C\r\n');
      r.reject(new Error('cancelled'));
      return;
    }
    if (ch === '\x7f' || ch === '\b') {
      if (r.buffer.length) {
        const last = [...r.buffer].pop();
        r.buffer = r.buffer.slice(0, -last.length);
        if (r.echo) term.write('\b \b');
      }
      continue;
    }
    if (ch === '\x15') { // ctrl-u clears the line
      if (r.echo) term.write('\b \b'.repeat([...r.buffer].length));
      r.buffer = '';
      continue;
    }
    if (ch < ' ') continue;
    r.buffer += ch;
    if (r.echo) term.write(ch);
  }
}

function send(data) {
  skipTyping = true;
  if (lineReader) feedLine(data);
  else if (session && session.live) session.write(data);
  else if (onEnter && /[\r\n]/.test(data)) { const fn = onEnter; onEnter = null; fn(); }
}
term.onData(send);

document.querySelectorAll('#keys [data-key]').forEach((b) => {
  b.addEventListener('pointerdown', (e) => e.preventDefault()); // keep the keyboard up
  b.addEventListener('click', () => { send(b.dataset.key); term.focus(); });
});
screenEl.addEventListener('click', () => term.focus());

// ---------------------------------------------------------------- ssh core

// Start downloading the SSH core while the boot animation plays.
const coreResponse = fetch('killdeer-ssh.wasm');

function loadCore() {
  return new Promise((resolve, reject) => {
    if (typeof WebAssembly !== 'object' || typeof Go !== 'function') {
      reject(new Error('this browser cannot run WebAssembly'));
      return;
    }
    addEventListener('killdeerssh:ready', () => resolve(globalThis.killdeerSSH), { once: true });
    const go = new Go();
    coreResponse
      .then((response) => {
        if (!response.ok) throw new Error('the ssh core is unavailable right now');
        // Streaming compile needs application/wasm; fall back to buffering.
        return WebAssembly.instantiateStreaming(response.clone(), go.importObject)
          .catch(async () => WebAssembly.instantiate(await response.arrayBuffer(), go.importObject));
      })
      .then((result) => go.run(result.instance ?? result))
      .catch(reject);
  });
}

function relayURL() {
  const url = new URL('_ws/killdeer.digital', location.href);
  url.protocol = url.protocol === 'http:' ? 'ws:' : 'wss:';
  return url.href;
}

// ---------------------------------------------------------------- the show

async function boot() {
  applySettings();
  await document.fonts.load(`16px ${FONT}`).catch(() => {});
  term.open(termEl);
  // WebGL draws box-drawing and block characters itself, so guest consoles
  // line up even though Martian Mono has no such glyphs.
  try {
    const webgl = new WebglAddon();
    webgl.onContextLoss(() => webgl.dispose());
    term.loadAddon(webgl);
  } catch { /* fall back to the DOM renderer */ }
  layout();
  if (!reducedMotion) screenEl.classList.add('on');
  term.focus();
  await sleep(reducedMotion ? 0 : 650);

  for (const line of WORDMARK) {
    await type('  ' + line, 900);
    term.write('\r\n');
  }
  term.write('\r\n');
  await type(dim('  killdeer terminal · your ssh client runs in this tab') + '\r\n\r\n', 220);

  led('core', 'busy');
  await step('ssh core');
  let core;
  try {
    core = await loadCore();
  } catch (err) {
    stepFailed();
    led('core', 'bad');
    term.write(`\r\n  ${err.message}.\r\n  Use your own client instead: ${bold('ssh [username]@killdeer.digital')}\r\n`);
    return;
  }
  stepDone();
  led('core', 'ok');
  document.getElementById('fingerprint').textContent = 'host key ' + core.hostKeyFingerprint;
  term.write('\r\n');
  login(core);
}

async function login(core) {
  skipTyping = reducedMotion;
  resetLeds();
  let username = '';
  while (!username) {
    try {
      username = (await readLine(`  ${bold('username')}: `, true)).trim();
    } catch {
      username = '';
    }
    if (/[\s@]/.test(username)) {
      term.write(dim('  just the name, without spaces or @killdeer.digital') + '\r\n');
      username = '';
    }
  }
  term.write(dim('  new here? type the username you want and follow the prompts.') + '\r\n\r\n');

  let signedIn = false;
  session = core.connect({
    url: relayURL(),
    username,
    cols: COLS,
    rows: term.rows,
    onStatus(stage, detail) {
      inOrder(() => status(stage, detail));
    },
    onPrompt(question, echo) {
      return inOrder(() => readLine('  ' + question, echo));
    },
    onOutput(bytes) {
      inOrder(() => term.write(bytes));
    },
    onClose(message) {
      inOrder(() => closed(message));
    },
  });
  session.live = false;

  async function status(stage, detail) {
    switch (stage) {
      case 'relay':
        led('relay', 'busy');
        await step('relay');
        break;
      case 'relay-ok':
        stepDone(); led('relay', 'ok'); led('key', 'busy');
        await step('host key');
        break;
      case 'hostkey-ok':
        stepDone('pinned');
        led('key', 'ok'); led('auth', 'busy');
        term.write('\r\n');
        break;
      case 'auth-ok':
        signedIn = true;
        led('auth', 'ok');
        announce('Signed in to Killdeer');
        break;
      case 'shell':
        if (session) session.live = true;
        break;
    }
  }

  function closed(message) {
    const wasSignedIn = signedIn;
    session = null;
    lineReader = null;
    if (pendingStep) stepFailed();
    if (!wasSignedIn) {
      led('relay', document.getElementById('led-relay').dataset.state === 'ok' ? 'ok' : 'bad');
      led('auth', 'bad');
    } else {
      resetLeds();
    }
    const reason = message ? ': ' + message.replace(/^ssh: handshake failed: /, '') : '';
    term.write(`\r\n\r\n  ${ESC}7m connection closed${reason} ${ESC}27m\r\n  ${dim('press enter to sign in again')}\r\n`);
    announce('Connection closed' + reason);
    onEnter = () => { term.write('\r\n'); login(core); };
  }
}

addEventListener('pagehide', () => session?.close());
boot();
