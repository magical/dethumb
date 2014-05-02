package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	AUNDEF = iota
	AADC
	AADD
	AAND
	AASR
	AB
	ABCC
	ABCS
	ABEQ
	ABGE
	ABGT
	ABHI
	ABIC
	ABKPT
	ABL
	ABLE
	ABLS
	ABLT
	ABLX
	ABMI
	ABNE
	ABPL
	ABVC
	ABVS
	ABX
	ACMN
	ACMP
	AEOR
	ALDR
	ALDRB
	ALDRH
	ALDRPC
	ALDSB
	ALDSH
	ALSL
	ALSR
	AMOV
	AMUL
	AMVN
	ANEG
	AORR
	AROR
	ASBC
	ASTR
	ASTRB
	ASTRH
	ASUB
	ASWI
	ATST

	AMax
)

var anames = [AMax]string{
	AADC: "adc",
	AADD: "add",
	AAND: "and",
	AASR: "asr",
	AB: "b",
	ABCC: "blo",
	ABCS: "bhs",
	ABEQ: "beq",
	ABGE: "bge",
	ABGT: "bgt",
	ABHI: "bhi",
	ABIC: "bic",
	ABKPT: "bkpt",
	ABL: "bl",
	ABLE: "ble",
	ABLS: "bls",
	ABLX: "blx",
	ABLT: "blt",
	ABMI: "bmi",
	ABNE: "bne",
	ABPL: "bpl",
	ABVC: "bvc",
	ABVS: "bvs",
	ABX: "bx",
	ACMN: "cmn",
	AEOR: "eor",
	ALDR: "ldr",
	ALDRB: "ldrb",
	ALDRH: "ldrh",
	ALDRPC: "ldr",
	ALDSB: "lds",
	ALDSH: "lds",
	ALSL: "lsl",
	ALSR: "lsr",
	AMOV: "mov",
	AMUL: "mul",
	AMVN: "mvn",
	ANEG: "neg",
	AORR: "orr",
	AROR: "ror",
	ASBC: "sbc",
	ASTR: "str",
	ASTRB: "strb",
	ASTRH: "strh",
	ASUB: "sub",
	ASWI: "swi",
	ATST: "tst",
	AUNDEF: "undefined",
}

func decode(v uint32) int {
	switch {
	case extract(v, 11, 15) == 3:
		// THree-operand ADD/SUB with register or immediate
		switch extract(v, 9, 10) {
		case 0: return AADD
		case 1: return ASUB
		case 2: return AADD
		case 3: return ASUB
		}
	case extract(v, 13, 15) == 0:
		// Three-operand shifts
		switch extract(v, 11, 12) {
		case 0: return ALSL
		case 1: return ALSR
		case 2: return AASR
		// case 3: add/sub
		}
	case extract(v, 13, 15) == 0:
		// MOVE/CMP/ADD/SUB with 8-bit immediate
		switch extract(v, 11, 12) {
		case 0: return AMOV
		case 1: return ACMP
		case 2: return AADD
		case 3: return ASUB
		}
	case extract(v, 10, 15) == 0x10:
		switch extract(v, 6, 9) {
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
	case extract(v, 10, 15) == 0x11:
		switch extract(v, 8, 9) {
		case 0: return AADD
		case 1: return ACMP
		case 2: return AMOV
		}
		if extract(v, 7, 7) == 0 {
			return ABX
		} else {
			return ABLX
		}
	case extract(v, 11, 15) == 0x9:
		// PC-relative load
		return ALDRPC
	case extract(v, 12, 15) == 0x5:
		if extract(v, 9, 9) == 0 {
			switch extract(v, 10, 11) {
			case 0: return ASTR
			case 1: return ASTRB
			case 2: return ALDR
			case 3: return ALDRB
			}
		} else {
			switch extract(v, 10, 11) {
			case 0: return ASTRH
			case 1: return ALDSB
			case 2: return ALDRH
			case 3: return ALDSH
			}
		}
	case extract(v, 13, 15) == 0x3:
		switch extract(v, 11, 12) {
		case 0: return ASTR
		case 1: return ALDR
		case 2: return ASTRB
		case 3: return ALDRB
		}
	case extract(v, 12, 15) == 0xD:
		switch extract(v, 8, 11) {
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
			if extract(v, 8, 15) == 0xDF {
				return ASWI
			}
			if extract(v, 8, 15) == 0xBE {
				return ABKPT
			}
		}
	case extract(v, 11, 15) == 0x1C:
		return AB
	case extract(v, 11, 15) == 0x1F:
		return ABL
	case extract(v, 11, 15) == 0x1D:
		return ABX
	}
	return AUNDEF
}

// Extract returns bits low..high of an integer.
func extract(a uint32, low, high uint) uint32 {
	return a <<  (31-high) >> (31-high+low)
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

func formatAlu(w io.Writer, a int, v uint32) {
	var d, m uint32
	if extract(v, 11, 15) == 0x3 {
		//immediate
	} else if extract(v, 13, 15) == 0x1 {
		// 8-bit immediate
	} else if extract(v, 13, 15) == 0x0 {
		
	} else if extract(v, 13, 15) == 0x03 {
		// 3 op
	} else if extract(v, 10, 15) == 0x10 {
		// normal
		d = extract(v, 0, 2)
		m = extract(v, 3, 5)
	} else if extract(v, 10, 15) == 0x11 {
		// High registers
	}
	_ = d
	_ = m
}

func formatBL(w io.Writer, a int, v uint32, addr uint32) {
	offset := extract(v, 0, 10) << 1
	offset += extract(v, 16, 26) << 12
	addr += uint32(signextend(offset, 22))
	fmt.Fprintf(w, "%08x", addr)
}

var regnames = []string{"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10", "r11", "r12", "sp", "lr", "pc"}

func formatBX(w io.Writer, a int, v uint32) {
	s := extract(v, 3, 6)
	fmt.Fprintf(w, "%s", regnames[s])
}

func main() {
	filename := os.Args[1]
	x, err := strconv.ParseInt(os.Args[2], 0, 32)
	addr := int(x)
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
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	var v uint32
	var buf bytes.Buffer
	for {
		_, err = f.ReadAt(b[:], int64(addr - base))
		if err != nil {
			break
		}
		wlen := 2
		v = uint32(b[0]) + uint32(b[1])<<8
		a := decode(v)
		if extract(v, 11, 15) == 0x1E {
			_, err = f.ReadAt(b[:], int64(addr - base + 2))
			if err != nil {
				break
			}
			v = uint32(b[0]) + uint32(b[1])<<8 + v<<16
			wlen = 4
			fmt.Fprintf(&buf, "%08x: %08x ", addr, v)
		} else {
			fmt.Fprintf(&buf, "%08x: %04x     ", addr, v)
		}
		fmt.Fprintf(&buf, "%s ", anames[a])
		switch a {
		case AMOV, AAND, ATST, ABIC, AORR, AEOR, AADD, AADC, ASUB, ASBC, ANEG, ACMP, ACMN, AMUL:
			formatAlu(&buf, a, v)
		case ABL:
			formatBL(&buf, a, v, uint32(addr))
		case ABX, ABLX:
			formatBX(&buf, a, v)
		}
		fmt.Fprint(&buf, "\n")
		buf.WriteTo(os.Stdout)
		addr += wlen
		// Return instructions:
		// pop lr
		// mov pc, X
		// bx
		if a == ABX  {
			break
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
