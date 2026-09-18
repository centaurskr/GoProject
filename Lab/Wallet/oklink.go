package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OKLink 응답 구조체 (필요한 필드만 정의)
type OKLinkResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ChainFullName string `json:"chainFullName"`
		Address       string `json:"address"`
		Tag           string `json:"tag"`          // 일반 태그
		EntityName    string `json:"entityName"`   // 거래소 이름 (중요!)
	} `json:"data"`
}

const apiKey = "YOUR_OKLINK_API_KEY" // 발급받은 API 키 입력

func callOkLinkAPI(address string) string {
	// BTC 체인의 주소 상세 정보를 요청하는 URL
	url := fmt.Sprintf("https://www.oklink.com/api/v5/blockchain/address/information?chainShortName=btc&address=%s", address)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	
	// OKLink는 헤더에 API 키를 넣어야 함
	req.Header.Set("Ok-Access-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("API 요청 에러:", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var okResponse OKLinkResponse
	if err := json.Unmarshal(body, &okResponse); err != nil {
		fmt.Println("JSON 파싱 에러:", err)
		return ""
	}

	// 데이터가 있고, EntityName이 비어있지 않은 경우 반환
	if len(okResponse.Data) > 0 {
		data := okResponse.Data[0]
		if data.EntityName != "" {
			return data.EntityName
		}
		if data.Tag != "" {
			return data.Tag
		}
	}

	return "Unknown"
}

func main() {
	// 테스트 주소 (업비트 의심 주소 등)
	testAddr := "35gF8GPfHRCjSdg1g8FKswHaqS1DPtJbYd"
	exchange := callOkLinkAPI(testAddr)
	
	fmt.Printf("🔍 주소: %s\n🏢 판별된 거래소: %s\n", testAddr, exchange)
}
