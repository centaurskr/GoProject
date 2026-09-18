///
/// @brief  : Structure copy TEST
/// @file   : copytest.go
/// @date   : 2024. 12. 27. (금) 17:45:37 KST
/// @author : Cento
///
package main
import (
	Bd "centolib/bindecimal"
	"strconv"
	"fmt"
)
/////////// Memory에 있는 주문 구조 //////////////////////////
type Order struct{           // Sorting key OrdTime
    UserId   int             // User Id
    OrdNum   uint64          // 주문번호 - 9(19)
    BuySell  int             // 매도수구분 (1:매도, 2:매수)
    OrdType  int             // 주문종류(1:LIMIT, 2:MARKET)
    OrdTime  uint64          // 주문시각, UinxNano(), 1728372731578396031(19)
    RegNum   uint64          // 접수번호 - 9(19)
    CheSeq   uint            // 최종체결번호
    Price    Bd.BinDecimal   // 주문가격
    Qty      Bd.BinDecimal   // 주문수량
    RemQty   Bd.BinDecimal   // 주문잔량
    MakeRate Bd.BinDecimal   // Maker 수수료율
    TakeRate Bd.BinDecimal   // Taker 수수료율
    MinTrade Bd.BinDecimal   // 최소주문금액(결제통화기준)
    Margin   Bd.BinDecimal   // 주문증거금
    Fee      Bd.BinDecimal   // 체결 예상 증거금

    OrdPacket [1024]byte // 주문 수신 JSON
}
////////////// 내부통신 주문 Packet ///////////////////////////////
type InnerPacket struct{
    Cancel  int             // 취소구분 (0:정상, 1:취소)
    OrgNum  uint64          // 원주문번호
    order   Order
}

type Inner struct {
	dec      Bd.BinDecimal
	stream   [12]byte
}
type AA struct{
	userid   int 
	inner    Inner
}

func main(){
	dec := Bd.New()
	dec.SetFromString("0.1")

	str := "1234"
	var stream  [12]byte
	copy(stream[:], str)

	aa := AA{
		userid : 100,
		inner:Inner{
			dec    : dec,
			stream : stream,
		},
	}
	fmt.Printf("초기  AA : %+v\n", aa)

	var bb AA
	bb = aa
	fmt.Printf("      bb : %+v\n", bb)

	fmt.Printf("AA.Inner.dec : %s\n",bb.inner.dec.String())

	var u64str = "1234567890123456789"
	u64, err := strconv.Atoi(u64str)
	if err != nil{
		fmt.Println(err)
	} else {
		fmt.Printf("uint64 Test strconv.Atoi(%s) =  %d\n\n", u64str, u64)
	}
	
	inner := Inner{
		dec : dec,
		stream : stream,
	}
	aA := AA{
		userid : 200,
	}
	aA.inner = inner
	fmt.Printf("      aA : %+v\n", aA)


	innerpk := InnerPacket{}

	fmt.Printf("Inner : %+v\n", innerpk)
	fmt.Printf("Inner.order : %+v\n", innerpk.order)
}
