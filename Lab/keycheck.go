package main
import (
	"fmt"
	"os"
	sc "centolib/screen"
)
func main(){
    reset, err := sc.SetRawMode()
    if err != nil{
        fmt.Printf("%v\n", err)
        os.Exit(0)
    }
    defer reset()

	for {
		key, err := sc.GetKey()
		if err != nil{
			fmt.Printf("%v\n", err)
			os.Exit(0)
		}

		if key == "q" {break;}
		switch (key) {
		case sc.ESC   :fmt.Printf("ESC  \n"); 
		case sc.ENTER :fmt.Printf("ENTER\n"); 
		case sc.SPACE :fmt.Printf("SPACE\n"); 
		case sc.TAB   :fmt.Printf("TAB  \n"); 
		case sc.F1    :fmt.Printf("F1   \n"); 
		case sc.F2    :fmt.Printf("F2   \n"); 
		case sc.F3    :fmt.Printf("F3   \n"); 
		case sc.F4    :fmt.Printf("F4   \n"); 
		case sc.F5    :fmt.Printf("F5   \n");
		case sc.F6    :fmt.Printf("F6   \n");
		case sc.F7    :fmt.Printf("F7   \n");
		case sc.F9    :fmt.Printf("F9   \n");
		case sc.F10   :fmt.Printf("F10  \n");
		case sc.INS   :fmt.Printf("INS  \n");
		case sc.DEL   :fmt.Printf("DEL  \n");
		case sc.HOME  :fmt.Printf("HOME \n"); 
		case sc.END   :fmt.Printf("END  \n"); 
		case sc.PGUP  :fmt.Printf("PGUP \n");
		case sc.PGDN  :fmt.Printf("PGDN \n"); 
		case sc.UPARR :fmt.Printf("UPARR\n"); 
		case sc.DNARR :fmt.Printf("DNARR\n"); 
		case sc.RTARR :fmt.Printf("RTARR\n"); 
		case sc.LTARR :fmt.Printf("LTARR\n"); 
		case sc.CTL_A :fmt.Printf("CTL_A\n"); 
		case sc.CTL_B :fmt.Printf("CTL_B\n"); 
		case sc.CTL_C :fmt.Printf("CTL_C\n"); 
		case sc.CTL_D :fmt.Printf("CTL_D\n"); 
		case sc.CTL_E :fmt.Printf("CTL_E\n"); 
		case sc.CTL_F :fmt.Printf("CTL_F\n"); 
		case sc.CTL_G :fmt.Printf("CTL_G\n"); 
		case sc.CTL_H :fmt.Printf("CTL_H\n"); 
		//case CTL_I :fmt.Printf("CTL_I\n"); 
		//case CTL_J :fmt.Printf("CTL_J\n"); 
		case sc.CTL_K :fmt.Printf("CTL_K\n"); 
		case sc.CTL_L :fmt.Printf("CTL_L\n"); 
		//case CTL_M :fmt.Printf("CTL_M\n"); 
		case sc.CTL_N :fmt.Printf("CTL_N\n"); 
		case sc.CTL_O :fmt.Printf("CTL_O\n"); 
		case sc.CTL_P :fmt.Printf("CTL_P\n"); 
		case sc.CTL_Q :fmt.Printf("CTL_Q\n"); 
		case sc.CTL_R :fmt.Printf("CTL_R\n"); 
		case sc.CTL_S :fmt.Printf("CTL_S\n"); 
		case sc.CTL_T :fmt.Printf("CTL_T\n"); 
		case sc.CTL_U :fmt.Printf("CTL_U\n"); 
		case sc.CTL_V :fmt.Printf("CTL_V\n"); 
		case sc.CTL_W :fmt.Printf("CTL_W\n"); 
		case sc.CTL_X :fmt.Printf("CTL_X\n"); 
		case sc.CTL_Y :fmt.Printf("CTL_Y\n"); 
		case sc.CTL_Z :fmt.Printf("CTL_Z\n"); 
		case sc.ALT_A :fmt.Printf("ALT_A\n"); 
		case sc.ALT_B :fmt.Printf("ALT_B\n"); 
		case sc.ALT_C :fmt.Printf("ALT_C\n"); 
		case sc.ALT_D :fmt.Printf("ALT_D\n"); 
		case sc.ALT_E :fmt.Printf("ALT_E\n"); 
		case sc.ALT_F :fmt.Printf("ALT_F\n"); 
		case sc.ALT_G :fmt.Printf("ALT_G\n"); 
		case sc.ALT_H :fmt.Printf("ALT_H\n"); 
		case sc.ALT_I :fmt.Printf("ALT_I\n"); 
		case sc.ALT_J :fmt.Printf("ALT_J\n"); 
		case sc.ALT_K :fmt.Printf("ALT_K\n"); 
		case sc.ALT_L :fmt.Printf("ALT_L\n"); 
		case sc.ALT_M :fmt.Printf("ALT_M\n"); 
		case sc.ALT_N :fmt.Printf("ALT_N\n"); 
		case sc.ALT_O :fmt.Printf("ALT_O\n"); 
		case sc.ALT_P :fmt.Printf("ALT_P\n"); 
		case sc.ALT_Q :fmt.Printf("ALT_Q\n"); 
		case sc.ALT_R :fmt.Printf("ALT_R\n"); 
		case sc.ALT_S :fmt.Printf("ALT_S\n"); 
		case sc.ALT_T :fmt.Printf("ALT_T\n"); 
		case sc.ALT_U :fmt.Printf("ALT_U\n"); 
		case sc.ALT_V :fmt.Printf("ALT_V\n"); 
		case sc.ALT_W :fmt.Printf("ALT_W\n"); 
		case sc.ALT_X :fmt.Printf("ALT_X\n"); 
		case sc.ALT_Y :fmt.Printf("ALT_Y\n"); 
		case sc.ALT_Z :fmt.Printf("ALT_Z\n"); 
		default : fmt.Printf("%s\n", key);
/*
		default:
			// Print ASCII code for unknown keys
			fmt.Printf("Key: %q (ASCII: %v)\n", key, bytes)
*/
		}
	}
}
