# ZenLeader fork notes (plugNmeet-server)

This repository is the **silentlamp** fork of upstream Plug-N-Meet server, customized for ZenLeader.

## Branch & CI/CD

| Item | Value |
|------|--------|
| Working / deploy branch | **`zenleader/dev`** |
| Remote | `origin` → `https://github.com/silentlamp/plugNmeet-server.git` |
| CI workflow | `.github/workflows/deploy-zenleader.yml` |
| Trigger | `push` to **`zenleader/dev`** (or `workflow_dispatch`) |
| Image tag | `:zenleader` → VPS `/opt/plugNmeet` |

Do **not** merge ZenLeader production work only to `main` — that branch does not run the ZenLeader deploy workflow.

```bash
git fetch origin
git checkout zenleader/dev
git pull origin zenleader/dev
```

Ops source of truth: [zen-leader-deploy/AGENTS.md](https://github.com/MiraiMagicLab/zen-leader-deploy/blob/main/AGENTS.md).

## ZenLeader create-room feature policy

PlugNMeet `PrepareDefaultRoomFeatures` uses proto3 `proto.Merge`, which **cannot** apply
`false` over default `true` for nested feature bools. Java `MeetGatewayClient` already sends
`whiteboardFeatures.isAllow=false` (and other disables), but without a post-merge fix the
server keeps PlugNMeet defaults (`whiteboard` on, chat file upload on, …).

`setRoomDefaults` therefore:

1. Snapshots the create request features
2. Calls `PrepareDefaultRoomFeatures`
3. Re-applies explicit bools via `applyExplicitCreateRoomFeatureBools`
4. Forces whiteboard off via `enforceZenLeaderCreateRoomPolicy`
