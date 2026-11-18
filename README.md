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

A small, fast todo TUI: keep several lists, add items, set priorities, check them off, and reorder them, all from the keyboard.

```text
┌───────────────────────────────────────────────────────────────────────────┐
│ [ Work ]  2/4 done  Home  Groceries                                       │
│                                                                           │
│ ▶[ ] fix login bug        (high, red)                                     │
│  [ ] write docs           (medium, yellow)                                │
│  [ ] call mom                                                             │
│  [x] tidy desk            (low, blue)                                     │
│                                                                           │
│ a:add  space:toggle  1/2/3/0:priority  s:sort  J/K:move  d:delete  q:quit │
│ tab/shift+tab:switch list  n:new list  r:rename list  D:delete list       │
└───────────────────────────────────────────────────────────────────────────┘
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

All lists are saved to `~/.local/share/todoui/todos.json` (respects `XDG_DATA_HOME`).
Set `TODOUI_FILE` to use a different file. A file from v0.1.0 (a single list) is
upgraded automatically and opens as a list named "Todo".

## Keys

| Key                      | Action                                                 |
| ------------------------ | ------------------------------------------------------ |
| **Items**                |                                                        |
| `a`                      | Add an item below the cursor                           |
| `space` / `x` / `enter`  | Toggle the item complete / not complete                |
| `j` / `k` / arrows       | Move the cursor                                        |
| `1` / `2` / `3`          | Set priority high / medium / low (red / yellow / blue) |
| `0`                      | Clear priority                                         |
| `s`                      | Sort by priority (high → low → none, completed last)   |
| `J` / `K` / shift+arrows | Move the item down / up                                |
| `g` / `G`                | Jump to top / bottom                                   |
| `d`                      | Delete the item                                        |
| **Lists**                |                                                        |
| `tab` / `l` / →          | Next list                                              |
| `shift+tab` / `h` / ←    | Previous list                                          |
| `n`                      | New list                                               |
| `r`                      | Rename the current list                                |
| `D`                      | Delete the current list (asks to confirm)              |
| `q`                      | Quit                                                   |

When adding, `enter` saves and `esc` cancels.

Sorting is stable: items with the same priority keep the order you gave them,
and you can still move items with `J` / `K` afterwards.

Each list remembers its own cursor while the app is open, and todoui reopens on
the list you were last viewing. The last remaining list can't be deleted.

See [TODO.md](TODO.md) for planned features.
