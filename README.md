<pre align="center">
████████╗ ██████╗ ██████╗  ██████╗ ██╗   ██╗██╗
╚══██╔══╝██╔═══██╗██╔══██╗██╔═══██╗██║   ██║██║
   ██║   ██║   ██║██║  ██║██║   ██║██║   ██║██║
   ██║   ██║   ██║██║  ██║██║   ██║██║   ██║██║
   ██║   ╚██████╔╝██████╔╝╚██████╔╝╚██████╔╝██║
   ╚═╝    ╚═════╝ ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝

  ✓ a simple todo list for your terminal ✓
</pre>

# todoui

A small, fast todo TUI: add items, check them off, and reorder them, all from the keyboard.

```text
┌─────────────────────────────────────────────────────────────┐
│ todoui  (1/3 done)                                          │
│                                                             │
│ ▶[ ] write code                                             │
│  [x] buy milk                                               │
│  [ ] walk dog                                               │
│                                                             │
│ a:add  space:complete  J/K:move  d:delete  j/k:nav  q:quit  │
└─────────────────────────────────────────────────────────────┘
```

## Install

Download a prebuilt binary, `.deb`, or `.rpm` for Linux or macOS from the
[Releases page](https://github.com/skyler-v1dot0/todoui/releases).

> **macOS:** the release binaries aren't signed, so macOS may block `todoui`
> the first time you run it ("cannot be opened because the developer cannot be
> verified"). After extracting the download, remove the quarantine flag:
>
> ```sh
> xattr -d com.apple.quarantine todoui
> ```

### With Go

Requires [Go](https://go.dev/dl/) 1.26 or newer.

```sh
go install github.com/skyler-v1dot0/todoui@latest
```

This puts the `todoui` binary in `~/go/bin` (make sure that's on your `PATH`).

## Build from source

```sh
git clone https://github.com/skyler-v1dot0/todoui.git
cd todoui
go build
./todoui
```

Items are saved to `~/.local/share/todoui/todos.json` (respects `XDG_DATA_HOME`).
Set `TODOUI_FILE` to use a different file.

## Keys

| Key                      | Action                       |
| ------------------------ | ---------------------------- |
| `a`                      | Add an item below the cursor |
| `space` / `x` / `enter`  | Mark the item complete       |
| `j` / `k` / arrows       | Move the cursor              |
| `J` / `K` / shift+arrows | Move the item down / up      |
| `g` / `G`                | Jump to top / bottom         |
| `d`                      | Delete the item              |
| `q`                      | Quit                         |

When adding, `enter` saves and `esc` cancels.

See [TODO.md](TODO.md) for planned features.
