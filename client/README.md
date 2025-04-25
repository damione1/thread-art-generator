# Thread Art Generator - Go+HTMX Frontend

This is the Go+HTMX frontend for the Thread Art Generator application.

## Architecture

The frontend is built using:

- Go for server-side logic
- [Templ](https://github.com/a-h/templ) for templating
- [HTMX](https://htmx.org/) for interactive UI
- [Tailwind CSS](https://tailwindcss.com/) for styling

## Development

### Prerequisites

- Go 1.21+
- Node.js and npm (for Tailwind CSS)
- Templ CLI (`go install github.com/a-h/templ/cmd/templ@latest`)

### Local Setup

1. Install the Templ CLI:

   ```
   go install github.com/a-h/templ/cmd/templ@latest
   ```

2. Install dependencies:

   ```
   cd client
   npm install
   ```

3. Build Tailwind CSS:

   ```
   npm run build
   ```

4. Generate Templ files:

   ```
   templ generate ./internal/templates
   ```

5. Run the server:
   ```
   go run cmd/frontend/main.go
   ```

Or use Tilt for development:

```
tilt up
```

## Project Structure

- `cmd/frontend`: Entry point for the web server
- `internal/handlers`: Request handlers
- `internal/middleware`: HTTP middleware
- `internal/templates`: Templ templates
- `public`: Static assets (CSS, JS, images)
- `styles`: Tailwind CSS input files
