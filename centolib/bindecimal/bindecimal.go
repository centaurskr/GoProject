/// @brief   Decimal을 File or Mapped file에 저장하기 위한 lib
/// @file    bindecimal.go
/// @date    2024. 10. 08. (화) 15:13:39 KST
/// @author  Cento
///
package bindecimal
import (
    Dc "github.com/shopspring/decimal"
)

type BinDecimal struct{
    Val   [36]byte
    ValLen    int
}

const (
	ROUND_UP = iota  // 올림
	ROUND_DN         // 버림
	ROUND            // 반올림
	NOROUND          // NON
)
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func New() BinDecimal{
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func New() BinDecimal{
	var bindec BinDecimal
	bindec.SetFromString("0")
	return bindec
}
////////////////////////////////////////////////////////////////////////////////
/// BinDecimal의 Decimal 값을 축출한다
/// @fn      func (p *BinDecimal) Decimal() (Dc.Decimal, err)
/// @param  
/// @return Decimal number
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal) Decimal() Dc.Decimal{
    var value Dc.Decimal
    value.UnmarshalBinary(p.Val[:p.ValLen])
    return value
}

////////////////////////////////////////////////////////////////////////////////
/// BinDecimal의 Decimal 값을 string으로 축출한다
/// @fn      func (p *BinDecimal) String()string
/// @param  
/// @return  Decimal string
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal)String()string {
    var value Dc.Decimal
    value.UnmarshalBinary(p.Val[:p.ValLen])
    return value.String()
}

func NewFromString(num string)(BinDecimal, int){
	bindec := New()
	ecode := bindec.SetFromString(num)
	return bindec, ecode	
}
////////////////////////////////////////////////////////////////////////////////
/// decimal string을 BinDecimal에 넣기
/// @fn      func (p *BinDecimal) SetFromString(string num)
///                     (int, error)
/// @param   num : decimal string 
/// @return  error code -1(Outof lange), 0(Success)
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal) SetFromString(num string)int{
    /// convert string to Decimal
    dcnum , _ := Dc.NewFromString(num)
    // Marshalling Decimal(convert to binary)
    binnum , _:= dcnum.MarshalBinary()

    // Copy binary
    p.ValLen = int(len(binnum))
	if p.ValLen > len(p.Val) {return -1}
    copy(p.Val[:], binnum)
	return 0
}

