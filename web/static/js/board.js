import * as THREE from 'three';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

const TILE = 1;
const GAP = 0.02;
const colors = {
  field: 0x5a8f4e,
  road: 0xc4a574,
  city: 0x8b4a3a,
  cityWall: 0x6b3a2e,
  monastery: 0xd4c4a0,
  base: 0x3d4a32,
  highlight: 0xf0d78c,
  red: 0xb83a2e,
  blue: 0x2f5f8a,
  green: 0x3a7a48,
  yellow: 0xc9a227,
  black: 0x2a2622,
};

let state = null;
let selectedRot = 0;
let hoverSpot = null;

const root = document.getElementById('board-root');
const scene = new THREE.Scene();
scene.background = new THREE.Color(0x1a241c);
scene.fog = new THREE.Fog(0x1a241c, 18, 42);

const camera = new THREE.PerspectiveCamera(42, 1, 0.1, 100);
camera.position.set(6, 10, 8);

const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setPixelRatio(Math.min(devicePixelRatio, 2));
root.appendChild(renderer.domElement);

const controls = new OrbitControls(camera, renderer.domElement);
controls.enableDamping = true;
controls.maxPolarAngle = Math.PI * 0.45;
controls.minDistance = 4;
controls.maxDistance = 28;
controls.target.set(0, 0, 0);

const hemi = new THREE.HemisphereLight(0xe8dcc4, 0x24362a, 1.1);
scene.add(hemi);
const sun = new THREE.DirectionalLight(0xfff2d6, 1.2);
sun.position.set(8, 14, 6);
scene.add(sun);

const boardGroup = new THREE.Group();
scene.add(boardGroup);
const ghostGroup = new THREE.Group();
scene.add(ghostGroup);
const MeepleGroup = new THREE.Group();
scene.add(MeepleGroup);

const ground = new THREE.Mesh(
  new THREE.CircleGeometry(40, 64),
  new THREE.MeshStandardMaterial({ color: 0x24362a, roughness: 0.95, metalness: 0 })
);
ground.rotation.x = -Math.PI / 2;
ground.position.y = -0.02;
ground.receiveShadow = true;
scene.add(ground);

const raycaster = new THREE.Raycaster();
const pointer = new THREE.Vector2();
const spotMeshes = [];

function resize() {
  const w = root.clientWidth;
  const h = root.clientHeight;
  camera.aspect = w / Math.max(h, 1);
  camera.updateProjectionMatrix();
  renderer.setSize(w, h, false);
}
window.addEventListener('resize', resize);
resize();

function edgeOf(def, rot, dir) {
  const src = (dir - rot + 4) % 4;
  return def.edges[src];
}

function hexCss(n) {
  return `#${n.toString(16).padStart(6, '0')}`;
}

const faceCache = new Map();

