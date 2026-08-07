package terminalview

const (
	terminalEscape byte = 0x1b
	terminalBell   byte = 0x07
)

type terminalStreamState uint8

const (
	terminalStreamText terminalStreamState = iota
	terminalStreamEscape
	terminalStreamCSI
	terminalStreamOSC
	terminalStreamOSCEscape
)

// terminalStreamFilter removes terminal metadata sequences that FLTK's native
// Fl_Terminal does not implement and would otherwise render as visible garbage.
// Display-oriented CSI sequences remain byte-for-byte intact. State survives
// arbitrary PTY chunk boundaries.
type terminalStreamFilter struct {
	state terminalStreamState
	csi   []byte
}

func (f *terminalStreamFilter) Reset() {
	f.state = terminalStreamText
	f.csi = f.csi[:0]
}

func (f *terminalStreamFilter) Filter(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	out := make([]byte, 0, len(data))
	for _, b := range data {
		switch f.state {
		case terminalStreamText:
			if b == terminalEscape {
				f.state = terminalStreamEscape
			} else {
				out = append(out, b)
			}
		case terminalStreamEscape:
			switch b {
			case ']':
				f.state = terminalStreamOSC
			case '[':
				f.csi = append(f.csi[:0], terminalEscape, '[')
				f.state = terminalStreamCSI
			default:
				out = append(out, terminalEscape, b)
				f.state = terminalStreamText
			}
		case terminalStreamCSI:
			f.csi = append(f.csi, b)
			if b >= 0x40 && b <= 0x7e {
				if !isUnsupportedPrivateMode(f.csi) {
					out = append(out, f.csi...)
				}
				f.csi = f.csi[:0]
				f.state = terminalStreamText
			}
		case terminalStreamOSC:
			switch b {
			case terminalBell:
				f.state = terminalStreamText
			case terminalEscape:
				f.state = terminalStreamOSCEscape
			}
		case terminalStreamOSCEscape:
			switch b {
			case '\\', terminalBell:
				f.state = terminalStreamText
			case terminalEscape:
				// Remain here so repeated ESC bytes cannot leak OSC payload.
			default:
				f.state = terminalStreamOSC
			}
		}
	}
	return out
}

func isUnsupportedPrivateMode(sequence []byte) bool {
	if len(sequence) < 4 || sequence[0] != terminalEscape || sequence[1] != '[' || sequence[2] != '?' {
		return false
	}
	final := sequence[len(sequence)-1]
	return final == 'h' || final == 'l'
}
