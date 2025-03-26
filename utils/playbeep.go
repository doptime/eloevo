package utils

import (
	"math"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

// sineWave is a custom stream that generates a sine wave
type sineWave struct {
	freq       float64
	sampleRate beep.SampleRate
	pos        float64
}

func (s *sineWave) Stream(samples [][2]float64) (n int, ok bool) {
	// Generate sine wave samples
	for i := range samples {
		angle := 2 * math.Pi * s.freq * s.pos / float64(s.sampleRate)
		value := math.Sin(angle)
		samples[i][0] = value // Left channel
		samples[i][1] = value // Right channel
		s.pos++
	}
	return len(samples), true
}

func (s *sineWave) Err() error {
	return nil
}

// PlayBeep plays 6 intermittent 0.2-second 440Hz beeps with 0.2-second intervals
func PlayBeep() {
	// Initialize the speaker with a sample rate of 44100 Hz
	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	// Create a sine wave generator for 440Hz frequency
	beepSound := &sineWave{
		freq:       440, // Frequency: 440Hz (A4)
		sampleRate: sampleRate,
		pos:        0,
	}

	// Play 6 beeps
	for i := 0; i < 6; i++ {
		// Play a 0.2-second beep
		speaker.Play(beep.Take(sampleRate.N(200*time.Millisecond), beepSound))
		time.Sleep(200 * time.Millisecond) // Wait for the beep to finish

		// Wait for an additional 0.2-second interval
		if i < 5 { // No delay after the last beep
			time.Sleep(200 * time.Millisecond)
		}
	}
}
