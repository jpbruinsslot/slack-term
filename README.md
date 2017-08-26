Slack-Term
==========

A [Slack](https://slack.com) client for your terminal.

![Screenshot](/screenshot.png?raw=true)

Installation
------------

#### Binary installation

[Download](https://github.com/erroneousboat/slack-term/releases) a
compatible binary for your system. For convenience place `slack-term` in a
directory where you can access it from the command line. Usually this is
`/usr/local/bin`.

```bash
$ mv slack-term /usr/local/bin
```

#### Via Go

If you want you can also get `slack-term` via Go:

```bash
$ go get -u github.com/erroneousboat/slack-term
```

Setup
-----

1. Get a slack token, click [here](https://api.slack.com/docs/oauth-test-tokens) 

2. Create a `slack-term.json` file, place it in your home directory. Below is
   an an example file, you can leave out the `OPTIONAL` parts, you are only
   required to specify a `slack_token`. Remember that your file should be
   a valid json file so don't forget to remove the comments.

```javascript
{
    "slack_token": "yourslacktokenhere",

    // OPTIONAL: add the following to use light theme, default is dark
    "theme": "light",

    // OPTIONAL: set the width of the sidebar (between 1 and 11), default is 1
    "sidebar_width": 3,

    // OPTIONAL: define custom key mappings, defaults are:
    "key_map": {
        "command": {
            "i":          "mode-insert",
            "k":          "channel-up",
            "j":          "channel-down",
            "g":          "channel-top",
            "G":          "channel-bottom",
            "<previous>": "chat-up",
            "C-b":        "chat-up",
            "C-u":        "chat-up",
            "<next>":     "chat-down",
            "C-f":        "chat-down",
            "C-d":        "chat-down",
            "q":          "quit",
            "<f1>":       "help"
        },
        "insert": {
            "<left>":      "cursor-left",
            "<right>":     "cursor-right",
            "<enter>":     "send",
            "<escape>":    "mode-command",
            "<backspace>": "backspace",
            "C-8":         "backspace",
            "<delete>":    "delete",
            "<space>":     "space"
        },
        "search": {
            "<left>":      "cursor-left",
            "<right>":     "cursor-right",
            "<escape>":    "clear-input",
            "<enter>":     "clear-input",
            "<backspace>": "backspace",
            "C-8":         "backspace",
            "<delete>":    "delete",
            "<space>":     "space"
        }
    }
}
```

Usage
-----

When everything is setup correctly you can run `slack-term` with the following
command: 

```bash
$ slack-term
```

You can also specify the location of the config file, this will give you
the possibility to run several instances of `slack-term` with different
accounts.

```bash
$ slack-term -config [path-to-config-file]
```

Default Key Mapping
-------------------

Below are the default key-mapping for `slack-term`, you can change them
in your `slack-term.json` file.

| mode    | key       | action                     |
|---------|-----------|----------------------------|
| command | `i`       | insert mode                |
| command | `/`       | search mode                |
| command | `k`       | move channel cursor up     |
| command | `j`       | move channel cursor down   |
| command | `g`       | move channel cursor top    |
| command | `G`       | move channel cursor bottom |
| command | `pg-up`   | scroll chat pane up        |
| command | `ctrl-b`  | scroll chat pane up        |
| command | `ctrl-u`  | scroll chat pane up        |
| command | `pg-down` | scroll chat pane down      |
| command | `ctrl-f`  | scroll chat pane down      |
| command | `ctrl-d`  | scroll chat pane down      |
| command | `q`       | quit                       |
| command | `f1`      | help                       |
| insert  | `left`    | move input cursor left     |
| insert  | `right`   | move input cursor right    |
| insert  | `enter`   | send message               |
| insert  | `esc`     | command mode               |
| search  | `esc`     | command mode               |
| search  | `enter`   | command mode               |
