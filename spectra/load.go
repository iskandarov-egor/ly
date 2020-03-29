package spectra

import (
	"os"
	"fmt"
	"bufio"
)

func Load(path string) (Spectr, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	table := NewSpectrTable()
	scanner := bufio.NewScanner(f)
	lineno := 1
	var prevWave float32
	for scanner.Scan() {
		var wave, power float32
		_, err := fmt.Sscanf(scanner.Text(), "%f %f", &wave, &power)
		if err != nil || prevWave > wave {
			return nil, fmt.Errorf("failed to parse spd: line %d, file %q %v", lineno, path, err)
		}
		table.AppendSample(wave, power)
		prevWave = wave

		lineno++
	}
	return table.MakeRGBSpectr(), nil
}
