package main
import (
	"github.com/dixonwille/wmenu"
	//_ "github.com/mattn/go-sqlite3"
	"github.com/projectdiscovery/gologger"
)
func handleFunc(_opts []wmenu.Opt){
	switch _opts[0].Value{
		case 0 :
			gologger.Info().Msgf("Case 0")	
		case 1 :
			gologger.Info().Msgf("Case 1")	
		case 5 : 
			gologger.Info().Msgf("Exit : %d", _opts[0].Value)
	}
}
func main(){
	menu := wmenu.NewMenu("\n메뉴를 선택하시오")
	menu.Action(func (_opts []wmenu.Opt) error {handleFunc(_opts);return nil})
	menu.Option("Case 0", 0, false, nil)
	menu.Option("Case 1", 1, false, nil)
	menu.Option("Case 5", 5, false, nil)
	menu.Run()
}

