///
///  structure -> binary, binary -> structure
/// @brief   Test go structure (decimal을 포함한)
/// @file    bindecimal.go
/// @date    2024. 10. 08. (화) 12:35:08 KST
/// @author  Cento
///
package main
import (
	"fmt"
	//"encoding/binary"
	//"bytes"
	Dc "github.com/shopspring/decimal"
)

type BinDecimal struct{
	Val   [20]byte
	ValLen    int32
}

////////////////////////////////////////////////////////////////////////////////
/// BinDecimal의 Decimal 값을 축출한다
/// @fn      func (p *BinDecimal) GetBinDecimal() (Dc.Decimal, err)
/// @param  
/// @return Decimal number, err
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal) GetBinDecimal() (Dc.Decimal, error){
	var value Dc.Decimal
	err := value.UnmarshalBinary(p.Val[:p.ValLen])
	return value, err
}

////////////////////////////////////////////////////////////////////////////////
/// BinDecimal의 Decimal 값을 string으로 축출한다
/// @fn      func (p *BinDecimal) GetBinDecimalString()(string,  error)
/// @param  
/// @return  Decimal string, err
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal) GetBinDecimalString()(string,  error) {
	var value Dc.Decimal
	err := value.UnmarshalBinary(p.Val[:p.ValLen])
	return value.String(), err
}

////////////////////////////////////////////////////////////////////////////////
/// decimal string을 BinDecimal에 넣기
/// @fn      func (p *BinDecimal) SetBinDecimalFromString(string num)
///                     (int32, error)
/// @param   num : decimal string 
/// @return  Marshal된 길이, error
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal) SetBinDecimalFromString(num string)(int32, error){
	/// convert string to Decimal
	dcnum , err := Dc.NewFromString(num)
	if err != nil {
		return 0, err
	}
	// Marshalling Decimal(convert to binary)
	binnum , _:= dcnum.MarshalBinary()

	// Copy binary
	p.ValLen = int32(len(binnum))
	copy(p.Val[:], binnum)

	return p.ValLen, nil
}

////////////////////////////////////////////////////////////////////////////////
/// decimal을 BinDecimal에 넣기
/// @fn      func (p *BinDecimal) SetBinDecimalFromDecimal(num Dc.Decimal)
///               (int32, error)
/// @param   num : decimal  
/// @return  Marshal된 길이, error
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal) SetBinDecimalFromDecimal(num Dc.Decimal)(int32, error){
	// Marshalling Decimal(convert to binary)
	binnum , _:= num.MarshalBinary()

	// Copy binary
	p.ValLen = int32(len(binnum))
	copy(p.Val[:], binnum)

	return p.ValLen, nil
}



func main(){
	var d1, d2 BinDecimal
	var dc1 Dc.Decimal

	dc1,_ = Dc.NewFromString("999999999999.9999999999")
	
	rtn, err := d1.SetBinDecimalFromString("988888888888.123123123123")	
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println("-. SetBinDecimalFromString(\"988888888888.123123123123\"")	
	fmt.Printf("Return len : %d\nBinDecimal : %v\n\n", rtn, d1)

	dc1, err = d1.GetBinDecimal()
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println("-. GetBinDecimal()")	
	fmt.Printf("Decimal : %v\n\n", dc1)

	str, err:= d1.GetBinDecimalString()
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println("-. GetBinDecimalString()")	
	fmt.Printf("Decimal string : %s\n\n", str)

	rtn, err = d2.SetBinDecimalFromDecimal(dc1)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println("-. SetBinDecimalFromDecimal():", dc1)	
	fmt.Println("Return : ", rtn)
	fmt.Println("BinDecimal : ", d2)

}
