package handlers

const homeHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Carcassonne</title>
<link rel="preconnect" href="https://fonts.googleapis.com"/>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
<link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:wght@500;700&family=Source+Sans+3:wght@400;600&display=swap" rel="stylesheet"/>
<link rel="stylesheet" href="/static/css/app.css"/>
<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.3/bundles/datastar.js"></script>
</head>
<body class="home" data-signals="{
  playerName:'', name:'Carcassonne', maxPlayers:4, private:false, code:'',
  farmers:true, river:false, abbot:false, innsAndCathedrals:false,
  tradersAndBuilders:false, princessAndDragon:false, tower:false,
  abbeyAndMayor:false, bridgesCastlesBazaars:false, hillsAndSheep:false
}">
  <div class="home-bg"></div>
  <main class="home-main">
    <header class="brand">
      <p class="eyebrow">Multiplayer · Three.js · Datastar</p>
      <h1>Carcassonne</h1>
      <p class="lede">Build roads, claim cities, and score with friends on a living 3D board.</p>
    </header>
    <section class="home-actions">
      <label>Your name <input data-bind:player-name placeholder="Knight of the Midi"/></label>
      <div class="split">
        <div class="panel">
          <h2>Create table</h2>
          <label>Table name <input data-bind:name/></label>
          <label>Players
            <select data-bind:max-players>
              <option value="2">2</option>
              <option value="3">3</option>
              <option value="4" selected>4</option>
              <option value="5">5</option>
              <option value="6">6</option>
            </select>
          </label>
          <h3>Expansions</h3>
          <div class="exp-grid">
            <label class="check"><input type="checkbox" data-bind:farmers/> Farmers</label>
            <label class="check"><input type="checkbox" data-bind:river/> The River</label>
            <label class="check"><input type="checkbox" data-bind:abbot/> The Abbot</label>
            <label class="check"><input type="checkbox" data-bind:inns-and-cathedrals/> Inns &amp; Cathedrals</label>
            <label class="check"><input type="checkbox" data-bind:traders-and-builders/> Traders &amp; Builders</label>
            <label class="check"><input type="checkbox" data-bind:princess-and-dragon/> Princess &amp; Dragon</label>
            <label class="check"><input type="checkbox" data-bind:tower/> The Tower</label>
            <label class="check"><input type="checkbox" data-bind:abbey-and-mayor/> Abbey &amp; Mayor</label>
            <label class="check"><input type="checkbox" data-bind:bridges-castles-bazaars/> Bridges, Castles &amp; Bazaars</label>
            <label class="check"><input type="checkbox" data-bind:hills-and-sheep/> Hills &amp; Sheep</label>
          </div>
          <h3>Visibility</h3>
          <div class="mode-row">
            <button type="button" class="mode-opt" data-class:active="!$private" data-on:click="$private = false">Public</button>
            <button type="button" class="mode-opt" data-class:active="$private" data-on:click="$private = true">Private</button>
          </div>
          <p class="muted" data-show="!$private">Listed under Open tables — anyone can join.</p>
          <p class="muted" data-show="$private" style="display:none">Hidden from Open tables. Friends join with the room code only.</p>
          <button class="btn primary" data-on:click="@post('/api/rooms')">Create game</button>
        </div>
        <div class="panel">
          <h2>Join</h2>
          <label>Room code <input data-bind:code maxlength="5" placeholder="ABCDE" style="text-transform:uppercase"/></label>
          <button class="btn" data-on:click="@post('/api/join')">Join with code</button>
          <p class="muted">Private rooms never appear below — use the code your host shared.</p>
          <h3>Open tables</h3>
          <div id="room-list" data-init="@get('/api/rooms')"></div>
        </div>
      </div>
    </section>
  </main>
</body>
</html>`

const roomHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Carcassonne — {{CODE}}</title>
<link rel="preconnect" href="https://fonts.googleapis.com"/>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin/>
<link href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:wght@500;700&family=Source+Sans+3:wght@400;600&display=swap" rel="stylesheet"/>
<link rel="stylesheet" href="/static/css/app.css"/>
<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.3/bundles/datastar.js"></script>
<script type="importmap">
{"imports":{"three":"https://cdn.jsdelivr.net/npm/three@0.170.0/build/three.module.js","three/addons/":"https://cdn.jsdelivr.net/npm/three@0.170.0/examples/jsm/"}}
</script>
<script>
window.__carcassonneLatest = null;
document.addEventListener("carcassonne:state", function (e) {
  window.__carcassonneLatest = e.detail;
});
</script>
</head>
<body class="room" data-signals="{kind:'', slot:0, x:0, y:0, rot:0, abbot:false}" data-init="@get('/api/rooms/{{CODE}}/stream')">
  <div class="room-shell">
    <aside class="side">
      <a class="back" href="/">← Lobby</a>
      <h1 class="side-brand">Carcassonne</h1>
      <p class="side-sub">{{NAME}}</p>
      {{LOBBY}}
      {{HUD}}
    </aside>
    <div id="board-root" class="board-root"></div>
  </div>
  <script id="initial-state" type="application/json">{{STATE}}</script>
  <script type="module" src="/static/js/board.js"></script>
</body>
</html>`
