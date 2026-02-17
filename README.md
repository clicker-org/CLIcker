<div align="center">
<pre>
   ✦  ·    ·      ✦    .      ✦      ·    .     ✦   ·
       .        ✦          ·       ✦      .       ✦

  ██████╗██╗     ██╗ ██████╗██╗  ██╗███████╗██████╗
 ██╔════╝██║     ██║██╔════╝██║ ██╔╝██╔════╝██╔══██╗
 ██║     ██║     ██║██║     █████╔╝ █████╗  ██████╔╝
 ██║     ██║     ██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗
 ╚██████╗███████╗██║╚██████╗██║  ██╗███████╗██║  ██╗
  ╚═════╝╚══════╝╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝

   ·     ✦    .    ·     ✦     .      ·    ✦    .  ·
</pre>

*a clicker game. in your terminal. across the cosmos.*

</div>

---

You already know what a clicker game is. CLIcker is that — numbers going up, things to buy, the slow burn of watching your idle income tick higher — except it lives entirely in your terminal, runs on `space`, and looks like it actually belongs there.

Seven worlds. Each one has its own currency, its own passive income ecosystem, its own ambient animation playing in the background (bubbles rising in the deep-ocean world, embers floating up in the volcanic one). Navigate between them on a galaxy map. Grind coins. Buy upgrades. Prestige when you're ready to trade everything you've built for a permanent multiplier, then do it all faster the second time. There's a hidden eighth world too, but figuring out how to get there is part of the game.

The whole thing is keyboard-driven. `space` to click, vim motions or arrows to navigate, first letters to jump between tabs. No mouse. 


## playing

```
$ clicker
```

Galaxy map. Pick a world, press `Enter`.

```
 [C]lick  [S]hop  [P]restige  [A]chievements
 ─────────────────────────────────────────────

    ·  ✦  ·   ·    ✦   ·  ✦   ·  ✦   ·

        ┌─────────────────────────┐
        │      PRESS SPACEBAR     │
        └─────────────────────────┘
                  +12 TC

    Click Power: 1.5 TC     CPS: 12.5

    ·   ✦   ·  ✦   ·    ·   ✦  ·   ·  ✦

 ─────────────────────────────────────────────
 Terra │ TC: 1,234K │ CPS: 12.5 │ Prestige: 2
 Progress: ████████░░░░  67%  │  LVL: 7
```

Hit `S` to open the shop. Spend your world coins on passive income generators that keep earning while you're clicking, or not clicking. Hit `P` when you've hit the prestige threshold — hard reset on the world, permanent multiplier and a chunk of **General Coins** in your pocket. Each world has its own tab layout, its own shop, its own things to unlock. The bottom bar never goes away.

## under the hood

General Coins are the cross-world meta-currency and the most interesting design decision in the game. You earn them by prestiging worlds, but also through **Exchange Boosts** — a softer mechanism where you sacrifice a portion of your current balance for GC without fully resetting, trading a smaller reward for zero risk. Over time, boosts also improve your exchange rate, so the two systems feed each other.

Close the game while grinding Terra and it'll keep a fraction of your CPS running. Come back an hour later and there's a report waiting for you: time offline, coins earned, whether you were in a world or idling in the galaxy map (overview generates a small trickle of GC directly). The cap on offline time is upgradeable with General Coins, per world.

Your **account level** sits above all of this. It accumulates from achievements and prestige and never resets — not even on global prestige. It quietly gates some of the best buy-ons across all worlds, which means sometimes the fastest path forward in World 1 is to go play World 4 for a while and come back leveled up. That's intentional. Achievements count toward the global completion percentage alongside per-world progress, so ignoring them isn't really an option.

Global completion runs from 0% to 99.99%. The last fraction of a percent is locked. How to get there is left as an exercise for the player.


> **disclaimer; early development**: architecture and planning phase, nothing is playable yet. the codebase is still being built with extensibility as the main constraint: adding a new world, achievement, or upgrade should never require touching the engine.