function paintTileFace(def, rot) {
  const key = `${def.id}:${rot}`;
  if (faceCache.has(key)) return faceCache.get(key);

  const S = 256;
  const c = document.createElement('canvas');
  c.width = S;
  c.height = S;
  const ctx = c.getContext('2d');

  ctx.fillStyle = hexCss(colors.field);
  ctx.fillRect(0, 0, S, S);

  const band = S * 0.28;
  const roadW = S * 0.18;
  const edges = [0, 1, 2, 3].map((d) => edgeOf(def, rot, d));

  const cityPath = (dir) => {
    ctx.beginPath();
    if (dir === 0) {
      ctx.moveTo(0, 0);
      ctx.lineTo(S, 0);
      ctx.lineTo(S - band, band);
      ctx.lineTo(band, band);
    } else if (dir === 1) {
      ctx.moveTo(S, 0);
      ctx.lineTo(S, S);
      ctx.lineTo(S - band, S - band);
      ctx.lineTo(S - band, band);
    } else if (dir === 2) {
      ctx.moveTo(0, S);
      ctx.lineTo(S, S);
      ctx.lineTo(S - band, S - band);
      ctx.lineTo(band, S - band);
    } else {
      ctx.moveTo(0, 0);
      ctx.lineTo(0, S);
      ctx.lineTo(band, S - band);
      ctx.lineTo(band, band);
    }
    ctx.closePath();
  };

  for (let d = 0; d < 4; d++) {
    if (edges[d] !== 2) continue;
    cityPath(d);
    ctx.fillStyle = hexCss(colors.city);
    ctx.fill();
    ctx.strokeStyle = hexCss(colors.cityWall);
    ctx.lineWidth = 3;
    ctx.stroke();
  }

  const mid = S / 2;
  const roadTo = (dir, inward) => {
    ctx.fillStyle = hexCss(colors.road);
    if (dir === 0) ctx.fillRect(mid - roadW / 2, 0, roadW, inward);
    if (dir === 1) ctx.fillRect(S - inward, mid - roadW / 2, inward, roadW);
    if (dir === 2) ctx.fillRect(mid - roadW / 2, S - inward, roadW, inward);
    if (dir === 3) ctx.fillRect(0, mid - roadW / 2, inward, roadW);
  };

  const roadDirs = [];
  for (let d = 0; d < 4; d++) if (edges[d] === 1) roadDirs.push(d);

  if (roadDirs.length === 2 && (roadDirs[0] + 2) % 4 === roadDirs[1]) {
    ctx.fillStyle = hexCss(colors.road);
    if (roadDirs[0] % 2 === 0) ctx.fillRect(mid - roadW / 2, 0, roadW, S);
    else ctx.fillRect(0, mid - roadW / 2, S, roadW);
  } else if (roadDirs.length === 2) {
    for (const d of roadDirs) roadTo(d, mid + roadW / 2);
    ctx.beginPath();
    ctx.arc(mid, mid, roadW * 0.7, 0, Math.PI * 2);
    ctx.fillStyle = hexCss(colors.road);
    ctx.fill();
  } else {
    for (const d of roadDirs) roadTo(d, mid + roadW / 2);
    if (roadDirs.length >= 3) {
      ctx.beginPath();
      ctx.arc(mid, mid, roadW * 0.85, 0, Math.PI * 2);
      ctx.fillStyle = '#b08968';
      ctx.fill();
    }
  }

  if (def.monastery) {
    ctx.fillStyle = hexCss(colors.monastery);
    ctx.fillRect(mid - 28, mid - 28, 56, 56);
    ctx.fillStyle = '#6b4a3a';
    ctx.beginPath();
    ctx.moveTo(mid, mid - 48);
    ctx.lineTo(mid + 32, mid - 16);
    ctx.lineTo(mid - 32, mid - 16);
    ctx.closePath();
    ctx.fill();
    ctx.fillStyle = '#3d2a20';
    ctx.fillRect(mid - 8, mid + 4, 16, 24);
  }

  if (def.shields > 0) {
    ctx.fillStyle = '#d4af37';
    ctx.beginPath();
    ctx.moveTo(S * 0.72, S * 0.18);
    ctx.lineTo(S * 0.82, S * 0.22);
    ctx.lineTo(S * 0.82, S * 0.34);
    ctx.lineTo(S * 0.72, S * 0.4);
    ctx.lineTo(S * 0.62, S * 0.34);
    ctx.lineTo(S * 0.62, S * 0.22);
    ctx.closePath();
    ctx.fill();
  }

  ctx.strokeStyle = 'rgba(28,24,18,0.35)';
  ctx.lineWidth = 6;
  ctx.strokeRect(3, 3, S - 6, S - 6);

  const tex = new THREE.CanvasTexture(c);
  tex.colorSpace = THREE.SRGBColorSpace;
  tex.anisotropy = Math.min(8, renderer.capabilities.getMaxAnisotropy?.() || 1);
  faceCache.set(key, tex);
  return tex;
}

function makeTileMesh(def, rot) {
  const g = new THREE.Group();
  const size = TILE - GAP;
  const base = new THREE.Mesh(
    new THREE.BoxGeometry(size, 0.1, size),
    new THREE.MeshStandardMaterial({ color: colors.base, roughness: 0.9 })
  );
  base.position.y = 0.05;
  g.add(base);

  const face = new THREE.Mesh(
    new THREE.PlaneGeometry(size * 0.98, size * 0.98),
    new THREE.MeshStandardMaterial({
      map: paintTileFace(def, rot),
      roughness: 0.85,
      metalness: 0,
    })
  );
  face.rotation.x = -Math.PI / 2;
  face.position.y = 0.101;
  g.add(face);
  return g;
}

function clearGroup(g) {
  while (g.children.length) {
    const c = g.children.pop();
    c.traverse((o) => {
      if (o.geometry) o.geometry.dispose();
      if (o.material) {
        if (Array.isArray(o.material)) o.material.forEach((m) => m.dispose());
        else o.material.dispose();
      }
    });
  }
}

function worldPos(x, y) {
  return new THREE.Vector3(x * TILE, 0, y * TILE);
}

