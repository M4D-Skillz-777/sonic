package fingerprint

import (
	"encoding/binary"
	"io"
	"math"
	"math/cmplx"

	"audio-fingerprinting/fft"
	"github.com/argusdusty/gofft"
	mp3 "github.com/hajimehoshi/go-mp3"
)

const (
	SampleRate     = 44100
	WindowSize     = 2048
	HopSize        = 512
	TopFrequencies = 8
	MaxDurationSec = 10
	MinFreq        = 300
	MaxFreq        = 12000
)

type Fingerprint struct {
	Hashes map[uint64]struct{}
}

func DecodeAndFingerprint(data io.Reader) (*Fingerprint, error) {
	return DecodeAndFingerprintGofft(data)
}

func DecodeAndFingerprintGofft(data io.Reader) (*Fingerprint, error) {
	samples, err := decodeAudio(data)
	if err != nil {
		return nil, err
	}
	return generateFingerprintGofft(samples), nil
}

func DecodeAndFingerprintCustomFFT(data io.Reader) (*Fingerprint, error) {
	samples, err := decodeAudio(data)
	if err != nil {
		return nil, err
	}
	return generateFingerprintCustomFFT(samples), nil
}

func decodeAudio(data io.Reader) ([]float64, error) {
	buf := make([]byte, 12)
	n, err := data.Read(buf)
	if err != nil {
		return nil, err
	}

	var reader io.Reader = io.MultiReader(newReader(buf[:n]), data)

	if string(buf[:4]) == "RIFF" {
		return decodeWAV(reader)
	}

	return decodeMP3(reader)
}

func newReader(b []byte) io.Reader {
	return &byteReader{data: b}
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func decodeWAV(data io.Reader) ([]float64, error) {
	header := make([]byte, 44)
	_, err := io.ReadFull(data, header)
	if err != nil {
		return nil, err
	}

	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return nil, ErrInvalidWAV
	}

	audioFormat := binary.LittleEndian.Uint16(header[20:22])
	numChannels := binary.LittleEndian.Uint16(header[22:24])
	sampleRate := binary.LittleEndian.Uint32(header[24:28])
	bitsPerSample := binary.LittleEndian.Uint16(header[34:36])

	_ = sampleRate

	samples := make([]float64, 0, SampleRate*MaxDurationSec*2)

	for len(samples) < SampleRate*MaxDurationSec*2 {
		if audioFormat == 1 {
			if bitsPerSample == 16 {
				var sample int16
				if err := binary.Read(data, binary.LittleEndian, &sample); err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
				samples = append(samples, float64(sample)/32768.0)
				if numChannels == 2 {
					var sample2 int16
					if err := binary.Read(data, binary.LittleEndian, &sample2); err != nil {
						break
					}
				}
			} else if bitsPerSample == 8 {
				var sample uint8
				if err := binary.Read(data, binary.LittleEndian, &sample); err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
				samples = append(samples, (float64(sample)-128)/128.0)
				if numChannels == 2 {
					var sample2 uint8
					if err := binary.Read(data, binary.LittleEndian, &sample2); err != nil {
						break
					}
				}
			}
		} else {
			break
		}
	}

	if len(samples) < WindowSize {
		return nil, ErrAudioTooShort
	}

	return samples, nil
}

func decodeMP3(data io.Reader) ([]float64, error) {
	decoder, err := mp3.NewDecoder(data)
	if err != nil {
		return nil, err
	}

	samples := make([]float64, 0, SampleRate*MaxDurationSec*2)
	buf := make([]byte, 4096)
	for len(samples) < SampleRate*MaxDurationSec*2 {
		n, err := decoder.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for i := 0; i < n; i += 2 {
			if i+1 >= n {
				break
			}
			sample := int16(buf[i]) | int16(buf[i+1])<<8
			samples = append(samples, float64(sample)/32768.0)
		}
	}

	if len(samples) < WindowSize {
		return nil, ErrAudioTooShort
	}

	return samples, nil
}

