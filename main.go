// TODO: use integer arithmetic
package main

import (
	"flag"
	"log"
	"math"
	"os"

	"github.com/go-audio/wav"
)

func divWhileBigger(target, divider, limit float64) float64 {
	for target > limit {
		target /= divider
	}

	return target
}

const samplesPerSecond = 48000

//const samplesPerSecond = sample(96000)
//const samplesPerSecond = sample((192000)

const baseFrequency = 432.0

const sineT = 2 * math.Pi

const numOfVoices = 512

type voices [numOfVoices]float64

// Optimization - Pre-generate frequencies for each voice to reduce computations
func newVoices() voices {
	const octave = 2
	const harmonicStep = 1.5
	var voices voices
	// Optimization - Hold the multiplication factor to prevent excessive divisions
	multiplier := harmonicStep
	for voice := 0; voice < numOfVoices; voice++ {
		voices[voice] = baseFrequency * multiplier
		multiplier = divWhileBigger(multiplier*harmonicStep, octave, octave)
	}
	return voices
}

func genSample(voices *voices, table *sinTable, sampleIdx int) uint32 {
	amp := 0.0
	for voice := 0; voice < numOfVoices; voice++ {
		amp += table.sin(sampleIdx, voices[voice])
	}

	return uint32(
		// This division corrects the volume(db)
		(amp / numOfVoices) * 2147483647,
	)
}

// TODO: Abstract the table to emulate any function including the result of this program :)
const tableSize = samplesPerSecond / 18 // Should be enough
type sinTable [tableSize]float64

// Optimization - use table to prevent unnecessary complex sin calculations
func newSinTable() *sinTable {
	var table sinTable
	step := sineT / float64(tableSize)
	for x := 0; x < tableSize; x++ {
		table[x] = math.Sin(float64(x) * step)
	}
	return &table
}

// TODO: Now it's worse than math.Sin ( Optimize this
func (t *sinTable) sin(x int, fq float64) float64 {
	_, phase := math.Modf((float64(x) / float64(samplesPerSecond)) * fq)
	return t[int(phase*float64(tableSize))%tableSize]
}

func main() {
	output := flag.String("output", "output.wav", "filename to write to")
	length := flag.Float64("length", 5, "length in seconds of output file")
	flag.Parse()

	log.Printf("generating a %f sec sample", *length)
	f, err := os.Create(*output)
	if err != nil {
		log.Fatalf("error creating %s: %s", *output, err)
	}
	defer f.Close()

	wavOut := wav.NewEncoder(f, samplesPerSecond, 32, 1, 1)
	totalSamples := int(math.Round(samplesPerSecond * *length))
	defer wavOut.Close()

	voices := newVoices()
	progressStep := totalSamples / 100
	table := newSinTable()
	for x := 0; x < totalSamples; x++ {
		wavOut.WriteFrame(genSample(&voices, table, x))
		if x%(progressStep) == 0 {
			log.Printf("Done %d of 100", x/progressStep)
		}
	}

	log.Printf("Done!")
}
