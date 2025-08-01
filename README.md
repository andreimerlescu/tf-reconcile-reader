# Terraform Reconcile Reader

This package is both interactive and non-interactive, and its designed to allow you to interface with the rendered
`report.<env>.json` from the [reconcile-tfstate](https://github.com/andreimerlescu/reconcile-tfstate) repo. 

Added additional vim keys including:

## Installation

Since this is a private repository, you're required to use the `GONOSUMDB`, `GONOPROXY`, and `GOPRIVATE` environment
variables, set to `github.com/andreimerlescu/*`, but only when installing private resources; thus why you don't want this
set all the time. Just copy the single line. It's the easiest.

```bash
GONOSUMDB="github.com/andreimerlescu/*" \
GONOPROXY="github.com/andreimerlescu/*" \
GOPRIVATE="github.com/andreimerlescu/*" \
go install github.com/andreimerlescu/tf-reconcile-reader@latest
```

## Usage

```bash
andrei@localhost.dev:~/Downloads|⇒  tf-reconcile-reader -v
v1.0.0

andrei@localhost.dev:~/Downloads|⇒  tf-reconcile-reader -h
Usage of tf-reconcile-reader (powered by figtree v2.0.14):
-c|-contains [String] Substring search of executed command to select when using -non-interactive
-i|-input[=input.json] [String] Path to input file (.json)
-j|-json[=false] [Bool] format output as JSON
-non-interactive[=false] [Bool] enable non-interactive mode (auto responds with no)
-v[=false] [Bool] print version
-vi[=true] [Bool] enable vim-style keybindings (h to go back, j, k to scroll)

2025/07/24 12:00:52 failed to load flags: flag: help requested
```

## List View

* `s` or `l` or `enter` to Select a Command in List View to Open Detail View
* `j` or `↓` or mouse wheel / trackpad to arrow to scroll down
* `k` or `↑` or mouse wheel / trackpad to arrow to scroll up
* `q` or `esc` or `ctrl+c` to Exit Reader Application
* You can now loop through the list using the navigation where if you go back on the first, you get to the last item etc.

## Detail View

* `h` or `q` or `esc` to Go Back to List View
* `j` or `↓` or mouse wheel / trackpad to arrow to scroll down
* `k` or `↑` or mouse wheel / trackpad to arrow to scroll up
* `c` to copy the command being viewed to the clipboard
* `b` to scroll to the bottom of the window
* `t` or `n` to scroll to the top of the window

Using `VIM` mode requires an **ENV** to be set in either your `~/.bashrc` or ~/.zshrc` files respectively such that you have:

```bash
export FIGS_VIM_MODE=true
```

Then whenever the binary launches, `-vi` will be tagged onto the command automatically and you'll be able to use keyboard commands as you would expect in `vi`, but in a limited manner as defined by the above keymap. 
