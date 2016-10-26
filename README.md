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
    $ go get github.com/erroneouboat/slack-term
    ```

2. Get a slack token, click [here](https://api.slack.com/docs/oauth-test-tokens) 

3. Create a `slack-term.json` file, place it in your home directory. The file
   should resemble the following structure:

    ```javascript
    {
        "slack_token": "yourslacktokenhere",

        // optional: add the following to use light theme, default is dark
        "theme": "light",

        // optional: set the width of the sidebar (between 1 and 11), default is 1
        "sidebar_width": 3
    }
    ```

4. Run `slack-term`: 

    ```bash
    $ slack-term

    // or specify the location of the config file
    $ slack-term -config [path-to-config-file]
    ```

Usage
-----

| mode   | key       | action                     |
|--------|-----------|----------------------------|
| normal | `i`       | insert mode                |
| normal | `k`       | move channel cursor up     |
| normal | `j`       | move channel cursor down   |
| normal | `gg`      | move channel cursor top    |
| normal | `G`       | move channel cursor bottom |
| normal | `pg-up`   | scroll chat pane up        |
| normal | `ctrl-b`  | scroll chat pane up        |
| normal | `ctrl-u`  | scroll chat pane up        |
| normal | `pg-down` | scroll chat pane down      |
| normal | `ctrl-f`  | scroll chat pane down      |
| normal | `ctrl-d`  | scroll chat pane down      |
| normal | `pg-down` | scroll chat pane down      |
| normal | `q`       | quit                       |
| insert | `left`    | move input cursor left     |
| insert | `right`   | move input cursor right    |
| insert | `enter`   | send message               |
| insert | `esc`     | normal mode                |
