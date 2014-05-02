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
	ALDMIA
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
	APOP
	APUSH
	AROR
	ASBC
	ASTMIA
	ASTR
	ASTRB
	ASTRH
	ASUB
	ASWI
	ATST

	AMax
)

const (
	Undefined = iota

	Add3
	AddSP
	Alu
	AluHi
	Branch
	BranchReg
	Call
	Goto
	Immed8
	Interrupt
	Load
	LoadImmed
	LoadMultiple
	LoadPC
	LoadReg
	LoadSP
	Push
	Shift
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
	ACMP: "cmp",
	ACMN: "cmn",
	AEOR: "eor",
	ALDMIA: "ldmia",
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
	APOP: "pop",
	APUSH: "push",
	AROR: "ror",
	ASBC: "sbc",
	ASTMIA: "stmia",
	ASTR: "str",
	ASTRB: "strb",
	ASTRH: "strh",
	ASUB: "sub",
	ASWI: "swi",
	ATST: "tst",
	AUNDEF: "undefined",
}

func decode(v uint32) (int, int) {
	switch {
	case extract(v, 11, 15) == 0x3:
		// THree-operand ADD/SUB with register or immediate
		switch extract(v, 9, 10) {
		case 0: return AADD, Add3
		case 1: return ASUB, Add3
		case 2: return AADD, Add3
		case 3: return ASUB, Add3
		}
	case extract(v, 13, 15) == 0x0:
		// Three-operand shifts
		switch extract(v, 11, 12) {
		case 0: return ALSL, Shift
		case 1: return ALSR, Shift
		case 2: return AASR, Shift
		// case 3: add/sub
		}
	case extract(v, 13, 15) == 0x1:
		// MOVE/CMP/ADD/SUB with 8-bit immediate
		switch extract(v, 11, 12) {
		case 0: return AMOV, Immed8
		case 1: return ACMP, Immed8
		case 2: return AADD, Immed8
		case 3: return ASUB, Immed8
		}
	case extract(v, 10, 15) == 0x10:
		switch extract(v, 6, 9) {
		case 0: return AAND, Alu
		case 1: return AEOR, Alu
		case 2: return ALSL, Alu
		case 3: return ALSR, Alu
		case 4: return AASR, Alu
		case 5: return AADC, Alu
		case 6: return ASBC, Alu
		case 7: return AROR, Alu
		case 8: return ATST, Alu
		case 9: return ANEG, Alu
		case 10: return ACMP, Alu
		case 11: return ACMN, Alu
		case 12: return AORR, Alu
		case 13: return AMUL, Alu
		case 14: return ABIC, Alu
		case 15: return AMVN, Alu
		}
	case extract(v, 10, 15) == 0x11:
		switch extract(v, 8, 9) {
		case 0: return AADD, AluHi
		case 1: return ACMP, AluHi
		case 2: return AMOV, AluHi
		}
		if extract(v, 7, 7) == 0 {
			return ABX, BranchReg
		} else {
			return ABLX, BranchReg
		}
	case extract(v, 11, 15) == 0x9:
		// PC-relative load
		return ALDR, LoadPC
	case extract(v, 12, 15) == 0x5:
		// Load/store with register offset
		if extract(v, 9, 9) == 0 {
			switch extract(v, 10, 11) {
			case 0: return ASTR, LoadReg
			case 1: return ASTRB, LoadReg
			case 2: return ALDR, LoadReg
			case 3: return ALDRB, LoadReg
			}
		} else {
			switch extract(v, 10, 11) {
			case 0: return ASTRH, LoadReg
			case 1: return ALDSB, LoadReg
			case 2: return ALDRH, LoadReg
			case 3: return ALDSH, LoadReg
			}
		}
	case extract(v, 13, 15) == 0x3:
		// Load/store with immediate offset
		switch extract(v, 11, 12) {
		case 0: return ASTR, LoadImmed
		case 1: return ALDR, LoadImmed
		case 2: return ASTRB, LoadImmed
		case 3: return ALDRB, LoadImmed
		}
	case extract(v, 12, 15) == 0x8:
		switch extract(v, 11, 11) {
		case 0: return ASTRH, LoadImmed
		case 1: return ALDRH, LoadImmed
		}
	case extract(v, 12, 15) == 0x9:
		// Load/store SP-relative
		switch extract(v, 11, 11) {
		case 0: return ASTR, LoadSP
		case 1: return ALDR, LoadSP
		}
	case extract(v, 12, 15) == 0xA:
		switch extract(v, 11, 11) {
		case 0: return AADD, LoadPC
		case 1: return AADD, LoadSP
		}
	case extract(v, 12, 15) == 0xB:
		switch extract(v, 8, 11) {
		case 0: return AADD, AddSP
		case 4, 5: return APUSH, Push
		case 12, 13: return APOP, Push
		case 15: return ABKPT, Interrupt
		}
		return AUNDEF, Undefined
	case extract(v, 12, 15) == 0xC:
		switch extract(v, 11, 11) {
		case 0: return ASTMIA, LoadMultiple
		case 1: return ALDMIA, LoadMultiple
		}
	case extract(v, 12, 15) == 0xD:
		switch extract(v, 8, 11) {
		case 0: return ABEQ, Branch
		case 1: return ABNE, Branch
		case 2: return ABCS, Branch
		case 3: return ABCC, Branch
		case 4: return ABMI, Branch
		case 5: return ABPL, Branch
		case 6: return ABVS, Branch
		case 7: return ABVC, Branch
		case 8: return ABHI, Branch
		case 9: return ABLS, Branch
		case 10: return ABGE, Branch
		case 11: return ABLT, Branch
		case 12: return ABGT, Branch
		case 13: return ABLE, Branch
		case 14: return AUNDEF, Undefined
		case 15: return ASWI, Interrupt
		}
	case extract(v, 11, 15) == 0x1C:
		return AB, Goto
	case extract(v, 11, 15) == 0x1E:
		return ABL, Call
	case extract(v, 11, 15) == 0x1F:
		return ABL, Call
	case extract(v, 11, 15) == 0x1D:
		return ABLX, Call
	}
	return AUNDEF, Undefined
}

