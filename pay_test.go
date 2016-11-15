package gopay

import (
	"fmt"
	"github.com/guidao/gopay/client"
	"github.com/guidao/gopay/common"
	"github.com/guidao/gopay/constant"
	"net/http"
	"testing"
)

func TestPay(t *testing.T) {
	initClient()
	initHandle()
	charge := new(common.Charge)
	charge.PayMethod = constant.WECHAT
	charge.MoneyFee = 1
	charge.Describe = "test pay"
	charge.TradeNum = "1111111111"

	fdata, err := Pay(charge)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(fdata)
}

func initClient() {
	client.InitAliAppClient(&client.AliAppClient{
		PartnerID:  "xxx",
		SellerID:   "xxxx",
		AppID:      "xxx",
		PrivateKey: nil,
		PublicKey:  nil,
	})
}

func initHandle() {
	http.HandleFunc("callback/aliappcallback", func(w http.ResponseWriter, r *http.Request) {
		aliResult, err := AliAppCallback(w, r)
		if err != nil {
			fmt.Println(err)
			//log.xxx
			return
		}
		selfHandler(aliResult)
	})
}

func selfHandler(i interface{}) {
}
