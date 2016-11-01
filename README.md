Slack-Term
==========

A [Slack](https://slack.com) client for your terminal.

![Screenshot](/screenshot.png?raw=true)

Getting started
---------------

1. [Download](https://github.com/erroneousboat/slack-term/releases) a
   compatible version for your system, and place where you can access it from
   the command line like, `~/bin`, `/usr/local/bin`, or `/usr/local/sbin`. Or
   get it via Go:


    ```bash
    $ go get github.com/erroneousboat/slack-term
    ```

2. Get a slack token, click [here](https://api.slack.com/docs/oauth-test-tokens) 

3. Create a `slack-term.json` file, place it in your home directory. The file
   should resemble the following structure:

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
                "<space>":     "space",
            }
        }
    }
    ```

4. Run `slack-term`: 

    ```bash
    $ slack-term

    // or specify the location of the config file
    $ slack-term -config [path-to-config-file]
    ```

Default Key Mapping
-------------------

| mode    | key       | action                     |
|---------|-----------|----------------------------|
| command | `i`       | insert mode                |
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
