package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	AUNDEF = iota
	AADC
	AADD
	AADD8
	AADDHi
	AADDI
	AAND
	AASR
	AASRI
	AB
	ABCC
	ABCS
	ABEQ
	ABGE
	ABGT
	ABHI
	ABIC
	ABKPT
	ABLE
	ABLS
	ABLT
	ABMI
	ABNE
	ABPL
	ABVC
	ABVS
	ABX
	ABXHi
	ACMN
	ACMP
	ACMP8
	ACMPHi
	AEOR
	ALDR
	ALDRB
	ALDRBI
	ALDRH
	ALDRI
	ALDRPC
	ALDSB
	ALDSH
	ALSL
	ALSLI
	ALSR
	ALSRI
	AMOV
	AMOV8
	AMOVHi
	AMUL
	AMVN
	ANEG
	AORR
	AROR
	ASBC
	ASTR
	ASTRB
	ASTRBI
	ASTRH
	ASTRI
	ASUB
	ASUB8
	ASUBI
	ASWI
	ATST

	AMax
)

var anames = [AMax]string{
	AADC: "adc",
	AADD: "add",
	AADD8: "add",
	AADDHi: "add",
	AADDI: "add",
	AAND: "and",
	AASR: "asr",
	AASRI: "asr",
	AB: "b",
	ABCC: "blo",
	ABCS: "bhs",
	ABEQ: "beq",
	ABGE: "bge",
	ABGT: "bgt",
	ABHI: "bhi",
	ABIC: "bic",
	ABKPT: "bkpt",
	ABLE: "ble",
	ABLS: "bls",
	ABLT: "blt",
	ABMI: "bmi",
	ABNE: "bne",
	ABPL: "bpl",
	ABVC: "bvc",
	ABVS: "bvs",
	ABX: "bx",
	ABXHi: "bx",
	ACMN: "cmn",
	ACMP: "cmp",
	ACMP8: "cmp",
	ACMPHi: "cmp",
	AEOR: "eor",
	ALDR: "ldr",
	ALDRB: "ldrb",
	ALDRBI: "ldrb",
	ALDRH: "ldrh",
	ALDRI: "ldr",
	ALDRPC: "ldr",
	ALDSB: "lds",
	ALDSH: "lds",
	ALSL: "lsl",
	ALSLI: "lsl",
	ALSR: "lsr",
	ALSRI: "lsr",
	AMOV: "mov",
	AMOV8: "mov",
	AMOVHi: "mov",
	AMUL: "mul",
	AMVN: "mvn",
	ANEG: "neg",
	AORR: "orr",
	AROR: "ror",
	ASBC: "sbc",
	ASTR: "str",
	ASTRB: "strb",
	ASTRBI: "strb",
	ASTRH: "strh",
	ASTRI: "str",
	ASUB: "sub",
	ASUB8: "sub",
	ASUBI: "sub",
	ASWI: "swi",
	ATST: "tst",
	AUNDEF: "undefined",
}

