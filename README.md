# CS2 Demo Parser

> **Work in Progress:** This project is currently under active development. 

A small Counter-Strike 2 demo parser written in Go and based on the [`demoinfocs-golang`](https://github.com/markus-wa/demoinfocs-golang) package by markus-wa. It extracts match statistics, player performance, and chat logs directly from `.dem` files, outputting to terminal & web interface.

## Current Features

- **Scoreboard:** Extracts Kills, Deaths, Assists, Damage, HS%, Utility Damage and Enemies Flashed. Ensures that events occuring during knife rounds or warmups are not counted towards the final score.
- **Chat Logging:** Captures all `[ALL]` and `[TEAM]` chat messages sent during the match.
- **JSON Output:** Structured output to terminal.
- **API:** Service written using the [`fiber`](https://github.com/gofiber/fiber) package to serve the frontend.
- **Web interface:** Basic web interface written in Vue that allows for uploading `.dem` files and displays a basic scoreboard with all the statistics received from the GO backend.

## Upcoming
 In the immediate future I want to implement functionalities to process more information from a match and display these statistics in a web interface. Furthermore, I'd like to implement a database connection to store statistics about players and also move the demo file storage there rather than locally.

## Instructions to Run (the terminal output functionality)

1. Create a `demos` folder in the root of the directory and drop the `.dem` file inside
2. From the root of the directory:
    ```bash
    go run ./cmd/cli/main.go ./demos/DEMO_FILE_NAME.dem