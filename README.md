# timterests

Personal blog to cover topics that I am interested in.
Tools and platforms utized for this project:

- Go
- Templ
- HTMX
- TailwindCSS
- SQLlite
- Docker
- AWS
- GitHub Actions
-

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## Tailwind

For any style updates in tailwindcss, a tailwindcss standalone cli binary needs to be installed from below.
<https://tailwindcss.com/blog/standalone-cli>

Available versions:
<https://github.com/tailwindlabs/tailwindcss/releases/tag/v3.4.17>

To run an update to the css, perform one of the following.
For a minified version:

```bash
./tailwindcss -i ./cmd/web/assets/css/styles.css -o ./cmd/web/assets/css/output.css --minify
```

For a watch version for dev environments:

```bash
./tailwindcss -i ./cmd/web/assets/css/styles.css -o ./cmd/web/assets/css/output.css --watch
```

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB Container

```bash
make docker-down
```

DB Integrations Test:

```bash
make itest
```

Live reload the application:

```bash
make watch
```

Run the test suite:

```bash
make test
```

Clean up binary from the last build:

```bash
make clean
```
