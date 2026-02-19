# Sonic - AI-Powered Audio Fingerprinting

A modern audio fingerprinting service with a beautiful dark-themed web interface. Built with Go (backend) and SvelteKit (frontend).

## ✨ Features

### Backend (Go)
- **Dual FFT Implementations**: Optimized gofft + custom Cooley-Tukey algorithm
- **Fuzzy Matching**: 60% similarity threshold for robust recognition
- **Redis Storage**: Fast fingerprint lookup and storage
- **REST API**: Full CRUD operations for fingerprint management
- **CORS Enabled**: Ready for frontend integration

### Frontend (SvelteKit - "Sonic")
- **Modern Dark Theme**: Black gradient with glassmorphism UI
- **Inter Font**: Professional typography
- **Drag & Drop Upload**: Smooth file handling
- **Audio Visualizer**: Real-time FFT frequency bars with play/pause
- **Waveform Display**: Visual audio representation
- **Circular Progress**: Spring-animated progress indicators
- **Confetti Effects**: Celebration on successful match
- **No-Match Flow**: Auto-redirect to register unknown songs
- **Spotify Integration**: Album art and metadata lookup
- **AcoustID Ready**: External fingerprint database integration

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Redis (localhost:6379)

### Installation

```bash
# Clone the project
cd audio-fingerprinting

# Install Go dependencies
go mod tidy

# Build backend
go build -o audio-fingerprinting .

# Install frontend dependencies
cd sonic && npm install && cd ..
```

### Running

```bash
# Terminal 1: Start Redis
redis-server

# Terminal 2: Start Backend
./audio-fingerprinting

# Terminal 3: Start Frontend
cd sonic && npm run dev
```

- **Backend**: http://localhost:8080
- **Frontend**: http://localhost:5173

## 🎨 UI Features

| Feature | Description |
|---------|-------------|
| **Gradient Theme** | Black gradient with cyan/purple accents |
| **Glassmorphism** | Frosted glass cards with blur effects |
| **Spring Animations** | Smooth physics-based interactions |
| **Circular Progress** | Animated ring progress indicators |
| **Confetti** | Celebration particles on match success |
| **Responsive** | Mobile-first design |

## 🔌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/fingerprint` | Register song (gofft) |
| POST | `/fingerprint/custom` | Register song (Cooley-Tukey) |
| POST | `/recognize` | Identify song (gofft) |
| POST | `/recognize/custom` | Identify song (Cooley-Tukey) |
| DELETE | `/fingerprint/:name` | Delete song |
| GET | `/songs` | List all songs |

## 🌐 External APIs

### AcoustID Integration
- Free, open-source fingerprint database
- 34+ million tracks
- Automatic metadata lookup

### Spotify Integration
- Album art display
- Artist/album metadata
- Direct Spotify links

## 📁 Project Structure

```
audio-fingerprinting/
├── main.go                    # Entry point
├── config/config.go           # Configuration
├── fingerprint/
│   └── fingerprint.go         # FFT fingerprinting
├── fft/
│   └── fft.go                 # Custom Cooley-Tukey FFT
├── matcher/
│   └── matcher.go             # Redis matching
├── handler/
│   └── handler.go             # HTTP handlers
├── sonic/                     # SvelteKit frontend
│   ├── src/
│   │   ├── routes/+page.svelte
│   │   ├── lib/
│   │   │   ├── api.ts
│   │   │   └── components/
│   │   └── app.css
│   └── package.json
└── README.md
```

## 🔧 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `PORT` | `8080` | Backend port |

## 🎯 Technical Highlights

### Fingerprinting Algorithm
1. Decode MP3 to PCM samples (44100 Hz)
2. Apply Hann window (2048 samples, 50% overlap)
3. FFT transform to frequency domain
4. Extract top 5 peaks (300-10000 Hz)
5. Generate 64-bit hash from (freq, time_delta)
6. Store in Redis for O(1) lookup

### Matching Algorithm
1. Generate fingerprint from query
2. Count matching hashes per song
3. Calculate similarity ratio
4. Return best match ≥ 60%

## 📊 Performance

- **Fingerprinting**: ~1-2s for 10-second audio
- **Recognition**: ~100-500ms
- **Memory**: Minimal (64-bit hashes only)

## 📝 License

