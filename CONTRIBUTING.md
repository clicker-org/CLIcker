# Contributing

Thanks for wanting to contribute. Here's how the process works.

## Before you write any code

Every change starts with a GitHub issue. If there's already an issue for what you want to do, great — leave a comment saying you're picking it up. If there isn't one, open it first and describe what you're planning. This avoids duplicate work and gives a place to discuss the approach before anyone writes code.

## The workflow

1. Fork the repo
2. Create a branch off of `develop` (not `main`) — name it whatever makes sense
3. Make your changes
4. Open a pull request back into `develop`

Don't PR into `main`. Main is only updated when a new version is released.

## Commit messages

Keep them short and reference the issue:

```
Add offline income calculation (#12)
Fix prestige multiplier stacking (#34)
Refactor world registry to use generics (#8)
```

The format is just `Action being done (#issue-number)`. Use the imperative tense. No need for a long body unless something actually needs explaining.

## CI

When you open a PR into `develop`, the pipeline runs the test suite. It needs to pass before merging. If something's failing on your branch, fix it before asking for review.

`develop` is the integration branch — it's where things land and get tested together. Once things are stable there, it gets merged into `main` and tagged as a new release following `vX.Y.Z` (semver). That part is automated.

## Joining the org

If you've made a few contributions and want to be more involved, you can request to join the [clicker-org](https://github.com/clicker-org) GitHub organization. Just ask in an issue or reach out directly.
