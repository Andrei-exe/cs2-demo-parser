# CS2 Demo Parser

> **Work in Progress:** This project is currently under active development. 

A small Counter-Strike 2 demo parser written in Go and based on the [`demoinfocs-golang`](https://github.com/markus-wa/demoinfocs-golang) package by markus-wa. It extracts match statistics, player performance, and chat logs directly from `.dem` files, outputting to terminal.

## Current Features

- **Scoreboard:** Extracts Kills, Deaths, and Assists. Ensures that kills occuring during knife rounds (for FACEIT) or warmups are not counted towards the final K/D/A.
- **Chat Logging:** Captures all `[ALL]` and `[TEAM]` chat messages sent during the match.
- **JSON Output:** Structured output to terminal (for now).

## Upcoming
 In the immediate future I want to implement functionalities to process more information from a match (e.g. total damage, hs%, utiliy damage etc.), display these statistics in a web interface and allow for uploading demo files from that web interface. Furthermore I'd like to implement a database connection to store statistics about players and also move the demo file storage there rather than locally.

## Instructions to Run
1. Create a `demos` folder in the root of the directory and drop the `.dem` file inside
2. From the root of the directory:
    ```bash
    go run ./cmd/cli/main.go ./demos/DEMO_FILE_NAME.dem