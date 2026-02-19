# AGENTS.md

Guidelines for AI coding agents working on the Sonic audio fingerprinting project.

## Project Overview

Sonic is an audio fingerprinting service with:
- **Backend**: Go REST API using FFT for fingerprinting, Redis for storage
- **Frontend**: SvelteKit web app with dark-themed UI

## Build Commands

### Go Backend

```bash
# Install dependencies
go mod tidy

# Build binary
go build -o audio-fingerprinting .

# Run server (requires Redis at localhost:6379)
./audio-fingerprinting

# Run directly without building
go run .
```

### SvelteKit Frontend

```bash
# Install dependencies
cd sonic && npm install

# Development server (http://localhost:5173)
cd sonic && npm run dev

# Production build
cd sonic && npm run build

# Preview production build
cd sonic && npm run preview
```

### Full Stack

```bash
# Terminal 1: Redis
redis-server

# Terminal 2: Backend
./audio-fingerprinting

# Terminal 3: Frontend
cd sonic && npm run dev
```

## Testing

Currently no test suite is configured. When adding tests:

```bash
# Go tests (when added)
go test ./...

# Go tests with coverage
go test -cover ./...

# Run specific package tests
go test ./fingerprint -v
```

## Project Structure

```
audio-fingerprinting/
├── main.go                 # Entry point, server setup
├── config/
│   └── config.go           # Configuration (env vars)
├── handler/
│   └── handler.go          # HTTP handlers for all endpoints
├── fingerprint/
│   └── fingerprint.go      # MP3 decoding, FFT, fingerprint generation
├── fft/
│   └── fft.go              # Custom Cooley-Tukey FFT implementation
├── matcher/
│   └── matcher.go          # Redis storage and fuzzy matching
├── sonic/                  # SvelteKit frontend
│   ├── src/
│   │   ├── routes/
│   │   │   ├── +page.svelte    # Main app page
│   │   │   └── +layout.svelte  # Root layout
│   │   ├── lib/
│   │   │   ├── api.ts          # API client functions
│   │   │   └── components/     # Reusable Svelte components
│   │   ├── app.css             # Global styles (CSS variables)
│   │   └── app.html            # HTML template
│   ├── static/                 # Static assets
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.ts
├── go.mod
└── go.sum
```

## Code Style Guidelines

### Go

**Imports** - Group in this order:
1. Standard library
2. External packages
3. Local packages

```go
import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "audio-fingerprinting/config"
)
```

**Naming Conventions**:
- Variables/functions: `camelCase`
- Exported functions/types: `PascalCase`
- Constants: `UPPER_SNAKE_CASE` or `PascalCase`
- Interfaces: `PascalCase` with `er` suffix (e.g., `Matcher`)

**Error Handling**:
- Return errors from functions, don't log and exit
- Handle errors in handlers with appropriate HTTP status codes
- Use `gin.H{"error": err.Error()}` for error responses

```go
// Good
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}

// Avoid
if err != nil {
    log.Fatal(err)
}
```

**HTTP Handlers**:
- Use `c.JSON(status, gin.H{...})` for responses
- Check required params early and return
- Use `c.PostForm()` for form data, `c.Param()` for URL params

### TypeScript/Svelte

**TypeScript**:
- Use strict mode
- Define interfaces for all API responses
- Use `async/await` over `.then()` chains
- Export interfaces from `api.ts`

```typescript
// Good
export interface MatchResult {
  song_name: string;
  confidence: number;
}

// API calls
async function fetchData(): Promise<MatchResult> {
  const response = await fetch(url);
  return handleResponse<MatchResult>(response);
}
```

**Svelte Components**:
- Use Svelte 5 runes: `$state`, `$derived`, `$effect`
- Import from `$lib/` for local modules
- Use `createEventDispatcher()` for component events
- Use `export let` for component props

```svelte
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export let value: string;
  
  let count = $state(0);
  let doubled = $derived(count * 2);
  
  const dispatch = createEventDispatcher();
</script>
```

**Styling**:
- Use CSS variables defined in `app.css`
- Follow naming: `--accent-cyan`, `--text-primary`, etc.
- Scoped styles in `<style>` block within components
- Use BEM-like naming for CSS classes

```css
.card { }
.card-title { }
.card--active { }
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/fingerprint` | Register song (gofft) |
| POST | `/fingerprint/custom` | Register song (Cooley-Tukey) |
| POST | `/recognize` | Identify song |
| POST | `/recognize/custom` | Identify song (custom FFT) |
| DELETE | `/fingerprint/:name` | Delete song |
| GET | `/songs` | List all songs |

## Configuration

**Environment Variables**:
- `REDIS_ADDR`: Redis address (default: `localhost:6379`)
- `PORT`: Server port (default: `8080`)

**Frontend API Base**:
- Located in `sonic/src/lib/api.ts`
- `API_BASE = 'http://localhost:8080'`

## Key Patterns

### Adding a New API Endpoint

1. Add route in `handler/handler.go` → `SetupRoutes()`
2. Create handler method on `Handler` struct
3. Call matcher for data operations
4. Return JSON response

### Adding a New Frontend Feature

1. Create component in `sonic/src/lib/components/`
2. Import in `+page.svelte` or parent component
3. Add API function in `sonic/src/lib/api.ts` if needed
4. Use CSS variables for consistent styling

### Redis Data Model

```
songs                   → SET of song names
song:{song_name}:hashes → SET of fingerprint hashes
```

## Common Tasks

**Add a new FFT algorithm**:
1. Add function in `fft/fft.go`
2. Add new endpoint or parameter in `fingerprint/fingerprint.go`
3. Add corresponding handler in `handler/handler.go`

**Modify UI theme**:
1. Update CSS variables in `sonic/src/app.css`
2. Use variables in component styles

**Add external API integration**:
1. Add interface in `sonic/src/lib/api.ts`
2. Create fetch function with error handling
3. Integrate into relevant component