function rebuildBoard() {
  clearGroup(boardGroup);
  clearGroup(ghostGroup);
  clearGroup(MeepleGroup);
  spotMeshes.length = 0;
  if (!state) return;
  const catalog = state.tileCatalog || {};
  for (const t of state.board || []) {
    const def = catalog[t.defId];
    if (!def) continue;
    const mesh = makeTileMesh(def, t.rotation);
    const p = worldPos(t.x, t.y);
    mesh.position.copy(p);
    boardGroup.add(mesh);
  }
  for (const m of state.meeples || []) {
    const player = (state.players || []).find((p) => p.id === m.playerId);
    const color = colors[player?.color] || 0xffffff;
    const body = new THREE.Mesh(
      new THREE.CylinderGeometry(0.08, 0.1, 0.28, 8),
      new THREE.MeshStandardMaterial({ color, roughness: 0.55 })
    );
    const p = worldPos(m.x, m.y);
    const ox = (m.slot % 3) * 0.15 - 0.15;
    const oz = Math.floor(m.slot / 3) * 0.15 - 0.1;
    body.position.set(p.x + ox, 0.28, p.z + oz);
    MeepleGroup.add(body);
  }
  if (state.phase === 'placeTile' && state.validSpots) {
    for (const spot of state.validSpots) {
      const mat = new THREE.MeshStandardMaterial({
        color: 0xf5d76e,
        transparent: true,
        opacity: 0.55,
        roughness: 1,
        emissive: 0xc9a227,
        emissiveIntensity: 0.35,
      });
      const mesh = new THREE.Mesh(new THREE.BoxGeometry(TILE * 0.92, 0.04, TILE * 0.92), mat);
      const p = worldPos(spot.x, spot.y);
      mesh.position.set(p.x, 0.02, p.z);
      mesh.userData.spot = spot;
      ghostGroup.add(mesh);
      spotMeshes.push(mesh);
    }
  }
  updatePreview();
}

let previewTile = null;

function syncRotationToSpot(spot) {
  if (!spot?.rots?.length) return;
  if (!spot.rots.includes(selectedRot)) {
    selectedRot = spot.rots[0];
  }
}

function updatePreview() {
  if (previewTile) {
    ghostGroup.remove(previewTile);
    previewTile.traverse((o) => {
      if (o.geometry) o.geometry.dispose();
      if (o.material) {
        if (Array.isArray(o.material)) o.material.forEach((m) => m.dispose());
        else o.material.dispose();
      }
    });
    previewTile = null;
  }
  if (!state?.currentTile || !state.tileCatalog?.[state.currentTile] || !hoverSpot) return;
  syncRotationToSpot(hoverSpot);
  previewTile = makeTileMesh(state.tileCatalog[state.currentTile], selectedRot);
  previewTile.position.copy(worldPos(hoverSpot.x, hoverSpot.y));
  previewTile.traverse((o) => {
    if (o.material) {
      o.material = o.material.clone();
      o.material.transparent = true;
      o.material.opacity = 0.72;
    }
  });
  ghostGroup.add(previewTile);
}

function isYourTurn() {
  if (!state?.players?.length) return false;
  const cur = state.players[state.turn];
  return !!cur && cur.id === state.youId;
}

async function placeAt(spot) {
  if (!state || state.phase !== 'placeTile' || !isYourTurn()) return;
  syncRotationToSpot(spot);
  const code = state.roomCode;
  const signals = JSON.stringify({ x: spot.x, y: spot.y, rot: selectedRot });
  await fetch(`/api/rooms/${code}/place-tile`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'text/event-stream' },
    body: signals,
  });
}

function onPointer(ev) {
  const rect = renderer.domElement.getBoundingClientRect();
  pointer.x = ((ev.clientX - rect.left) / rect.width) * 2 - 1;
  pointer.y = -((ev.clientY - rect.top) / rect.height) * 2 + 1;
  raycaster.setFromCamera(pointer, camera);
  const hits = raycaster.intersectObjects(spotMeshes, false);
  if (hits.length) {
    hoverSpot = hits[0].object.userData.spot;
    document.body.style.cursor = 'pointer';
  } else {
    hoverSpot = null;
    document.body.style.cursor = '';
  }
  updatePreview();
}

renderer.domElement.addEventListener('pointermove', onPointer);
renderer.domElement.addEventListener('click', () => {
  if (hoverSpot) placeAt(hoverSpot);
});

window.addEventListener('keydown', (e) => {
  if (e.key === 'r' || e.key === 'R') {
    if (!hoverSpot?.rots?.length) return;
    const rots = hoverSpot.rots;
    const i = rots.indexOf(selectedRot);
    selectedRot = rots[(i + 1) % rots.length];
    updatePreview();
  }
});

document.addEventListener('carcassonne:state', (e) => {
  applyState(e.detail);
});

function applyState(next) {
  if (!next) return;
  state = next;
  selectedRot = 0;
  resize();
  rebuildBoard();
}

const boot = document.getElementById('initial-state');
if (boot?.textContent) {
  try {
    applyState(JSON.parse(boot.textContent));
  } catch (_) {}
}
if (window.__carcassonneLatest) {
  applyState(window.__carcassonneLatest);
}

const roomCode = location.pathname.split('/').pop();
if (roomCode) {
  fetch(`/api/rooms/${roomCode}/state`, { credentials: 'same-origin' })
    .then((r) => (r.ok ? r.json() : null))
    .then(applyState)
    .catch(() => {});
}

function animate() {
  requestAnimationFrame(animate);
  controls.update();
  const t = performance.now() * 0.001;
  for (const m of spotMeshes) {
    m.material.opacity = 0.4 + Math.sin(t * 3 + m.position.x) * 0.15;
  }
  renderer.render(scene, camera);
}
animate();
