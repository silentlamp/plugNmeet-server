# ZenLeader fork notes (zenleader-meet-server)

This repository is the **silentlamp** ZenLeader Meet server fork (upstream Plug-N-Meet).  
Former name: `silentlamp/plugNmeet-server` (GitHub redirects).

## Branch & CI/CD

| Item | Value |
|------|--------|
| Working / deploy branch | **`zenleader/dev`** |
| Remote | `origin` → `https://github.com/silentlamp/zenleader-meet-server.git` |
| CI workflow | `.github/workflows/deploy-zenleader.yml` |
| Trigger | `push` to **`zenleader/dev`** (or `workflow_dispatch`) |
| Image tag | `plugnmeet-server:zenleader` → VPS `/opt/plugNmeet` (Docker image name unchanged) |

Do **not** merge ZenLeader production work only to `main` — that branch does not run the ZenLeader deploy workflow.

```bash
git fetch origin
git checkout zenleader/dev
git pull origin zenleader/dev
```

Ops source of truth: [zen-leader-deploy/AGENTS.md](https://github.com/MiraiMagicLab/zen-leader-deploy/blob/main/AGENTS.md).
