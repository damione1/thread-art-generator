# Redesign "atelier" — implementation contract

The visual specification is the set of static mockups in `.design/canvas/*.dc.html`.
Read the one named in your brief and port it to templ. They are plain HTML with an
inline `<style>` block; the numeric values in them (sizes, paddings, radii, opacities)
are the spec. Copy them exactly rather than rounding to a 4/8px grid.

Ignore in those files: the `<x-dc>` / `<helmet>` wrapper, the `support.js` script tag,
and the fact that data is hard-coded. In the real templates the data comes from
`*pb.Art`, `*pb.Composition`, `*templates.PageData`, etc.

## Foundations that already exist — use them, do not redefine them

`client/tailwind.config.js` (do NOT edit):

| token | value | meaning |
|---|---|---|
| `bg-ground` | #0A0A0B | page ground |
| `bg-surface` | #111113 | cards |
| `bg-surface-raised` | #17171A | hover / nested |
| `bg-surface-high` | #1D1D21 | slider track, avatar |
| `bg-surface-sunken` | #0C0C0E | inputs, disc background |
| `border-line` | #232327 | hairline, 1px, everywhere |
| `border-line-strong` | #33333A | button borders, dashed circles |
| `text-ink` | #F4F4F5 | primary text |
| `text-ink-muted` | #A2A2AA | secondary text |
| `text-ink-faint` | #6B6B74 | hints, meta |
| `brass` | #C79A4B | the single accent: the colour of the nails |
| `brass-hover` | #E0B463 | hover |
| `brass-ink` | #17130A | text on a brass fill |
| `thread` | #EDEDEE | the thread itself |
| `ok` / `danger` | #7FB98A / #D4756E | status only, never decoration |

Fonts: `font-serif` = Instrument Serif (display headings only), `font-sans` =
Instrument Sans (default, do not write the class), `font-mono` = JetBrains Mono
(every number the solver measures: nail counts, paths, millimetres, metres, dates
in meta lines, file sizes).

Component classes live in `client/styles/input.css` (do NOT edit):

- buttons: `btn` + `btn-primary` | `btn-ghost` | `btn-danger`, plus `btn-sm` | `btn-lg` | `btn-block`
- fields: `field` (column wrapper), `label`, `input`, `input-num`, `input-error`, `hint`, `msg-error`
- range: `slider` on `<input type="range">`
- surfaces: `card`, `hairline`
- status: `badge` + `badge-ok` | `badge-run` | `badge-error`, with a `<span class="badge-dot"></span>` inside
- messages: `note` + `note-ok` | `note-error`
- circles: `disc` (square container, `rounded-full`, `overflow-hidden`; put an `<img>` or `<svg>` inside)
- key/value: `kv` > `kv-k` + `kv-v`
- type: `kicker` (mono uppercase brass label), `display` (serif heading), `tabular`

Shared templ components in `client/internal/templates/components.templ` (do NOT edit):

- `@Icon(name, class)` — inline SVG, sized `1em` by default, override with `w-4 h-4` etc.
  Names: arrow-left, arrow-right, chevron-down, chevron-left, plus, close, search,
  upload, download, check-circle, alert-circle, info, file-text, image, mail, trash,
  copy, eye, zoom-in, zoom-out, lock, user, grid, sort, hoop, spinner.
- `@Spinner(class)`, `@Logo(withText)`, `@NailRing(class)`, `@ProgressRing(percent, class)`
- `@ErrorAlert([]string)`, `@SuccessAlert(string)`, `@InfoAlert(string)`
- `@Layout(data)` (app chrome: header + footer), `@BareLayout(data)` (no chrome, for auth),
  `@Shell()` (the centred 1320px column with the page gutters)

`@MaterialIcon` is a deprecated shim. Replace every call you touch with `@Icon`.

## Design rules

1. **The circle carries the interface.** Every piece of work is a disc, never a
   rounded rectangle: project thumbnails, previews, the crop mask, the upload target.
   Status is expressed on the ring around the disc — solid dotted brass ring = ready
   (`@NailRing`), arc = progress (`@ProgressRing`), dashed grey = no image yet.
2. Hairlines and flat surfaces. No gradients, no drop shadows except on overlays
   (modals, dropdowns), no `transform: translateY` hover lifts.
3. One accent. Brass is for the nails, the primary action and the kicker labels.
   Green and red appear only on status.
4. Buttons are pills (`rounded-full`), 40px high (32px for `btn-sm`).
5. Numbers are mono and tabular. Prose is sans. Page titles are serif.
6. Copy is in English, sentence case, specific. No exclamation marks, no marketing
   superlatives, no emoji.
7. Hit targets stay >= 44px on the mobile layouts.

## Ground rules for this task

- **Only touch the files listed in your brief.** Other agents are working in parallel
  on the rest of the templates. Never edit `tailwind.config.js`, `styles/input.css`,
  `layout.templ`, `components.templ` or anything under `core/`.
- **Do not change behaviour.** Keep every form `action`, `method`, `hx-*` attribute,
  input `name`, element `id`, and Alpine (`x-data`, `@click`, `x-show`, …) binding
  exactly as it is unless your brief says otherwise. JavaScript in
  `client/public/js/` and `client/src/` binds to those ids and names; breaking one
  breaks upload, polling or sign-in. This is a restyle, not a refactor.
- **Drop templUI.** Replace `github.com/axzilla/templui/component/*` usage
  (`button.Button`, `form.Item`, `input.Input`, `slider.*`, `alert.*`, `spinner.*`,
  `separator.*`) with plain markup using the classes above. Remove the now-unused
  imports. Do not remove the dependency from `go.mod`.
- Keep the Go helper functions that already live in the files you own
  (`extractArtID`, `extractUserID`, `extractCompositionID`, …).

## Build and verify

From the repository root:

```
go run github.com/a-h/templ/cmd/templ@v0.3.920 generate -f <path to each .templ you changed>
go build ./client/...
go test ./client/internal/templates/ ./client/internal/handlers/
```

Do NOT run `make generate-templ` — it regenerates every file and will collide with
the other agents. Generate only your own files with `-f`.

Rebuild the stylesheet only if you need to check a class exists:
`cd client && npx tailwindcss -i ./styles/input.css -o ./public/css/tailwind.css`

Report at the end: the files you changed, anything in the mockup you could not
reproduce and why, and any behaviour you had to touch.