func generateFingerprintGofft(samples []float64) *Fingerprint {
	numFrames := (len(samples) - WindowSize) / HopSize
	if numFrames < 1 {
		numFrames = 1
	}

	hashes := make(map[uint64]struct{}, numFrames)
	var prevFreqs []uint32

	for frame := 0; frame < numFrames; frame++ {
		start := frame * HopSize
		end := start + WindowSize
		if end > len(samples) {
			break
		}

		window := make([]float64, WindowSize)
		for i := 0; i < WindowSize; i++ {
			window[i] = samples[start+i] * hannWindow(i, WindowSize)
		}

		fftResult := performFFTGofft(window)
		peaks := extractPeaks(fftResult, SampleRate)

		if len(peaks) > 0 {
			hash := hashPeaks(peaks, prevFreqs, frame)
			if hash != 0 {
				hashes[hash] = struct{}{}
			}
			prevFreqs = peaks
		}
	}

	return &Fingerprint{Hashes: hashes}
}

func generateFingerprintCustomFFT(samples []float64) *Fingerprint {
	numFrames := (len(samples) - WindowSize) / HopSize
	if numFrames < 1 {
		numFrames = 1
	}

	hashes := make(map[uint64]struct{}, numFrames)

	var prevFreqs []uint32

	for frame := 0; frame < numFrames; frame++ {
		start := frame * HopSize
		end := start + WindowSize
		if end > len(samples) {
			break
		}

		window := samples[start:end]
		for i := 0; i < len(window); i++ {
			window[i] *= hannWindow(i, WindowSize)
		}

		fftResult := performFFT(window)
		peaks := extractPeaks(fftResult, SampleRate)

		if len(peaks) > 0 {
			hash := hashPeaks(peaks, prevFreqs, frame)
			if hash != 0 {
				hashes[hash] = struct{}{}
			}
			prevFreqs = peaks
		}
	}

	return &Fingerprint{Hashes: hashes}
}

func hannWindow(i, n int) float64 {
	return 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
}

func performFFT(samples []float64) []float64 {
	fftResult := fft.FFTReal(samples)

	result := make([]float64, len(fftResult))
	for i := 0; i < len(result); i++ {
		result[i] = cmplx.Abs(fftResult[i])
	}

	return result
}

func performFFTGofft(samples []float64) []float64 {
	n := len(samples)
	nextPow2 := 1
	for nextPow2 < n {
		nextPow2 *= 2
	}

	padded := make([]complex128, nextPow2)
	for i := 0; i < n; i++ {
		padded[i] = complex(samples[i], 0)
	}

	gofft.FFT(padded)

	result := make([]float64, nextPow2/2)
	for i := 0; i < len(result); i++ {
		result[i] = cmplx.Abs(padded[i])
	}

	return result
}

func extractPeaks(fftResult []float64, sampleRate int) []uint32 {
	type peak struct {
		freq      uint32
		magnitude float64
	}

	var peaks []peak
	binSize := sampleRate / len(fftResult)

	minBin := MinFreq / binSize
	maxBin := MaxFreq / binSize

	for bin := minBin; bin < maxBin && bin < len(fftResult); bin++ {
		mag := fftResult[bin]
		if mag > 0.001 {
			isPeak := true
			for j := bin - 3; j <= bin+3; j++ {
				if j >= 0 && j < len(fftResult) && j != bin {
					if fftResult[j] >= mag*0.95 {
						isPeak = false
						break
					}
				}
			}
			if isPeak {
				peaks = append(peaks, peak{freq: uint32(bin * binSize), magnitude: mag})
			}
		}
	}

	for i := 0; i < len(peaks); i++ {
		for j := i + 1; j < len(peaks); j++ {
			if peaks[j].magnitude > peaks[i].magnitude {
				peaks[i], peaks[j] = peaks[j], peaks[i]
			}
		}
	}

	if len(peaks) > TopFrequencies {
		peaks = peaks[:TopFrequencies]
	}

	result := make([]uint32, len(peaks))
	for i, p := range peaks {
		result[i] = p.freq
	}

	return result
}

func hashPeaks(peaks []uint32, prevPeaks []uint32, frame int) uint64 {
	if len(peaks) < 2 {
		return 0
	}

	var hash uint64

	hash = uint64(peaks[0]) << 48
	hash |= uint64(peaks[1]) << 32

	if len(prevPeaks) >= 2 {
		delta := frame
		hash |= uint64(delta&0xFFFF) << 16
		hash |= uint64((prevPeaks[0] ^ peaks[0]) & 0xFFFF)
	} else if len(peaks) >= 3 {
		hash |= uint64(peaks[2]) << 16
	}

	return hash
}

var (
	ErrAudioTooShort = &customError{message: "audio too short"}
	ErrInvalidWAV    = &customError{message: "invalid WAV format"}
)

type customError struct {
	message string
}

func (e *customError) Error() string {
	return e.message
}