////////////////////////////////////////////////////////////////////////////////
/// decimal을 BinDecimal에 넣기
/// @fn      func (p *BinDecimal) SetFromDecimal(num Dc.Decimal)
/// @param   num : decimal  
/// @return  Marshal된 길이, error
////////////////////////////////////////////////////////////////////////////////
func (p *BinDecimal) SetFromDecimal(num Dc.Decimal)int{
    // Marshalling Decimal(convert to binary)
    binnum , _:= num.MarshalBinary()

    // Copy binary
    p.ValLen = int(len(binnum))
	if p.ValLen > len(p.Val) {return -1}
    copy(p.Val[:], binnum)
	return 0
}
func (p *BinDecimal) Set(a interface{})int{
	val,_ := getDecimal(a)

    binnum , _:= val.MarshalBinary()
    // Copy binary
    p.ValLen = int(len(binnum))
	if p.ValLen > len(p.Val) {return -1}
    copy(p.Val[:], binnum)
	return 0
}
////////////////////////////////////////////////////////////////////////////////
/// Local function
/// @fn     : func getDecimal(val interface {})(Dc.Decimal, int){
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func getDecimal(val interface {})(Dc.Decimal, int){
	switch v := val.(type){
		case BinDecimal :
			return v.Decimal(), 0
		case Dc.Decimal :
			return v, 0
		case string :
			dec, _ := Dc.NewFromString(v)
			return dec, 0
		case float32 :
			dec := Dc.NewFromFloat32(v)
			return dec, 0
		case float64 :
			dec := Dc.NewFromFloat(v)
			return dec, 0
		case int64 :
			dec := Dc.NewFromInt(v)
			return dec, 0
		default :
			dec :=Dc.New(0,0)
			return dec, -1
	}
}
////////////////////////////////////////////////////////////////////////////////
/// Local function
/// @fn     : func roundFunction(val Dc.Decimal,mode,place int)
///               (Dc.Decimal, int)
/// @param  :
/// @return :
////////////////////////////////////////////////////////////////////////////////
func roundFunction(val Dc.Decimal, mode, place int)(Dc.Decimal, int){
	switch mode {
		case ROUND_UP :
			return val.RoundUp(int32(place)), 0
		case ROUND_DN :
			return val.RoundDown(int32(place)), 0
		case ROUND :
			return val.Round(int32(place)), 0
		case NOROUND :
			return val, 0
		default : return val, -1
	}
}
////////////////////////////////////////////////////////////////////////////////
/// 0.0 인가?
/// @fn     : func DecimalIsZero(a interface{}) bool{
/// @param  :
/// @return : true : Zero, false : 0아님
////////////////////////////////////////////////////////////////////////////////
func DecimalIsZero(a interface{}) bool{
	val, _ := getDecimal(a)
	return val.IsZero()
}
////////////////////////////////////////////////////////////////////////////////
/// Decimal Equal function
/// @fn     : func DecimalEqual(a,b interface{})int{
/// @param  :
/// @return : Equal(1), Not Equal(0), a value error(-1), b value error(-2)
////////////////////////////////////////////////////////////////////////////////
func DecimalEqual(a,b interface{})int{
	aval, ecode := getDecimal(a)
	if ecode !=0 {return -1}
	bval, ecode := getDecimal(b)
	if ecode !=0 {return -2}
	rtn := aval.Equal(bval)
	if rtn == false {return 0}
	return 1
}
////////////////////////////////////////////////////////////////////////////////
///
/// @fn     : func DecimalCompare(a, b interface{}) int {
/// @param  :
/// @return : a > b (1), a==b(0) a < b(-1)
////////////////////////////////////////////////////////////////////////////////
func DecimalCompare(a, b interface{}) int {
	aval, _ := getDecimal(a)
	bval, _ := getDecimal(b)
	return aval.Compare(bval)
}
func DecimalMax(a, b interface{})interface{}{
	rtn := DecimalCompare(a,b)
	if rtn == 1{ return a;} // a >  b
	return b         // b >= a
}
///// Add Function /////////////////////
func (p *BinDecimal)Add(val interface{}, mode, pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Add(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	bindec.SetFromDecimal(dec)
	return bindec, 0
}
func (p *BinDecimal)AddOutBinDec(val interface{},mode,pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Add(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	ecode = bindec.SetFromDecimal(dec)	
	if ecode != 0 {return bindec, -3}
	return bindec, 0	
}
func (p *BinDecimal)AddOutString(val interface{},mode,pl int)(string, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return string("0"), ecode}
	dec, ecode = roundFunction(p.Decimal().Add(dec), mode, pl)
	if ecode != 0 {return string("0"), -2}
	return dec.String(), 0
}
func (p *BinDecimal)AddOutDecimal(val interface{},mode,pl int)(Dc.Decimal, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return dec, ecode}
	dec, ecode = roundFunction(p.Decimal().Add(dec), mode, pl)
	if ecode != 0 {return dec, -2}
	return dec, 0
}
///// Sub Function /////////////////////
func (p *BinDecimal)Sub(val interface{}, mode, pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Sub(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	bindec.SetFromDecimal(dec)
	return bindec, 0
}
func (p *BinDecimal)SubOutBinDec(val interface{},mode,pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Sub(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	ecode = bindec.SetFromDecimal(dec)	
	if ecode != 0 {return bindec, -3}
	return bindec, 0	
}
func (p *BinDecimal)SubOutString(val interface{},mode,pl int)(string, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return string("0"), ecode}
	dec, ecode = roundFunction(p.Decimal().Sub(dec), mode, pl)
	if ecode != 0 {return string("0"), -2}
	return dec.String(), 0
}
func (p *BinDecimal)SubOutDecimal(val interface{},mode,pl int)(Dc.Decimal, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return dec, ecode}
	dec, ecode = roundFunction(p.Decimal().Sub(dec), mode, pl)
	if ecode != 0 {return dec, -2}
	return dec, 0
}
///// Mul Function /////////////////////
func (p *BinDecimal)Mul(val interface{},mode,pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Mul(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	ecode = bindec.SetFromDecimal(dec)	
	if ecode != 0 {return bindec, -3}
	return bindec, 0	
}
func (p *BinDecimal)MulOutBinDec(val interface{},mode,pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Mul(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	ecode = bindec.SetFromDecimal(dec)	
	if ecode != 0 {return bindec, -3}
	return bindec, 0	
}
func (p *BinDecimal)MulOutString(val interface{},mode,pl int)(string, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return string("0"), ecode}
	dec, ecode = roundFunction(p.Decimal().Mul(dec), mode, pl)
	if ecode != 0 {return string("0"), -2}
	return dec.String(), 0
}
func (p *BinDecimal)MulOutDecimal(val interface{},mode,pl int)(Dc.Decimal, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return dec, ecode}
	dec, ecode = roundFunction(p.Decimal().Mul(dec), mode, pl)
	if ecode != 0 {return dec, -2}
	return dec, 0
}
///// Div Function /////////////////////
func (p *BinDecimal)Div(val interface{},mode,pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Div(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	ecode = bindec.SetFromDecimal(dec)	
	if ecode != 0 {return bindec, -3}
	return bindec, 0	
}
func (p *BinDecimal)DivOutBinDec(val interface{},mode,pl int)(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec, ecode = roundFunction(p.Decimal().Div(dec), mode, pl)
	if ecode != 0 {return bindec, -2}
	ecode = bindec.SetFromDecimal(dec)	
	if ecode != 0 {return bindec, -3}
	return bindec, 0	
}
func (p *BinDecimal)DivOutString(val interface{},mode,pl int)(string, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return string("0"), ecode}
	dec, ecode = roundFunction(p.Decimal().Div(dec), mode, pl)
	if ecode != 0 {return string("0"), -2}
	return dec.String(), 0
}
func (p *BinDecimal)DivOutDecimal(val interface{},mode,pl int)(Dc.Decimal, int){
	dec, ecode := getDecimal(val)
	if ecode != 0 {return dec, ecode}
	dec, ecode = roundFunction(p.Decimal().Div(dec), mode, pl)
	if ecode != 0 {return dec, -2}
	return dec, 0
}
///// MOD Function /////////////////////
func (p *BinDecimal)Mod(val interface{})(BinDecimal, int){
	bindec := New()
	dec, ecode := getDecimal(val)
	if ecode != 0 {return bindec, ecode}
	dec = p.Decimal().Mod(dec)
	ecode = bindec.SetFromDecimal(dec)	
	if ecode != 0 {return bindec, -3}
	return bindec, 0	
}
