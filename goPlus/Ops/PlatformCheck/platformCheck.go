// Platform 정보를 CHECK하는 Program
//
package main
import (
	"fmt"
	"time"
	sc "centolib/screen"
	"os"
	"os/signal"
	"syscall"
)
type Menu struct{
	Name     string
	Function func()bool
	Comment string
}
var (
	MaxY, MaxX, AppTop, AppBottom, TitlePos int
	MenuList []Menu
)
func initMenu(){
	MenuList =  []Menu{
{"주문상태", OrderCondition , "각 코인별 주문 가능 상태를 조회한다"},
{"PROXY 정보", ProxyMenu, "Proxy.yaml에 있는 Proxy의 상태를 조회한다"},
{"코인별주문정보", CoinOrderStatus, "주문가능코인의 주문File을 점검한다"},
{"MENU3", func()bool{return true;}, "Menu 3"},
{"MENU4", func()bool{return true;}, "Menu 4"},
	}	
}
func clearArea(x, y, width, height int) {
    // Move cursor to position (x, y)
    fmt.Printf("\033[%d;%dH", y, x)
    
    // Loop over the area and clear each line
    for i := 0; i < height; i++ {
        fmt.Printf("\033[K") // Clear the current line
        fmt.Print("\033[B")  // Move down to the next line
    }
    
    // Reset cursor to the initial position after clearing
    fmt.Printf("\033[%d;%dH", y+height, x)
}
func displayTime(){
	sc.SAVE_CURSOR()
	sc.CURSOR_MOVE(1,1)
	sc.CLEAR_LINE()
	sc.ATTR_BOLD()
	str := time.Now().Format("현재시각 : 2006년 01월 02일 - 15:04:05")
	fmt.Printf("%s,  X(%d) Y(%d) T(%d) B(%d)", 
				str, MaxX, MaxY, AppTop, AppBottom)
	sc.ATTR_NONE()
	sc.RESTORE_CURSOR()
}
// DrawHLine은 지정 좌표에 수평 구분선을 출력한다.
func DrawHLine(front bool, x, y int){
	sc.SAVE_CURSOR()
	sc.CURSOR_MOVE(x,y)
	var i int
	var str string
	if front { 
		str = "├"
		i = x + 1
	} else{ 
		i = x
	}
	for ; i <= MaxX; i++{
		str = str + "─"
	}
	fmt.Printf("%s", str)
	sc.RESTORE_CURSOR()
}
func outStrXy(cl bool, x, y int, format string, a ...interface{}){
	sc.SAVE_CURSOR()
	sc.CURSOR_MOVE(x,y)
	if cl == true {
		sc.CLEAR_LINE_FROMCUR()
	}
	fmt.Printf(format, a...)
	sc.RESTORE_CURSOR()
}
func messageLine(format string, a ...interface{}){
	sc.SAVE_CURSOR()
	sc.CURSOR_MOVE(1,MaxY)
	sc.CLEAR_LINE()
	sc.ATTR_BOLD()
	sc.COLOR_SCYAN()
	fmt.Printf(format, a...)
	sc.RESTORE_CURSOR()
	sc.ATTR_NONE()
	sc.COLOR_RESET()
}

// MenuSelect는 메뉴 항목을 선택 상태(굵은 노란색)로 화면에 출력한다.
func MenuSelect(pos int, menu string){
	sc.ATTR_BOLD()
	sc.COLOR_YELLOW()

	fmt.Printf("[%d;1H%s", AppTop + pos, menu)

	sc.ATTR_NONE()
	sc.COLOR_RESET()
}
// MenuUnSelect는 메뉴 항목을 비선택 상태(일반 색상)로 화면에 출력한다.
func MenuUnSelect(pos int, menu string){
	sc.ATTR_NONE()
	sc.COLOR_RESET()
	fmt.Printf("[%d;1H%s", AppTop + pos, menu)
}
func displayTitle(title string){
	sc.SAVE_CURSOR()
	sc.CURSOR_MOVE(10,TitlePos)
	sc.CLEAR_LINE_FROMCUR()
	sc.ATTR_BOLD()
	sc.COLOR_BLUE()
	fmt.Print("====== ", title , " ======")
	sc.RESTORE_CURSOR()
	sc.ATTR_NONE()
	sc.COLOR_RESET()
}
func displayMainMenu(){
	sc.SAVE_CURSOR()
	sc.CLEAR_WIN()

	displayTime()
	DrawHLine(false, 1, AppTop - 1)
	for i, menu := range MenuList{
		outStrXy(true, 1, AppTop + i, "%s", menu.Name)
	}
	displayTitle("Menu Home")

	sc.RESTORE_CURSOR()
}
func initApplication()error{
	x, y, err := sc.GETWINDOWSIZE()
	if err != nil{ return err; }
	MaxY = y
	MaxX = x
	TitlePos = 2
	AppTop = TitlePos + 2  // Title line + H line
	AppBottom = MaxY - 2
	initMenu()
	return nil
}
func main(){
	err := initApplication()
	if err != nil {
		fmt.Printf("initApplication() : %v\n", err)
		os.Exit(0)
	}
	reset, err := sc.SetRawMode()
	if err != nil{
		fmt.Printf("%v\n", err)
		os.Exit(0)
	}
	defer reset()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	go func (){
		for{
			select {
				case <-sigs :
				x, y, err := sc.GETWINDOWSIZE()
				if err != nil{continue;}
				MaxY = y
				MaxX = x
				TitlePos = 2
				AppTop = TitlePos + 2  // Title line + H line
				AppBottom = MaxY - 2
				messageLine("윈도우 변경 : X(%d) Y(%d)", x, y)		
			}
		}
	}()

	// Main Loop
	sc.CURSOR_OFF()
	sc.WARAP_ON()
	var run = true
	var pos int
	pos = 0
	displayMainMenu()	
	MenuSelect(pos, MenuList[pos].Name)
	for run {
		char, err := sc.GetKey()
		if err != nil{
			fmt.Printf("%v\n", err)
			os.Exit(0)
		}
		switch {
			case char == "q" || char == "Q" :
				run = false
			case char == sc.UPARR || char == sc.LTARR:
				MenuUnSelect(pos, MenuList[pos].Name)
				pos--
				if pos < 0{pos = len(MenuList) -1;}
				MenuSelect(pos, MenuList[pos].Name)
			case char == sc.DNARR || char == sc.RTARR:
				MenuUnSelect(pos, MenuList[pos].Name)
				pos++
				if pos >= len(MenuList){pos = 0;}
				MenuSelect(pos, MenuList[pos].Name)
			case char == sc.CTR_L :
				displayMainMenu()
				MenuSelect(pos, MenuList[pos].Name)
			case char == sc.ENTER :
				displayTitle(MenuList[pos].Name)
				fg := MenuList[pos].Function()
				displayMainMenu()
				MenuSelect(pos, MenuList[pos].Name)
				if fg == false{run = false;}
					
		}
		messageLine("POS(%d)%s:%s ", pos, 
					MenuList[pos].Name, 
					MenuList[pos].Comment)
	}
	sc.CLEAR_WIN()
	sc.CURSOR_ON()
}
