///
/// @brief   Test BinDecimal package
/// @file    testbindecimal.go
/// @date    2024. 10. 23. (수) 15:39:22 KST
/// @author  Cento
///

package main

import (
	"fmt"
	Bd "centolib/bindecimal"
	Dc "github.com/shopspring/decimal"
	Log "centolib/logout"
)
func main() {
	var bdst, bdin, bdout Bd.BinDecimal
	var dcin, dcout Dc.Decimal
	var stin, stout string
	var ecode int

	Log.InitLogger("")
	defer Log.Close()

	bdst.SetFromString("12345.123456789")
	stin = "12345.123456789"
	dcin, _ = Dc.NewFromString("12345.123456789")
	bdin.SetFromString("12345.123456789")

	Log.Infor("========== ADD TEST ==============")
	fmt.Printf("AddOutBinDec() TEST ======\n")
	bdout , ecode = bdst.AddOutBinDec (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, bdout.String(), ecode)
	bdout , ecode = bdst.AddOutBinDec (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), bdout.String(), ecode)
	bdout , ecode = bdst.AddOutBinDec (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), bdout.String(), ecode)

	fmt.Printf("\nAddOutString() TEST ======\n")
	stout , ecode = bdst.AddOutString (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, stout, ecode)
	stout , ecode = bdst.AddOutString (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), stout, ecode)
	stout , ecode = bdst.AddOutString (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), stout, ecode)

	fmt.Printf("\nAddOutDecimal() TEST ======\n")
	dcout , ecode = bdst.AddOutDecimal (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, dcout.String(), ecode)
	dcout , ecode = bdst.AddOutDecimal (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), dcout.String(), ecode)
	dcout , ecode = bdst.AddOutDecimal (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n\n", bdst.String(), bdin.String(), dcout.String(), ecode)

	Log.Infor("========== SUB TEST ==============")
	stin = "1.000000789"
	dcin, _ = Dc.NewFromString("1.000000789")
	bdin.SetFromString("1.000000789")

	fmt.Printf("SubOutBinDec() TEST ======\n")
	bdout , ecode = bdst.SubOutBinDec (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, bdout.String(), ecode)
	bdout , ecode = bdst.SubOutBinDec (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), bdout.String(), ecode)
	bdout , ecode = bdst.SubOutBinDec (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), bdout.String(), ecode)

	fmt.Printf("\nSubOutString() TEST ======\n")
	stout , ecode = bdst.SubOutString (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, stout, ecode)
	stout , ecode = bdst.SubOutString (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), stout, ecode)
	stout , ecode = bdst.SubOutString (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), stout, ecode)

	fmt.Printf("\nSubOutDecimal() TEST ======\n")
	dcout , ecode = bdst.SubOutDecimal (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, dcout.String(), ecode)
	dcout , ecode = bdst.SubOutDecimal (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), dcout.String(), ecode)
	dcout , ecode = bdst.SubOutDecimal (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n\n", bdst.String(), bdin.String(), dcout.String(), ecode)

	Log.Infor("========== MUL TEST ==============12345.133197091")
	stin = "1.000000789"
	dcin, _ = Dc.NewFromString("1.000000789")
	bdin.SetFromString("1.000000789")

	fmt.Printf("MulOutBinDec() TEST ======\n")
	bdout , ecode = bdst.MulOutBinDec (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, bdout.String(), ecode)
	bdout , ecode = bdst.MulOutBinDec (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), bdout.String(), ecode)
	bdout , ecode = bdst.MulOutBinDec (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), bdout.String(), ecode)

	fmt.Printf("\nMulOutString() TEST ======\n")
	stout , ecode = bdst.MulOutString (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, stout, ecode)
	stout , ecode = bdst.MulOutString (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), stout, ecode)
	stout , ecode = bdst.MulOutString (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), stout, ecode)

	fmt.Printf("\nMulOutDecimal() TEST ======\n")
	dcout , ecode = bdst.MulOutDecimal (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, dcout.String(), ecode)
	dcout , ecode = bdst.MulOutDecimal (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), dcout.String(), ecode)
	dcout , ecode = bdst.MulOutDecimal (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n\n", bdst.String(), bdin.String(), dcout.String(), ecode)

	Log.Infor("========== DIV TEST ==============12345.113716494")
/*
	stin = "1.000000789"
	dcin, _ = Dc.NewFromString("1.000000789")
	bdin.SetFromString("1.000000789")
*/
	bdst.SetFromString("1.0")
	stin = "3.0"
	dcin, _ = Dc.NewFromString("3.0")
	bdin.SetFromString("3.0")

	fmt.Printf("DivOutBinDec() TEST ======\n")
	bdout , ecode = bdst.DivOutBinDec (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, bdout.String(), ecode)
	bdout , ecode = bdst.DivOutBinDec (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), bdout.String(), ecode)
	bdout , ecode = bdst.DivOutBinDec (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), bdout.String(), ecode)

	fmt.Printf("\nDivOutString() TEST ======\n")
	stout , ecode = bdst.DivOutString (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, stout, ecode)
	stout , ecode = bdst.DivOutString (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), stout, ecode)
	stout , ecode = bdst.DivOutString (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n", bdst.String(), bdin.String(), stout, ecode)

	fmt.Printf("\nDivOutDecimal() TEST ======\n")
	dcout , ecode = bdst.DivOutDecimal (stin, Bd.ROUND,    8)
	fmt.Printf("R  (8) : %s + String    (%s) = %s, error(%d)\n", bdst.String(), stin, dcout.String(), ecode)
	dcout , ecode = bdst.DivOutDecimal (dcin, Bd.ROUND_UP, 8)
	fmt.Printf("RUP(8) : %s + Decimal   (%s) = %s, error(%d)\n", bdst.String(), dcin.String(), dcout.String(), ecode)
	dcout , ecode = bdst.DivOutDecimal (bdin, Bd.ROUND_DN, 8)
	fmt.Printf("RDN(8) : %s + BinDecimal(%s) = %s, error(%d)\n\n", bdst.String(), bdin.String(), dcout.String(), ecode)

	Log.Infor("========== Equal TEST ==============12345.113716494")
	bdst.SetFromString("12345.123456789")
	stin = "1.000000789"
	dcin, _ = Dc.NewFromString("1.000000789")
	bdin.SetFromString("1.000000789")

	fmt.Printf("String(%s)     == BinDecimal(%s) = %d\n", stin, bdin.String(), Bd.DecimalEqual(stin, bdin))
	fmt.Printf("String(%s)     ==    Decimal(%s) = %d\n", stin, dcin.String(), Bd.DecimalEqual(stin, dcin))
	fmt.Printf("String(%s)     ==     String(%s) = %d\n\n", stin, bdin.String(), Bd.DecimalEqual(stin, bdin.String()))

	fmt.Printf("BinDecimal(%s) == BinDecimal(%s) = %d\n", bdin.String(), bdin.String(), Bd.DecimalEqual(bdin, bdin))
	fmt.Printf("BinDecimal(%s) ==    Decimal(%s) = %d\n", bdin.String(), dcin.String(), Bd.DecimalEqual(bdin, dcin))
	fmt.Printf("BinDecimal(%s) ==     String(%s) = %d\n\n", bdin.String(), stin         , Bd.DecimalEqual(bdin, stin))

	fmt.Printf("Decimal(%s)    == BinDecimal(%s) = %d\n", dcin.String(), bdin.String(), Bd.DecimalEqual(dcin, bdin))
	fmt.Printf("Decimal(%s)    ==    Decimal(%s) = %d\n", dcin.String(), dcin.String(), Bd.DecimalEqual(dcin, dcin))
	fmt.Printf("Decimal(%s)    ==     String(%s) = %d\n\n", dcin.String(), stin         , Bd.DecimalEqual(dcin, stin))

	stin  = "123.123123"
	fmt.Printf("String(%s)     != BinDecimal(%s) = %d\n", stin, bdin.String(), Bd.DecimalEqual(stin, bdin))
	fmt.Printf("String(%s)     !=    Decimal(%s) = %d\n", stin, dcin.String(), Bd.DecimalEqual(stin, dcin))
	fmt.Printf("String(%s)     !=     String(%s) = %d\n\n", stin, bdin.String(), Bd.DecimalEqual(stin, bdin.String()))

	stin = "1.000000789"
	bdin.SetFromString("123.123123")
	fmt.Printf("BinDecimal(%s) != BinDecimal(%s) = %d\n", bdin.String(), bdout.String(), Bd.DecimalEqual(bdin, bdout))
	fmt.Printf("BinDecimal(%s) !=    Decimal(%s) = %d\n", bdin.String(), dcin.String(), Bd.DecimalEqual(bdin, dcin))
	fmt.Printf("BinDecimal(%s) !=     String(%s) = %d\n\n", bdin.String(), stin         , Bd.DecimalEqual(bdin, stin))

	bdin.SetFromString("1.000000789")
	dcin, _ = Dc.NewFromString("123.123123")
	fmt.Printf("Decimal(%s)    != BinDecimal(%s) = %d\n", dcin.String(), bdin.String(), Bd.DecimalEqual(dcin, bdin))
	fmt.Printf("Decimal(%s)    !=    Decimal(%s) = %d\n", dcin.String(), dcout.String(), Bd.DecimalEqual(dcin, dcout))
	fmt.Printf("Decimal(%s)    !=     String(%s) = %d\n\n", dcin.String(), stin         , Bd.DecimalEqual(dcin, stin))

	Log.Infor("========== Compare TEST ==============")
	bdst.SetFromString("12345.123456789")
	stin    = "1.000000789"
	dcin, _ = Dc.NewFromString("1.000000789")
	bdin.SetFromString("1.000000789")

	var bdin1 Bd.BinDecimal
	var stin1 string
	var dcin1 Dc.Decimal

//// String
	bdin1.SetFromString("1231231230.1231123")	
	fmt.Printf("String(%s) > BinDecimal(%s) = %d\n",   stin, bdout.String(), Bd.DecimalCompare(stin, bdout))
	fmt.Printf("String(%s) = BinDecimal(%s) = %d\n",   stin, bdin.String(), Bd.DecimalCompare(stin, bdin))
	fmt.Printf("String(%s) < BinDecimal(%s) = %d\n", stin, bdin1.String(), Bd.DecimalCompare(stin, bdin1))

	dcin1,_ = Dc.NewFromString("1231231230.1231123")	
	fmt.Printf("String(%s) > Decimal(%s) = %d\n",   stin, dcout.String(), Bd.DecimalCompare(stin, dcout))
	fmt.Printf("String(%s) = Decimal(%s) = %d\n",   stin, dcin.String(), Bd.DecimalCompare(stin, dcin))
	fmt.Printf("String(%s) < Decimal(%s) = %d\n", stin, dcin1.String(), Bd.DecimalCompare(stin, dcin1))

	stin1 = "1231231230.1231123"
	fmt.Printf("String(%s) > String(%s) = %d\n",   stin, stout, Bd.DecimalCompare(stin, stout))
	fmt.Printf("String(%s) = String(%s) = %d\n",   stin, dcin.String(), Bd.DecimalCompare(stin, dcin.String()))
	fmt.Printf("String(%s) < String(%s) = %d\n\n", stin, stin1, Bd.DecimalCompare(stin, stin1))
//// BinDecimal
	bdin1.SetFromString("1231231230.1231123")	
	fmt.Printf("BinDecimal(%s) > BinDecimal(%s) = %d\n",   bdin.String(), bdout.String(), Bd.DecimalCompare(bdin, bdout))
	fmt.Printf("BinDecimal(%s) = BinDecimal(%s) = %d\n",   bdin.String(), bdin.String(), Bd.DecimalCompare(bdin, bdin))
	fmt.Printf("BinDecimal(%s) < BinDecimal(%s) = %d\n",   bdin.String(), bdin1.String(), Bd.DecimalCompare(bdin, bdin1))

	dcin1,_ = Dc.NewFromString("1231231230.1231123")	
	fmt.Printf("BinDecimal(%s) > Decimal(%s) = %d\n",   bdin.String(), dcout.String(), Bd.DecimalCompare(bdin, dcout))
	fmt.Printf("BinDecimal(%s) = Decimal(%s) = %d\n",   bdin.String(), dcin.String(), Bd.DecimalCompare(bdin, dcin))
	fmt.Printf("BinDecimal(%s) < Decimal(%s) = %d\n",   bdin.String(), dcin1.String(), Bd.DecimalCompare(bdin, dcin1))

	stin1 = "1231231230.1231123"
	fmt.Printf("BinDecimal(%s) > String(%s) = %d\n",   bdin.String(), stout, Bd.DecimalCompare(bdin, stout))
	fmt.Printf("BinDecimal(%s) = String(%s) = %d\n",   bdin.String(), dcin.String(), Bd.DecimalCompare(bdin, dcin.String()))
	fmt.Printf("BinDecimal(%s) < String(%s) = %d\n\n",   bdin.String(), stin1, Bd.DecimalCompare(bdin, stin1))
///// Decimal
	bdin1.SetFromString("1231231230.1231123")	
	fmt.Printf("Decimal(%s) > BinDecimal(%s) = %d\n",   dcin.String(), bdout.String(), Bd.DecimalCompare(dcin, bdout))
	fmt.Printf("Decimal(%s) = BinDecimal(%s) = %d\n",   dcin.String(), bdin.String(), Bd.DecimalCompare(dcin, bdin))
	fmt.Printf("Decimal(%s) < BinDecimal(%s) = %d\n",   dcin.String(), bdin1.String(), Bd.DecimalCompare(dcin, bdin1))

	dcin1,_ = Dc.NewFromString("1231231230.1231123")	
	fmt.Printf("Decimal(%s) > Decimal(%s) = %d\n",   dcin.String(), dcout.String(), Bd.DecimalCompare(dcin, dcout))
	fmt.Printf("Decimal(%s) = Decimal(%s) = %d\n",   dcin.String(), dcin.String(), Bd.DecimalCompare(dcin, dcin))
	fmt.Printf("Decimal(%s) < Decimal(%s) = %d\n",   dcin.String(), dcin1.String(), Bd.DecimalCompare(dcin, dcin1))

	stin1 = "1231231230.1231123"
	fmt.Printf("Decimal(%s) > String(%s) = %d\n",   dcin.String(), stout, Bd.DecimalCompare(dcin, stout))
	fmt.Printf("Decimal(%s) = String(%s) = %d\n",   dcin.String(), dcin.String(), Bd.DecimalCompare(dcin, dcin.String()))
	fmt.Printf("Decimal(%s) < String(%s) = %d\n\n",   dcin.String(), stin1, Bd.DecimalCompare(dcin, stin1))
}
