package screen
import (
	"fmt"
	"syscall"
	"unsafe"
	"os"
)
// Key mappings (ASCII sequences for special keys)
const (
	ESC   = "\x1b"
	//ENTER = "\r"
	ENTER = "\n"
	SPACE = " "
	TAB   = "\t"

	F1    = "\x1bOP"
	F2    = "\x1bOQ"
	F3    = "\x1bOR"
	F4    = "\x1bOS"
	F5    = "\x1b[15~"
	F6    = "\x1b[17~"
	F7    = "\x1b[18~"
	F8    = "\x1b[19~"
	F9    = "\x1b[20~"
	F10   = "\x1b[21~"
	INS   = "\x1b[2~"
	DEL   = "\x1b[3~"
	HOME  = "\x1b[H"
	END   = "\x1b[F"
	PGUP  = "\x1b[5~"
	PGDN  = "\x1b[6~"
		
	UPARR = "\x1b[A"
	DNARR = "\x1b[B"
	RTARR = "\x1b[C"
	LTARR = "\x1b[D"

	CTR_A = "\x01"
	CTR_B = "\x02"
	CTR_C = "\x03"
	CTR_D = "\x04"
	CTR_E = "\x05"
	CTR_F = "\x06"
	CTR_G = "\a"
	CTR_H = "\b"
	CTR_I = "\t"
	CTR_J = "\n"
	CTR_K = "\v"
	CTR_L = "\f"
	CTR_M = "\r"
	CTR_N = "\x0e"
	CTR_O = "\x0f"
	CTR_P = "\x10"
	CTR_Q = "\x11"
	CTR_R = "\x12"
	CTR_S = "\x13"
	CTR_T = "\x14"
	CTR_U = "\x15"
	CTR_V = "\x16"
	CTR_W = "\x17"
	CTR_X = "\x18"
	CTR_Y = "\x19"
	CTR_Z = "\x1a"

	ALT_A = "\x1ba"
	ALT_B = "\x1bb"
	ALT_C = "\x1bc"
	ALT_D = "\x1bd"
	ALT_E = "\x1be"
	ALT_F = "\x1bf"
	ALT_G = "\x1bg"
	ALT_H = "\x1bh"
	ALT_I = "\x1bi"
	ALT_J = "\x1bj"
	ALT_K = "\x1bk"
	ALT_L = "\x1bl"
	ALT_M = "\x1bm"
	ALT_N = "\x1bn"
	ALT_O = "\x1bo"
	ALT_P = "\x1bp"
	ALT_Q = "\x1bq"
	ALT_R = "\x1br"
	ALT_S = "\x1bs"
	ALT_T = "\x1bt"
	ALT_U = "\x1bu"
	ALT_V = "\x1bv"
	ALT_W = "\x1bw"
	ALT_X = "\x1bx"
	ALT_Y = "\x1by"
	ALT_Z = "\x1bz"
)

// Function to read key input
func GetKey() (string, error) {
	var b [6]byte
	n, err := os.Stdin.Read(b[:])
	if err != nil {
		return "", err
	}
	return string(b[:n]), nil
}

func SetRawMode() (func(), error) {
	fd := os.Stdin.Fd()
	var oldState syscall.Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, 
						uintptr(fd), 
						syscall.TCGETS, 
						uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
		return nil, err
	}
	newState := oldState
	newState.Lflag &^= syscall.ICANON | syscall.ECHO
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, 
						uintptr(fd), 
						syscall.TCSETS, 
						uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		return nil, err
	}
	return func(){
		oldState.Lflag |= syscall.ICANON | syscall.ECHO
		syscall.Syscall6(syscall.SYS_IOCTL, 
						uintptr(fd), 
						syscall.TCSETS, 
						uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
		}, nil
}
func GETWINDOWSIZE()(int, int, error){
	fd := int(os.Stdout.Fd())
	// Define winsize struct (same as C struct)
	var wsz struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	// Use ioctl to get terminal size
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, 
			uintptr(fd), uintptr(syscall.TIOCGWINSZ), 
			uintptr(unsafe.Pointer(&wsz)))
	if errno != 0 {
		return 0, 0, errno
	}

	return int(wsz.Col), int(wsz.Row), nil
}
func CLEAR_WIN	(){fmt.Print("[2J");}
func ECHO_OFF   (){fmt.Print("[12h");}
func ECHO_ON    (){fmt.Print("[12I");}
func KEY_LOCK   (){fmt.Print("[2h");}
func KEY_UNLOCK (){fmt.Print("[2I");}
func PAD_NUM    (){fmt.Print(">");}
func PAD_APP    (){fmt.Print("=");}
func TERM_RESET (){fmt.Print("c");}

func COL_132    (){fmt.Print("[?3h");}
func COL_80     (){fmt.Print("[?3I");}
func WARAP_ON   (){fmt.Print("[?7h");}
func WARAP_OFF  (){fmt.Print("[?7I");}
func REPEAT_ON  (){fmt.Print("[?8h");}
func REPEAT_OFF (){fmt.Print("[?8I");}

func CURSOR_HOME     (){fmt.Print("[1;1H");}
func CURSOR_MOVE(x , y int){fmt.Printf("[%d;%dH", y, x);}
func CURSOR_ON       (){fmt.Print("[?25h");}
func CURSOR_OFF      (){fmt.Print("[?25l");}
func UNDERLINE_CURSOR(){fmt.Print("[34h");}
func BLOCK_CURSOR    (){fmt.Print("[34I");}
func SAVE_CURSOR     (){fmt.Print("7");}
func RESTORE_CURSOR  (){fmt.Print("8");}

func ATTR_NONE      (){ fmt.Print("[0m");}
func ATTR_BOLD      (){ fmt.Print("[1m");}
func ATTR_DIM       (){ fmt.Print("[2m");}
func ATTR_UNDERLINE (){ fmt.Print("[4m");}
func ATTR_BLINK     (){ fmt.Print("[5m");}
func ATTR_REVERSE   (){ fmt.Print("[7m");}
func ATTR_NORMAL    (){ fmt.Print("[22m");}
func ATTR_DIMRESET  (){ fmt.Print("[23m");}
func ATTR_UNDERRESET(){ fmt.Print("[24m");}
func ATTR_BLINKRESET(){ fmt.Print("[25m");}
func ATTR_REVRESET  (){ fmt.Print("[27m");}

func SET_TOPBOTTOM(t,b int){fmt.Printf("[%d:%dr", t,b);}
func SCREEN_UP    ()   {fmt.Print("D");}
func SCREEN_DOWN  ()   {fmt.Print("M");}

func CLEAR_LINE_FROMCUR() {fmt.Print("[K");}
func CLEAR_LINE_TOCUR  () {fmt.Print("[1K");}
func CLEAR_LINE        () {fmt.Print("[2K");}

func COLOR_RED      (){fmt.Print("\x1B[31m");}
func COLOR_GREEN    (){fmt.Print("\x1B[32m");}
func COLOR_YELLOW   (){fmt.Print("\x1B[33m");}
func COLOR_BLUE     (){fmt.Print("\x1B[34m");}
func COLOR_MAGENTA  (){fmt.Print("\x1B[35m");}
func COLOR_SCYAN    (){fmt.Print("\x1B[36m");}
func COLOR_WHITE    (){fmt.Print("\x1B[37m");}
func COLOR_RESET    (){fmt.Print("\x1B[0m");}

