package pbrt

import (
	"fmt"
	"encoding/binary"
	"os"
	"io"
	"io/ioutil"
)

// 
// incident dir (negative cos)
//       _|      _
//         \     /|  outgoing dir (positive cos)
//          \   /
//		     \ /
// -----------*------------
//
// 1. mu_i, mu_o - cosines of zenith angles of incident and outgoing directions.
// 2. oi and oo are ints such that
// mu[oi] <= mu_i < mu[oi + 1]
// mu[oo] <= mu_o < mu[oo + 1]
// 3. pos is oo*NMu + oi
// 4. offset is AOffset[pos]
// 5. m is M[pos]
type FourierBSDFTable struct {
	Eta float32 // relative index of refraction on boundary
	MMax float32
	NChannels int32 // 3 for rgb, 1 for mono
	NMu int32
	Mu []float32 // zenith angles, sorted for binary search
	M []int32 // NMu x NMu. M[oo*NMu + oi] - how many fourier orders are needed for a pair
	          // of directions
	AOffset []int32
	A []float32 // fourier coefficients. For a pair of directions, they start at A[offset].
	            // for NChannels = 1, there are m of them.
				// for NChannels = 3, first m is luminance, then m red, then m blue
	A0 []float32
	Cdf []float32 // we invert cdf dimensions on import, because search will be
	              // this way
	Recip []float32
}

func ReadFourierBSDF(path string) (*FourierBSDFTable, error) {
	var table FourierBSDFTable
	file, err := os.Open(path)
	header := make([]uint8, 8)
	_, err = io.ReadFull(file, header)
	errwrap := func(format string, args ...interface{}) (*FourierBSDFTable, error) {
		s := fmt.Sprintf(format, args...)
		return nil, fmt.Errorf("read pbrt bsdf table from %q: %s", path, s)
	}
	if err != nil {
		return nil, err
	}
	if string(header) != "SCATFUN\x01" {
		return errwrap("header is not SCATFUN")
	}
	var flags, nCoeffs, nBases int32
	var unused3 [3]int32
	var unused4 [4]int32
	var offsetAndLength []int32
	if err = binary.Read(file, binary.LittleEndian, &flags); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &table.NMu); err != nil {
		return errwrap("corrupted file: %v")
	}
	if err = binary.Read(file, binary.LittleEndian, &nCoeffs); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &table.MMax); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &table.NChannels); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &nBases); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &unused3); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &table.Eta); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &unused4); err != nil {
		return errwrap("corrupted file")
	}

	if (flags != 1 || (table.NChannels != 1 && table.NChannels != 3) || nBases != 1) {
		return errwrap("unsupported file")
	}

	table.Mu = make([]float32, table.NMu)
	table.Cdf = make([]float32, table.NMu * table.NMu)
	table.A0 = make([]float32, table.NMu * table.NMu)
	offsetAndLength = make([]int32, table.NMu * table.NMu * 2)
	table.AOffset = make([]int32, table.NMu * table.NMu)
	table.M = make([]int32, table.NMu * table.NMu)
	table.A = make([]float32, nCoeffs)

	if err = binary.Read(file, binary.LittleEndian, &table.Mu); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &table.Cdf); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &offsetAndLength); err != nil {
		return errwrap("corrupted file")
	}
	if err = binary.Read(file, binary.LittleEndian, &table.A); err != nil {
		return errwrap("corrupted file")
	}

	for i := int32(0); i < table.NMu * table.NMu; i++ {
		offset := offsetAndLength[2 * i]
		length := offsetAndLength[2 * i + 1]
		table.AOffset[i] = offset
		table.M[i] = length
		if length > 0 {
			table.A0[i] = table.A[offset]
		}
	}

/*
	for i := int32(0); i < table.NMu; i++ {
		for j := int32(0); j < i; j++ {
			a := i*table.NMu + j
			b := j*table.NMu + i
			table.Cdf[a], table.Cdf[b] = table.Cdf[b], table.Cdf[a]
		}
	}*/

	rest, err := ioutil.ReadAll(file)
	if err != nil {
		return errwrap("corrupted file2: %v", err)
	}
	_ = rest
	//fmt.Println("rest", len(rest), "NMU", table.NMu, "len", len(table.M), table.Mu[len(table.Mu)-1])
	file.Close()

	return &table, nil
}

func Noop(){}

func main() {
	fmt.Println("vim-go")
}