// IsReturn reports whether the instruction returns execution to a calling function.
func isReturn(a int, c int, v uint32 ) bool {
	switch a {
	case ABX:
		return true
	case AADD, AMOV:
		if c == AluHi {
			d := extract(v, 0, 2)
			d += extract(v, 7, 7)<<3
			return d == 15
		}
	case APOP:
		return extract(v, 8, 8) == 1
	}
	return false
}

// Extract returns bits low..high of an integer.
func extract(v uint32, low, high uint) uint32 {
	return v <<  (31-high) >> (31-high+low)
}

func signextend(v uint32, n uint) uint32 {
	return uint32(int32(v<<(32-n))>>(32-n))
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
}

func formatGoto(w io.Writer, a int, v uint32, addr uint32) {
	offset := extract(v, 0, 10)
	offset = signextend(offset, 11)
	addr += 4 + offset*2
	fmt.Fprintf(w, "%08x", addr)
}

func formatBL(w io.Writer, a int, v uint32, addr uint32) {
	offset := extract(v, 0, 10) << 1
	offset += extract(v, 16, 26) << 12
	addr += 4 + signextend(offset, 23)
	fmt.Fprintf(w, "%08x", addr)
}

func formatBranch(w io.Writer, a int, v uint32, addr uint32) {
	offset := extract(v, 0, 7)
	offset = signextend(offset, 8)*2
	addr += 4 + offset
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
		a, c := decode(v)
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
		switch c {
		case Alu: formatAlu(&buf, a, v)
		case Call: formatBL(&buf, a, v, uint32(addr))
		case Goto: formatGoto(&buf, a, v, uint32(addr))
		case Branch: formatBranch(&buf, a, v, uint32(addr))
		case BranchReg: formatBX(&buf, a, v)
		}
		fmt.Fprint(&buf, "\n")
		buf.WriteTo(os.Stdout)
		addr += wlen
		// Return instructions:
		// pop lr
		// mov pc, X
		// bx
		if isReturn(a, c, v)  {
			break
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}
