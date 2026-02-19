package fft

import "math"

func CooleyTukeyFFT(input []complex128) []complex128 {
	n := len(input)
	if n <= 1 {
		result := make([]complex128, n)
		copy(result, input)
		return result
	}

	nextPow2 := 1
	for nextPow2 < n {
		nextPow2 *= 2
	}

	padded := make([]complex128, nextPow2)
	copy(padded, input)

	bitReverse(padded)

	for size := 2; size <= nextPow2; size *= 2 {
		halfSize := size / 2
		angle := -2 * math.Pi / float64(size)
		wLen := complex(math.Cos(angle), math.Sin(angle))

		for i := 0; i < nextPow2; i += size {
			w := complex(1, 0)
			for j := 0; j < halfSize; j++ {
				u := padded[i+j]
				t := w * padded[i+j+halfSize]
				padded[i+j] = u + t
				padded[i+j+halfSize] = u - t
				w *= wLen
			}
		}
	}

	return padded
}

func bitReverse(data []complex128) {
	n := len(data)
	bits := int(math.Log2(float64(n)))

	for i := 1; i < n; i++ {
		j := reverseBits(i, bits)
		if j > i {
			data[i], data[j] = data[j], data[i]
		}
	}
}

func reverseBits(n, bits int) int {
	result := 0
	for i := 0; i < bits; i++ {
		result = (result << 1) | (n & 1)
		n >>= 1
	}
	return result
}

func FFTReal(samples []float64) []complex128 {
	n := len(samples)
	input := make([]complex128, n)
	for i := 0; i < n; i++ {
		input[i] = complex(samples[i], 0)
	}
	return CooleyTukeyFFT(input)
}
