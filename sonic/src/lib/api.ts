const API_BASE = 'http://localhost:8080';
const ACOUSTID_API = 'https://api.acoustid.org/v2/lookup';
const ACOUSTID_KEY = 'cSpUJKoF';
const SPOTIFY_CLIENT_ID = '';

export interface Song {
  song_name: string;
  confidence?: number;
  hashes?: number;
  fft_impl?: string;
  fft_details?: string;
  message?: string;
}

export interface SongsResponse {
  songs: string[];
}

export interface MatchResult {
  song_name: string;
  confidence: number;
  message?: string;
  fft_impl?: string;
  source?: 'local' | 'acoustid';
  spotify?: SpotifyTrack;
}

export interface FingerprintResponse {
  song_name: string;
  hashes: number;
  fft_impl: string;
  fft_details?: string;
}

export interface SpotifyTrack {
  id: string;
  name: string;
  artist: string;
  album: string;
  album_art: string;
  preview_url: string | null;
  spotify_url: string;
}

export interface AcoustIDResult {
  id: string;
  score: number;
  recordings: Array<{
    id: string;
    title: string;
    artists: Array<{ name: string }>;
  }>;
}

async function handleResponse<T>(response: Response): Promise<T> {
  const text = await response.text();
  
  if (!response.ok) {
    try {
      const error = JSON.parse(text);
      throw new Error(error.error || 'Request failed');
    } catch {
      throw new Error(text || 'Request failed');
    }
  }
  
  try {
    return JSON.parse(text);
  } catch {
    throw new Error('Invalid JSON response');
  }
}

export async function registerSong(file: File, name: string, useCustomFFT = false): Promise<FingerprintResponse> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('name', name);

  const endpoint = useCustomFFT ? '/fingerprint/custom' : '/fingerprint';
  
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: 'POST',
    body: formData
  });

  return handleResponse<FingerprintResponse>(response);
}

export async function recognizeSong(file: File, useCustomFFT = false): Promise<MatchResult> {
  const formData = new FormData();
  formData.append('file', file);

  const endpoint = useCustomFFT ? '/recognize/custom' : '/recognize';
  
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: 'POST',
    body: formData
  });

  const result = await handleResponse<MatchResult>(response);

  if (result.song_name && result.confidence > 0.5) {
    const spotify = await fetchSpotifyMetadata(result.song_name);
    if (spotify) {
      result.spotify = spotify;
    }
  }

  return result;
}

export async function lookupAcoustID(fingerprint: string, duration: number): Promise<AcoustIDResult[]> {
  try {
    const params = new URLSearchParams({
      client: ACOUSTID_KEY,
      fingerprint: fingerprint,
      duration: duration.toString(),
      meta: 'recordings+releasegroups+compress'
    });

    const response = await fetch(`${ACOUSTID_API}?${params}`);
    const data = await response.json();

    if (data.status === 'ok' && data.results) {
      return data.results;
    }
    return [];
  } catch {
    return [];
  }
}

export async function fetchSpotifyMetadata(songName: string): Promise<SpotifyTrack | null> {
  try {
    const query = encodeURIComponent(songName);
    const response = await fetch(`https://api.spotify.com/v1/search?q=${query}&type=track&limit=1`, {
      headers: SPOTIFY_CLIENT_ID ? {
        'Authorization': `Bearer ${SPOTIFY_CLIENT_ID}`
      } : {}
    });

    if (!response.ok) return null;

    const data = await response.json();
    
    if (data.tracks?.items?.length > 0) {
      const track = data.tracks.items[0];
      return {
        id: track.id,
        name: track.name,
        artist: track.artists?.map((a: any) => a.name).join(', ') || 'Unknown',
        album: track.album?.name || 'Unknown',
        album_art: track.album?.images?.[0]?.url || '',
        preview_url: track.preview_url,
        spotify_url: track.external_urls?.spotify || ''
      };
    }
    return null;
  } catch {
    return null;
  }
}

export async function listSongs(): Promise<string[]> {
  const response = await fetch(`${API_BASE}/songs`);
  const data = await handleResponse<SongsResponse>(response);
  return data.songs || [];
}

export async function deleteSong(name: string): Promise<void> {
  const response = await fetch(`${API_BASE}/fingerprint/${encodeURIComponent(name)}`, {
    method: 'DELETE'
  });

  await handleResponse<void>(response);
}

export async function checkHealth(): Promise<boolean> {
  try {
    const response = await fetch(`${API_BASE}/health`);
    return response.ok;
  } catch {
    return false;
  }
}
