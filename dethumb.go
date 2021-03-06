package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
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
	ANOP
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
	Nop

	Add3
	AddPCSP
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
	ALDSB: "ldsb",
	ALDSH: "ldsh",
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

type Immed uint32

func (n Immed) String() string {
	s := strconv.FormatUint(uint64(n), 16)
	s = strings.ToUpper(s)
	if len(s) > 1 {
		if len(s) % 2 != 0 {
			s = "0" + s
		}
		s = "x" + s
	}
	return s
}

type Reg uint32

func (r Reg) String() string {
	return regnames[r]
}

type Regset uint32

func (r Regset) String() string {
	var s []byte
	s = append(s, '{')
	n := 0
	for i := 0; i < 16; i++ {
		if r&1 == 1 {
			if n > 0 {
				s = append(s, ',')
			}
			s = append(s, []byte(regnames[i])...)
			n++
		}
		r >>= 1
	}
	s = append(s, '}')
	return string(s)
}

func decode(v uint32) (int, int) {
	switch {
	case extract(v, 11, 15) == 0x3:
		// THree-operand ADD/SUB with register or immediate
		switch extract(v, 9, 10) {
		case 0: return AADD, Add3
		case 1: return ASUB, Add3
		case 2:
			if extract(v, 6, 8) == 0 {
				return AMOV, Alu
			}
			return AADD, Add3
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
		case 2:
			if extract(v, 0, 7) == 0xC0 {
				// (really mov r8, r8)
				return ANOP, Nop
			}
			return AMOV, AluHi
		case 3:
			if extract(v, 7, 7) == 0 {
				return ABX, BranchReg
			} else {
				return ABLX, BranchReg
			}
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
		// Get relative address
		switch extract(v, 11, 11) {
		case 0: return AADD, AddPCSP
		case 1: return AADD, AddPCSP
		}
	case extract(v, 12, 15) == 0xB:
		switch extract(v, 8, 11) {
		case 0:
			switch extract(v, 7, 7) {
			case 0: return AADD, AddSP
			case 1: return ASUB, AddSP
			}
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
		// add pc, ...
		// mov pc, ...
		if c == AluHi {
			d := extract(v, 0, 2)
			d += extract(v, 7, 7)<<3
			return d == 15
		}
	case APOP:
		// pop lr
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

func formatAdd3(w io.Writer, a int, v uint32) {
	d := Reg(extract(v, 0, 2))
	s := Reg(extract(v, 3, 5))
	n := extract(v, 6, 8)
	isImmed := extract(v, 10, 10) == 1
	if isImmed {
		n := Immed(n)
		fmt.Fprintf(w, "%s, %s, #%s", d, s, n)
	} else {
		n := Reg(n)
		fmt.Fprintf(w, "%s, %s, %s", d, s, n)
	}
}

func formatAlu(w io.Writer, a int, v uint32) {
	d := Reg(extract(v, 0, 2))
	s := Reg(extract(v, 3, 5))
	fmt.Fprintf(w, "%s, %s", d, s)
}

func formatAluHi(w io.Writer, a int, v uint32) {
	d := Reg(extract(v, 0, 2)) + Reg(extract(v, 7, 7)<<3)
	s := Reg(extract(v, 3, 6))
	fmt.Fprintf(w, "%s, %s", d, s)
}

func formatImmed8(w io.Writer, a int, v uint32) {
	d := Reg(extract(v, 8, 10))
	n := Immed(extract(v, 0, 7))
	fmt.Fprintf(w, "%s, #%s", d, n)
}

func formatShift(w io.Writer, a int, v uint32) {
	d := Reg(extract(v, 0, 2))
	s := Reg(extract(v, 3, 5))
	shift := Immed(extract(v, 6, 10))
	if shift == 0 && a != ALSL {
		shift = 32
	}
	fmt.Fprintf(w, "%s, %s, #%s", d, s, shift)
}

func formatLoadPC(w io.Writer, a int, v uint32, r io.ReaderAt, pos int64) {
	var b [4]byte
	offset := extract(v, 0, 7)
	d := Reg(extract(v, 8, 10))
	pos += 4 + int64(offset)*4
	pos &^= 3
	r.ReadAt(b[:], pos)
	n := uint32(b[0]) + uint32(b[1])<<8 + uint32(b[2])<<16 + uint32(b[3])<<24
	fmt.Fprintf(w, "%s,=#%s", d, Immed(n))
}

func formatLoadSP(w io.Writer, a int, v uint32) {
	n := Immed(extract(v, 0, 7)) * 4
	d := Reg(extract(v, 8, 10))
	if n == 0 {
		fmt.Fprintf(w, "%s,[sp]", d)
	} else {
		fmt.Fprintf(w, "%s,[sp, #%s]", d, n)
	}
}

func formatAddPCSP(w io.Writer, a int, v uint32) {
	n := Immed(extract(v, 0, 7))*4
	d := Reg(extract(v, 8, 10))
	b := Reg(15)
	if extract(v, 11, 11) == 1 {
		b = Reg(13)
	}
	fmt.Fprintf(w, "%s, %s, #%s", d, b, n)
}

func formatAddSP(w io.Writer, a int, v uint32) {
	n := Immed(extract(v, 0, 6)) * 4
	fmt.Fprintf(w, "sp, #%s", n)
}

func formatLoadReg(w io.Writer, a int, v uint32) {
	d := Reg(extract(v, 0, 2))
	b := Reg(extract(v, 3, 5))
	o := Reg(extract(v, 6, 8))
	fmt.Fprintf(w, "%s,[%s, %s]", d, b, o)
}

func formatLoadImmed(w io.Writer, a int, v uint32) {
	d := Reg(extract(v, 0, 2))
	b := Reg(extract(v, 3, 5))
	n := Immed(extract(v, 6, 10))
	switch a {
	case ALDR, ASTR: n *= 4
	case ALDRH, ASTRH: n *= 2
	}
	if n == 0 {
		fmt.Fprintf(w, "%s,[%s]", d, b)
	} else {
		fmt.Fprintf(w, "%s,[%s, #%s]", d, b, n)
	}
}

func formatLoadMultiple(w io.Writer, a int, v uint32) {
	r := Regset(extract(v, 0, 7))
	b := Reg(extract(v, 8, 10))
	fmt.Fprint(w, b, "!,", r)
}

func formatPush(w io.Writer, a int, v uint32) {
	r := extract(v, 0, 7)
	switch a {
	case APUSH: r += extract(v, 8, 8)<<14
	case APOP:  r += extract(v, 8, 8)<<15
	}
	fmt.Fprint(w, Regset(r))
}

func formatGoto(w io.Writer, a int, v uint32, addr uint32) {
	offset := extract(v, 0, 10)
	offset = signextend(offset, 11)
	addr += 4 + offset*2
	fmt.Fprintf(w, "%08X", addr)
}

func formatCall(w io.Writer, a int, v uint32, addr uint32) {
	offset := extract(v, 0, 10) << 1
	offset += extract(v, 16, 26) << 12
	addr += 4 + signextend(offset, 23)
	fmt.Fprintf(w, "%08X", addr)
}

func formatBranch(w io.Writer, a int, v uint32, addr uint32) {
	offset := extract(v, 0, 7)
	offset = signextend(offset, 8)*2
	addr += 4 + offset
	fmt.Fprintf(w, "%08X", addr)
}

func parseBranch(a int, v uint32, addr uint32) uint32 {
	offset := extract(v, 0, 7)
	offset = signextend(offset, 8)
	return addr + 4 + offset*2
}

func parseGoto(a int, v uint32, addr uint32) uint32 {
	offset := extract(v, 0, 10)
	offset = signextend(offset, 11)
	return addr + 4 + offset*2
}

func formatInterrupt(w io.Writer, a int, v uint32) {
	n := Immed(extract(v, 0, 7))
	fmt.Fprint(w, n)
}

var regnames = []string{"r0", "r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9", "r10", "r11", "r12", "sp", "lr", "pc"}

func formatBX(w io.Writer, a int, v uint32) {
	s := extract(v, 3, 6)
	fmt.Fprintf(w, "%s", regnames[s])
}

type Node struct {
	Addr  uint32  // Addr is the address of the node
	V     uint32  // W is the instruction
	From  []*Node // From records the nodes which can branch to this node
	To    *Node   // To records the node this node branches to
	Dest  uint32  // Dest records the address of the node which this node branches to
	Label string  // Label is a name for this node
}

func (n *Node) String() string {
	return strconv.FormatUint(uint64(n.Addr), 16)
}

// Flow analysis:
// If link branch, stop.
// If conditional branch, append to branchlist. Add backpointer.
// If unconditional branch, same, and stop.

// NodeSlice implements sort.Interface
type nodeSlice []*Node

func (s nodeSlice) Len() int           { return len(s) }
func (s nodeSlice) Less(i, j int) bool { return s[i].Addr < s[j].Addr }
func (s nodeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }


func main() {
	base := flag.Int64("base", 8<<24, "address of the start of the file")
	flag.Usage = func() {
		fmt.Println("Usage: dethumb [options] filename address")
		fmt.Println("Disassembles the function at the given address.")
		fmt.Println("Options:")
		fmt.Println("  -base=0x8000000: address of the start of the file")
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	filename := flag.Arg(0)
	addr, err := strconv.ParseInt(flag.Arg(1), 0, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if addr < *base {
		fmt.Fprintf(os.Stderr, "invalid address: %#x\n", addr)
		return
	}
	addr &^= 1

	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer f.Close()

	var alist []*Node
	var amap = make(map[uint32]*Node) // addr => instruction
	var deferred []*Node
	for {
		var b [2]byte
		var n, from *Node
		//fmt.Println(addr, deferred, alist)
		if addr == -1 {
			if len(deferred) == 0 {
				break
			}
			from = deferred[0]
			deferred = append(deferred[:0], deferred[1:]...)
			//fmt.Printf("Deferred: %08X -> %08X\n", from.Addr, from.Dest)
			n = amap[from.Dest]
			if n != nil {
				//fmt.Printf("Found %08X, skipping\n", n.Addr)
				n.From = append(n.From, from)
				from.To = n
				continue
			}
			//fmt.Printf("Didn't find %08X, adding\n", from.Dest)
			n = new(Node)
			n.Addr = from.Dest
			n.From = append(n.From, from)
			from.To = n
			addr = int64(n.Addr)
		} else {
			if amap[uint32(addr)] != nil {
				addr = -1
				continue
			}
			n = new(Node)
			n.Addr = uint32(addr)
		}
		alist = append(alist, n)
		amap[n.Addr] = n

		_, err = f.ReadAt(b[:], addr - *base)
		if err != nil {
			break
		}
		addr += 2
		v := uint32(b[0]) + uint32(b[1])<<8
		a, c := decode(v)
		if extract(v, 11, 15) == 0x1E {
			_, err = f.ReadAt(b[:], addr - *base)
			if err != nil {
				break
			}
			addr += 2
			v = uint32(b[0]) + uint32(b[1])<<8 + v<<16
		}
		n.V = v
		switch c {
		case Branch:
			n.Dest = parseBranch(a, v, n.Addr)
			deferred = append(deferred, n)
		case Goto:
			n.Dest = parseGoto(a, v, n.Addr)
			deferred = append(deferred, n)
		}
		//fmt.Printf("%08X\n", n.Dest)
		if isReturn(a, c, v) || c == Goto {
			addr = -1
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	sort.Sort(nodeSlice(alist))
	var label int
	for _, n := range alist {
		if len(n.From) > 0 {
			n.Label = ".label" + strconv.Itoa(label)
			label++
		}
	}
	var buf bytes.Buffer
	for _, n := range alist {
		v := n.V
		a, c := decode(n.V)
		if n.Label != "" {
			fmt.Fprintf(&buf, "%8s:\n", n.Label)
		}
		if v > 0xFFFF {
			fmt.Fprintf(&buf, "%08X: %08X ", n.Addr, v)
		} else {
			fmt.Fprintf(&buf, "%08X: %04X     ", n.Addr, v)
		}
		fmt.Fprintf(&buf, "%-6s ", anames[a])
		switch c {
		case Alu: formatAlu(&buf, a, v)
		case AluHi: formatAluHi(&buf, a, v)
		case Add3: formatAdd3(&buf, a, v)
		case Immed8: formatImmed8(&buf, a, v)
		case Shift: formatShift(&buf, a, v)
		case Call: formatCall(&buf, a, v, n.Addr)
		//case Goto: formatGoto(&buf, a, v, n.Addr)
		//case Branch: formatBranch(&buf, a, v, n.Addr)
		case BranchReg: formatBX(&buf, a, v)
		case AddPCSP: formatAddPCSP(&buf, a, v)
		case AddSP: formatAddSP(&buf, a, v)
		case LoadPC: formatLoadPC(&buf, a, v, f, int64(n.Addr) - *base)
		case LoadSP: formatLoadSP(&buf, a, v)
		case LoadReg: formatLoadReg(&buf, a, v)
		case LoadImmed: formatLoadImmed(&buf, a, v)
		case LoadMultiple: formatLoadMultiple(&buf, a, v)
		case Push: formatPush(&buf, a, v)
		case Interrupt: formatInterrupt(&buf, a, v)
		case Branch, Goto:
			fmt.Fprintf(&buf, n.To.Label)
		}
		buf.WriteByte('\n')
		buf.WriteTo(os.Stdout)
	}
}
