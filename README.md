# PortMyPack — Web

A small web front end for [portmypack](https://github.com/restartfu/portmypack): drag in a Java Edition
Minecraft resource pack (`.zip`), get back a converted Bedrock Edition pack (`.mcpack`).

- `index.html` — static upload page (no build step, no framework)
- `api/convert.go` — a Go serverless function (Vercel's Go runtime) that runs the actual conversion
- `portmypack/` — the original conversion library, vendored in as-is (with one small robustness fix:
  it no longer panics on packs that have no sky/cubemap textures)

## Deploy to Vercel

**Option A — Vercel CLI (fastest)**
```bash
npm i -g vercel      # if you don't have it
cd portmypack-web
vercel               # first deploy, follow the prompts
vercel --prod        # promote to production
```

**Option B — GitHub + Vercel dashboard**
1. Push this folder to a new GitHub repo.
2. Go to https://vercel.com/new and import that repo.
3. Vercel auto-detects the Go function in `/api` (via the root `go.mod`) and serves `index.html`
   as a static file — no build settings needed. Click Deploy.

No environment variables or extra configuration are required.

## How it works

1. The browser POSTs the uploaded `.zip` to `/api/convert` as `multipart/form-data` (field name `pack`).
2. The Go function saves it to `/tmp` (Vercel functions get up to 500MB of writable `/tmp` scratch space),
   runs `java.NewResourcePack` → `portmypack.PortJavaEditionPack`, and streams the resulting `.mcpack`
   back as the response body with a `Content-Disposition: attachment` header.
3. The browser triggers a normal file download. Nothing is persisted between requests.

## Local development

```bash
npm i -g vercel
vercel dev
```
This runs both the static page and the Go function locally, matching production behavior.

## Known limitations (inherited from the original CLI)

- Expects the Java pack to have a top-level `textures/` folder inside the zip (standard Java pack layout).
- Very large packs may approach the default serverless execution time limit (10s on Vercel's Hobby plan).
  If you hit timeouts on big packs, raise `maxDuration` for `api/convert.go` in `vercel.json`
  (requires a Pro plan for values above 10s).
