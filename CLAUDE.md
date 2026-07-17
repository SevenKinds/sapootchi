# Sapootchi — CLAUDE.md

Guidance for working in this repo. Keep it short; update it when decisions change.

## What this is

A tamagotchi-style virtual pet that acts as a **daily-habit engine for a brand
loyalty program**. It is not a game with loyalty bolted on — it's a habit loop
disguised as a pet. Brand actions (receipts, store visits, referrals) translate
into things the pet enjoys.

**North star:** every session must answer *"How did my pet change because I came
back today?"* If the only answer is "a bar went down", it failed. Session target
is under 2 minutes.

**Design test for any feature:** "does this make someone more loyal to the
brand?" — not "should this give points?"

Full design lives in `notes.txt` (living braindump; `[LOCKED]`/`[TODO]`/`[IDEA]`
tags). This file is the implementation-facing summary.

## Stack

- **Go + Ebitengine (Ebiten)** — 2D game, compiles to WASM (web), iOS, Android,
  desktop from one codebase. Chosen to stay in Go and reach web + mobile.
- No `go.mod` yet — greenfield. When scaffolding, init the module and pin the Go
  version to the installed toolchain.

## Architecture — the rules that matter

1. **`simulation` package has ZERO Ebiten imports.** Pet state is plain data;
   decay is a **pure function of elapsed wall-clock time** (not frames/sessions);
   rendering is a pure *view* of state. This is the single most valuable
   boundary — it keeps game logic unit-testable, server-portable (hard currency
   later), and renderer-agnostic. Do not leak `ebiten.*` into it.
2. **Scene manager:** a `Scene` interface (`Update`, `Draw`) with a stack.
   Scenes: Home, MiniGame, Shop, DressUp. Ebiten imposes no structure — this is
   ours to own.
3. **Mini-games are plug-ins** behind one interface:
   `Play() -> Result{Score, Coins, StatDelta, Items}`. A game may pay Items OR
   Coins — don't assume one currency per game. Ship **catch-food only** in POC.
4. **POC state is entirely local** (coins only; no brand points). Local clock =
   local cheat is a known, accepted POC limitation.

## Locked mechanics (POC)

- **Two currencies, strictly separated.** Coins (soft, client-side, generous).
  Brand Points (hard, scarce, server-authoritative) are **out of scope for POC**.
- **Four visible stats:** Happiness, Hunger, Hygiene, Energy. Wall-clock decay
  (except Energy — see below).
- **Hunger decays 25%/day** (full→empty in 4 days). Standard food = **+25%**
  hunger (4 items empty→full).
- **Energy is INVERTED:** it *regenerates* over wall-clock time (idle slowly,
  resting/asleep fast) and is *spent* by activity (mini-games). At **100% →
  Energized** (must play a mini-game to burn it off; idle only refills it). At
  **0% → falls asleep**, uninteractable, until Energy regenerates to **50%**
  (hysteresis: sleep at 0, wake at 50). `Pet.Energized()`, `Pet.Awake()`,
  `Pet.Asleep`.
- **Soft-fail:** Hunger hitting **0 → the pet runs away**. Recovery: prompt to
  **leave food out**; each following day there is a **28% chance** he returns.
  Recoverable, never permanent death.
- **Catch-food mini-game awards 1–3 food items** based on score (to inventory).
  Feeding consumes inventory items — it is not a direct hunger fill.
- **Egg phase is near-instant:** start → egg → hatches immediately. First real
  phase is the **baby**.
- **Personality** is assigned at hatch (random) for POC. Emergent personality is
  v2.
- **Mood** (Happy/Hungry/Excited/Bored/Sleepy/Curious/Lonely) is derived from
  stats + events and affects **only animation + dialogue**, never gameplay.
- **Dress-up** is its own Scene for skins/cosmetics.

## Assets

- `assets/blob.png` — the pet ("the blob"): green, big googly eyes. This is the
  **full/adult size**.
- **Baby = the same `blob.png` rendered at 60% scale.** No separate baby asset.
- POC art can be faked hard (shapes/emoji + squash-stretch tweens). The
  animation/mood system is the real cost and the source of the magic — budget for
  it; the sim itself is easy.

## POC scope

**In:** one pet · 4 stats + wall-clock decay + runs-away-at-0 · Home
(feed/bathe/rest + petting) · one mini-game (catch-food) · inventory + food ·
coins + small shop · DressUp scene · mood→animation · one needy moment · local
save · instant egg→baby + one evolution step (branching stubbed).

