package gopay

import (
	"errors"
	"github.com/guidao/gopay/client"
	"github.com/guidao/gopay/common"
	"github.com/guidao/gopay/constant"
	//"github.com/guidao/gopay/util"
	"strconv"
)

func Pay(charge *common.Charge) (string, error) {
	err := checkCharge(charge)
	if err != nil {
		//log.Error(err, charge)
		return "", err
	}

	ct := getPayClient(charge.PayMethod)
	re, err := ct.Pay(charge)
	if err != nil {
		//log.Error("支付失败:", err, charge)
		return "", err
	}
	return re, err
}

func checkCharge(charge *common.Charge) error {
	var id uint64
	var err error
	if charge.UserID == "" {
		id = 0
	} else {
		id, err = strconv.ParseUint(charge.UserID, 10, -1)
		if err != nil {
			return err
		}
	}
	if id < 0 {
		return errors.New("userID less than 0")
	}
	if charge.PayMethod < 0 {
		return errors.New("payMethod less than 0")
	}
	if charge.MoneyFee < 0 {
		return errors.New("totalFee less than 0")
	}

	if charge.CallbackURL == "" {
		return errors.New("callbackURL is NULL")
	}
	return nil
}

// getPayClient 得到需要支付的客户端
func getPayClient(payMethod int64) common.PayClient {
	//如果使用余额支付
	switch payMethod {
	case constant.ALI_WEB:
		return client.DefaultAliWebClient()
	case constant.ALI_APP:
		return client.DefaultAliAppClient()
	case constant.WECHAT:
		return client.DefaultWechatAppClient()
	}
	return nil
}