MIT

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok"
}
```

### Register a Song (gofft - Optimized FFT)

```bash
curl -X POST -F "file=@song.mp3" -F "name=My Song" http://localhost:8080/fingerprint
```

Response:
```json
{
  "song_name": "My Song",
  "hashes": 150,
  "fft_impl": "gofft",
  "fft_details": "Optimized in-place radix-2 FFT from gofft library"
}
```

### Register a Song (Custom Cooley-Tukey FFT)

```bash
curl -X POST -F "file=@song.mp3" -F "name=My Song" http://localhost:8080/fingerprint/custom
```

Response:
```json
{
  "song_name": "My Song",
  "hashes": 150,
  "fft_impl": "custom_cooley_tukey",
  "fft_details": "In-place radix-2 Cooley-Tukey FFT algorithm"
}
```

### Recognize a Song (gofft)

```bash
curl -X GET -F "file=@sample.mp3" http://localhost:8080/recognize
```

Response (match found):
```json
{
  "song_name": "My Song",
  "confidence": 0.75
}
```

Response (no match):
```json
{
  "song_name": "",
  "confidence": 0,
  "message": "no match found"
}
```

### Recognize a Song (Custom FFT)

```bash
curl -X GET -F "file=@sample.mp3" http://localhost:8080/recognize/custom
```

### List All Registered Songs

```bash
curl http://localhost:8080/songs
```

Response:
```json
{
  "songs": ["Song 1", "Song 2", "Song 3"]
}
```

### Delete a Song

```bash
curl -X DELETE http://localhost:8080/fingerprint/My%20Song
```

Response:
```json
{
  "song_name": "My Song",
  "deleted": true
}
```

## Inspecting Redis

### Connect to Redis CLI

```bash
redis-cli
```

### List All Registered Songs

```bash
redis-cli SMEMBERS songs
```

### View Fingerprints for a Song

```bash
redis-cli SMEMBERS "song:My Song:hashes"
```

### Count Fingerprints for a Song

```bash
redis-cli SCARD "song:My Song:hashes"
```

### List All Keys

```bash
redis-cli KEYS "*"
```

### Delete All Data (Reset)

```bash
redis-cli FLUSHDB
```

### Watch Real-time Operations

In one terminal:
```bash
redis-cli MONITOR
```

Then make API requests in another terminal to see all Redis operations in real-time.

## How It Works

### Fingerprinting Algorithm

1. **Decode MP3**: Convert MP3 to PCM float64 samples (44100 Hz sample rate)
2. **Windowing**: Apply Hann window to 2048-sample frames (50% overlap)
3. **FFT**: Transform time-domain samples to frequency domain
4. **Peak Extraction**: Find top 5 strongest frequencies (300-10000 Hz range)
5. **Hash Generation**: Create 64-bit hash from (freq1, freq2, time_delta)
6. **Storage**: Store hashes in Redis SET for O(1) lookup

### Matching Algorithm

1. Generate fingerprint hashes from query audio
2. For each stored song, count matching hashes
3. Calculate: `similarity = matching_hashes / total_query_hashes`
4. Return song with highest similarity >= 60%

## Redis Data Model

```
songs                   → SET of song names
song:{song_name}:hashes → SET of fingerprint hashes (uint64 as strings)
```

## Project Structure

```
audio-fingerprinting/
├── main.go              # Entry point, server setup
├── config/
│   └── config.go        # Configuration management
├── fingerprint/
│   └── fingerprint.go   # MP3 decoding, FFT, fingerprinting
├── fft/
│   └── fft.go           # Custom Cooley-Tukey FFT implementation
├── matcher/
│   └── matcher.go       # Redis storage and fuzzy matching
├── handler/
│   └── handler.go       # HTTP handlers for all endpoints
├── sonic/               # SvelteKit frontend
│   ├── src/
│   │   ├── routes/
│   │   │   └── +page.svelte    # Main app page
│   │   ├── lib/
│   │   │   ├── api.ts          # API client
│   │   │   └── components/     # UI components
│   │   ├── app.css             # Global styles
│   │   └── app.html            # HTML template
│   ├── static/                 # Static assets
│   └── package.json
├── go.mod
└── go.sum
```

## Sonic Frontend

The web interface "Sonic" provides a beautiful, dark-themed UI with:

- **Drag & Drop Upload**: Easy MP3 file selection
- **Waveform Display**: Visual representation of audio
- **Audio Visualizer**: Real-time FFT frequency bars
- **Mode Toggle**: Switch between Identify and Register modes
- **FFT Selection**: Choose between optimized or custom FFT
- **Progress Indicator**: Visual feedback during processing
- **Song Management**: View and delete registered songs

### Running Sonic

```bash
cd sonic
npm install
npm run dev
```

Open http://localhost:5173 in your browser.

## Example Workflow

```bash
# 1. Start the server
./audio-fingerprinting

# 2. Register a few songs
curl -X POST -F "file=@song1.mp3" -F "name=Song One" http://localhost:8080/fingerprint
curl -X POST -F "file=@song2.mp3" -F "name=Song Two" http://localhost:8080/fingerprint/custom
curl -X POST -F "file=@song3.mp3" -F "name=Song Three" http://localhost:8080/fingerprint

# 3. Check registered songs
curl http://localhost:8080/songs

# 4. Recognize a sample (5-10 sec snippet)
curl -X GET -F "file=@sample.mp3" http://localhost:8080/recognize

# 5. Check Redis data
redis-cli SMEMBERS songs
redis-cli SCARD "song:Song One:hashes"

# 6. Delete a song
curl -X DELETE http://localhost:8080/fingerprint/Song%20One
```

## Troubleshooting

### Redis Connection Error

```
Failed to connect to Redis: dial tcp [::1]:6379: connect: connection refused
```

Solution: Start Redis
```bash
# macOS (Homebrew)
brew services start redis

# Linux
sudo systemctl start redis

# Docker
docker run -d -p 6379:6379 redis
```

### Empty File Error

Make sure the MP3 file exists and is readable:
```bash
ls -la song.mp3
```

### No Match Found

- Ensure songs are registered first (check with `/songs` endpoint)
- Use 5-10 seconds of audio for better matching
- Audio quality affects matching accuracy
- Try both FFT implementations (`/recognize` vs `/recognize/custom`)

## Performance

- **Fingerprinting**: ~1-2 seconds for 10-second audio
- **Recognition**: ~100-500ms depending on number of stored songs
- **Memory**: Minimal (only stores 64-bit hashes)

## License

MIT