**Out (v2+):** social/NPC visits, achievements, home decor, daily quests,
weather, seasonal, brand points, full evolution tree.

## Layout

```
main.go              entry point (window + RunGame)
simulation/          ZERO ebiten imports — all game rules, unit-tested
  pet.go             types: Pet, Phase, Personality, Stats
  decay.go           wall-clock decay, runaway + 28%/day return
  food.go care.go    feeding, bathe/rest/pet, perfect-care
  mood.go            mood precedence (Hungry > Sleepy > Lonely > ...)
  evolution.go       age gate + stubbed branch (Cute/Smart/Wild)
  save.go            JSON marshal (I/O is the caller's job)
  pet_test.go        decay, feed, runaway (~28%), mood, save, evolve
minigame/            plug-in interface + Result; CatchFood, Runner
ui/                  shared look-and-feel: palette, fonts (Go typeface via
                     text/v2), buttons, stat bars, icons, gradient bg, and the
                     pointer/tap/swipe input tracker (ui/input.go)
game/                ebiten layer: Scene stack; MainScene = swipeable pager with
                     icon tab bar hosting the Pages; game catalog; save codec
assets/              embed blob.png
```

**Navigation:** `MainScene` hosts six `Page`s (Home, Play, Items, Shop, Dress,
Settings) as a horizontal pager — swipe the page area or tap an icon in the
single-row bottom tab bar. Mini-games are still pushed modally on the scene
stack. `ui.Button.Clicked()` is tap-based (press+release, small drag) so swipes
never fire buttons; `Game.Update` must call `ui.UpdateInput()` first.

Mini-games are a plug-in list in `game/games.go` (`gameCatalog`); factories get
`*Game` so they can honor `Settings.RealSpriteInGames` (real blob sprite as the
player character vs shape stand-ins — toggle in Settings). Four games ship:
Catch Food (Hunger via items, no energy), Runner (the energy-burner, coins by
distance), Scrub (Hygiene; rub dirt spots off), Simon (Happiness + hidden
Intelligence; the one game allowed real randomness — a fixed sequence would be
memorizable). `Result{Score, Coins, StatDelta, Hidden, Items}`; the shared
results card itemizes rewards, tap to dismiss. All shared visuals go through
the `ui` package.

**Energy economy (tuned):** idle regen 100%/day, sleep regen 480%/day
(nap 0→50 in ~2.5h); costs Runner 30 / Scrub 10 / Simon 8 / Catch Food 0.

**Save format:** `{Pet, Settings}` JSON (see `game/save.go`); legacy bare-Pet
saves still load via fallback.

Run: `go run .` (desktop; `-tags dev` adds the DEV section in Settings).
Web: `./web/serve.sh` (LAN testing) or `./web/deploy.sh` (Cloudflare Pages,
needs `wrangler login`). Pages has a 25 MiB per-file limit — animations are
NOT embedded on js builds (see assets/anims_*.go + game/anims_js.go: fetched
at runtime with browser fetch; net/http would cost ~8 MiB of WASM).
Test the rules: `go test ./simulation/`.

**Resolution model:** scenes author coordinates + font sizes in a fixed 360×640
design space. `ui.Scale` (default 2) multiplies every `ui` draw call, so the
framebuffer renders at 720×1280 (crisp, not an upscale). `game.Layout` returns
`360×640 * ui.Scale`; the window (`main.go`) is a comfortable resizable size
Ebiten downscales into. Input coords are divided back by `ui.Scale`
(`ui.Button`, catch-food cursor). To change render density, change `ui.Scale`
only — scene code is untouched.

## Status

Scaffold DONE and verified (builds desktop + WASM; sim tests pass; window runs;
save/decay confirmed). Working: hatch→baby, 4-stat wall-clock decay, runaway at
hunger 0 + leave-food-out recovery, Home care (feed/bathe/rest/pet), catch-food
mini-game → food items, coins + shop, DressUp stub, local save/load.

UI polished: shared `ui` package with the Go typeface (anti-aliased via
text/v2), cohesive palette, rounded hover buttons, rounded stat bars, gradient
background, pet drop-shadow. Two mini-games (catch-food, runner) behind the
plug-in interface + a picker menu.

Next likely work: real art/animation + mood-driven idle cycles (the actual cost
center), tune non-hunger decay rates, needy moments, evolution branch reveal,
Scrub (Hygiene) mini-game.