func decode(w uint32) int {
	switch {
	case extract(w, 11, 15) == 3:
		// THree-operand ADD/SUB with register or immediate
		switch extract(w, 9, 10) {
		case 0: return AADD
		case 1: return ASUB
		case 2: return AADDI
		case 3: return ASUBI
		}
	case extract(w, 13, 15) == 0:
		// Three-operand shifts
		switch extract(w, 11, 12) {
		case 0: return ALSLI
		case 1: return ALSRI
		case 2: return AASRI
		// case 3: add/sub
		}
	case extract(w, 13, 15) == 0:
		// MOVE/CMP/ADD/SUB with 8-bit immediate
		switch extract(w, 11, 12) {
		case 0: return AMOV8
		case 1: return ACMP8
		case 2: return AADD8
		case 3: return ASUB8
		}
	case extract(w, 10, 15) == 0x10:
		switch extract(w, 6, 9) {
		case 0: return AAND
		case 1: return AEOR
		case 2: return ALSL
		case 3: return ALSR
		case 4: return AASR
		case 5: return AADC
		case 6: return ASBC
		case 7: return AROR
		case 8: return ATST
		case 9: return ANEG
		case 10: return ACMP
		case 11: return ACMN
		case 12: return AORR
		case 13: return AMUL
		case 14: return ABIC
		case 15: return AMVN
		}
	case extract(w, 10, 15) == 0x11:
		switch extract(w, 8, 9) {
		case 0: return AADDHi
		case 1: return ACMPHi
		case 2: return AMOVHi
		case 3: return ABXHi
		}
	case extract(w, 11, 15) == 0x9:
		// PC-relative load
		return ALDRPC
	case extract(w, 12, 15) == 0x5:
		if extract(w, 9, 9) == 0 {
			switch extract(w, 10, 11) {
			case 0: return ASTR
			case 1: return ASTRB
			case 2: return ALDR
			case 3: return ALDRB
			}
		} else {
			switch extract(w, 10, 11) {
			case 0: return ASTRH
			case 1: return ALDSB
			case 2: return ALDRH
			case 3: return ALDSH
			}
		}
	case extract(w, 13, 15) == 0x3:
		switch extract(w, 11, 12) {
		case 0: return ASTRI
		case 1: return ALDRI
		case 2: return ASTRBI
		case 3: return ALDRBI
		}
	case extract(w, 12, 15) == 0xD:
		switch extract(w, 8, 11) {
		case 0: return ABEQ
		case 1: return ABNE
		case 2: return ABCS
		case 3: return ABCC
		case 4: return ABMI
		case 5: return ABPL
		case 6: return ABVS
		case 7: return ABVC
		case 8: return ABHI
		case 9: return ABLS
		case 10: return ABGE
		case 11: return ABLT
		case 12: return ABGT
		case 13: return ABLE
		case 14: return AUNDEF
		case 15: 
			if extract(w, 8, 15) == 0xDF {
				return ASWI
			}
			if extract(w, 8, 15) == 0xBE {
				return ABKPT
			}
		}
	case extract(w, 11, 15) == 0x1E:
		return ABX
	case extract(w, 11, 15) == 0x1C:
		return AB
	}
	return AUNDEF
}

// Extract returns bits low..high of an integer.
func extract(a uint32, low, high uint) uint32 {
	return a <<  (31-high) >> (31-high+low)
}

func decodeUnconditionalBranch(w uint32) int32 {
	off := signextend(extract(w, 0, 10), 11) * 4
	return off
}

func signextend(v uint32, n uint) int32 {
	return int32(v<<(32-n))>>(32-n)
}

// Flow analysis:
// If link branch, stop.
// If conditional branch, append to branchlist. Add backpointer.
// If unconditional branch, same, and stop.

type Node struct {
	Addr int // Addr is the address of the node
	A int // A is the type of instruction
	W uint32 // W is the whole instruction
	To *Node // To records the node which this node branches to
	From []*Node // From records the nodes which can branch to this node
	D int // D is the destination register
	M, N int // M and N are source registers
	I int // I is an immediate value
}

func main() {
	filename := os.Args[1]
	v, err := strconv.ParseInt(os.Args[2], 0, 32)
	addr := int(v)
	base := 0x8<<24
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if addr < base {
		fmt.Fprintf(os.Stderr, "invalid address: %#x\n", addr)
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer f.Close()

	var b [2]byte
	_, err = f.Seek(int64(addr - base), 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	var w uint32
	for {
		_, err = f.Read(b[:])
		if err != nil {
			break
		}
		w = uint32(b[0]) + uint32(b[1])<<8
		a := decode(w)
		if a == AB {
			_, err = f.Read(b[:])
			if err != nil {
				break
			}
			w = uint32(b[0]) + uint32(b[1])<<8 + w<<16
			addr = addr
			fmt.Printf("%08x: %08x     %s\n", addr, w, anames[a])
		} else {
			fmt.Printf("%08x: %04x     %s\n", addr, w, anames[a])
		}
		addr += 4
		if a == ABX || a == ABXHi {
			break
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
