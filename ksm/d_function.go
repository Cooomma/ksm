package ksm

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/cooomma/fairplay-ksm/cryptos"
)

const PRIME = uint32(813416437)
const NB_RD = 16
const APPLE_TESTING_ASk = "d87ce7a26081de2e8eb8acef3a6dc179"

type DFunction struct {
}

func (d DFunction) Compute(R2 []byte, ask []byte) ([]byte, error) {

	if hex.EncodeToString(ask) == APPLE_TESTING_ASk {
		return hex.DecodeString(APPLE_TESTING_ASk)
	}

	hashValue, err := d.ComputeHashValue(R2)
	if err != nil {
		return nil, err
	}

	DASk, err := cryptos.AESECBEncrypt(ask, hashValue)
	if err != nil {
		return nil, err
	}

	if len(DASk) != 16 {
		return nil, errors.New("DASk key length doesn't equal 16")
	}

	return DASk, nil
}

func (d DFunction) ComputeHashValue(R2 []byte) ([]byte, error) {
	var pad []byte
	pad = make([]byte, 64, 64)

	var MBlock [14]uint32
	var r uint32
	var P uint32
	P = PRIME

	var i uint32

	R2_sz := uint32(len(R2))

	if len(R2) == 0 {
		return nil, errors.New("R2 block doesn't exist.")
	}

	/* Padding until a multiple of 56B */
	for i = 0; i < R2_sz; i++ {
		pad[i] = R2[i]
	}

	pad[R2_sz] = 0x80
	for i = R2_sz + 1; i < 56; i++ {
		pad[i] = 0
	}

	/* Create 14 32b values */
	extPad := copyPad(pad)
	for i = 0; i < 14; i++ {
		MBlock[i] = (extPad[4*i] << 24) ^ (extPad[4*i+1] << 16) ^ (extPad[4*i+2] << 8) ^ (extPad[4*i+3])
	}

	/* Reunify into 2 32 bits values */
	for i = 1; i < 7; i++ {
		MBlock[0] += MBlock[i]
	}

	MBlock[1] = 0
	for i = 0; i < 7; i++ {
		MBlock[1] += MBlock[i+7]
	}

	// Apply the function
	// This block is the C_r function specified in the D Function specification document
	for i = 0; i < 2; i++ {
		for r = 0; r < NB_RD; r++ {
			if MBlock[i]&1 == 1 {
				MBlock[i] >>= 1
			} else {
				MBlock[i] = (3*MBlock[i] + 1) % P
			}
		}
	}

	/* append to M */
	for i = 0; i < 4; i++ {
		pad[56+i] = uint8(MBlock[0] >> (8 * i))
		pad[60+i] = uint8(MBlock[1] >> (8 * i))
	}

	//pad have to 64 size
	h := sha1.New()
	h.Write(pad)

	hh := h.Sum(nil)

	if len(hh) != 20 {
		return nil, fmt.Errorf("hash value length must be 20, but %d", len(hh))
	}

	return hh[0:16], nil

}

func copyPad(pad []byte) []uint32 {
	extPad := make([]uint32, len(pad))

	for i, p := range pad {
		extPad[i] = uint32(p)
	}

	return extPad
}
