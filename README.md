## What it does
Concurrently crawls pages of the given domain, making a local copy of the site. Only `.html` files are saved, links to other resources are converted to absolute.

## How to run
- Clone and open repository
- Run command
    ```bash
    go run . [-depth depth] [-par parallelism_degree] [-path path] [URL]...
    ```
  - Example: 
    ```bash
    go run . -depth 5 -par 3 -path . https://www.gr-oborona.ru/
    ```
- To get more information about the options use
    ```bash 
    go run . --help
    ```