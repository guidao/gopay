package client

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var defaultWechatWebClient *WechatWebClient

func init() {
	config.Register("client.wechatweb", func() error {
		defaultWechatWebClient = &WechatWebClient{
			appID:       appID,
			mchID:       mchID,
			callbackURL: callbackURL,
			key:         key,
			payURL:      payURL,
			queryURL:    queryURL,
		}
		return nil
	})
}

func DefaultWechatWebClient() *WechatWebClient {
	return defaultWechatWebClient
}

// WechatWebClient 微信公众号支付
type WechatWebClient struct {
	appID       string // 公众账号ID
	mchID       string // 商户号ID
	callbackURL string // 回调地址
	key         string // 密钥
	payURL      string // 支付地址
	queryURL    string // 查询地址
}

// Pay 支付
func (wechat *WechatWebClient) Pay(charge *common.Charge) (string, error) {
	var m = make(map[string]string)
	m["appid"] = wechat.appID
	m["mch_id"] = wechat.mchID
	m["nonce_str"] = util.RandomStr()
	m["body"] = charge.Describe
	m["out_trade_no"] = charge.TradeNum
	m["total_fee"] = fmt.Sprintf("%d", charge.MoneyFee)
	m["spbill_create_ip"] = util.LocalIP()
	m["notify_url"] = wechat.callbackURL
	m["trade_type"] = "JSAPI"
	m["openid"] = charge.OpenID

	sign, err := wechat.GenSign(m)
	if err != nil {
		log.Error("签名失败", err)
		return "", err
	}
	m["sign"] = sign
	// 转出xml结构
	buf := bytes.NewBufferString("")
	for k, v := range m {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	xmlStr := fmt.Sprintf("<xml>%s</xml>", buf.String())
	log.Info("xmlstr:", xmlStr)

	re, err := HTTPSC.PostData(wechat.payURL, "text/xml:charset=UTF-8", xmlStr)
	if err != nil {
		log.Error(err, string(re))
		return "", err
	}
	var xmlRe common.WeChatReResult
	err = xml.Unmarshal(re, &xmlRe)
	if err != nil {
		log.Error(err, string(re))
		return "", err
	}
	log.Info(string(re))

	if xmlRe.ReturnCode != "SUCCESS" {
		// 通信失败
		log.Error(xmlRe.ReturnMsg)
		return "", errors.New(xmlRe.ReturnMsg)
	}

	if xmlRe.ResultCode != "SUCCESS" {
		// 支付失败
		log.Error(xmlRe.ErrCode, xmlRe.ErrCodeDes)
		return "", errors.New(xmlRe.ErrCodeDes)
	}

	var c = make(map[string]string)
	c["appId"] = wechat.appID
	c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	c["nonceStr"] = util.RandomStr()
	c["package"] = fmt.Sprintf("prepay_id=%s", xmlRe.PrepayID)
	c["signType"] = "MD5"

	sign2, err := wechat.GenSign(c)
	if err != nil {
		log.Error(err, c)
		return "", err
	}
	c["paySign"] = sign2

	jsonC, err := json.Marshal(c)
	if err != nil {
		log.Error(err, c)
		return "", err
	}

	return string(jsonC), nil
}

// GenSign 产生签名
func (wechat *WechatWebClient) GenSign(m map[string]string) (string, error) {
	delete(m, "sign")
	delete(m, "key")
	var signData []string
	for k, v := range m {
		if v != "" {
			signData = append(signData, fmt.Sprintf("%s=%s", k, v))
		}
	}

	sort.Strings(signData)
	signStr := strings.Join(signData, "&")
	signStr = signStr + "&key=" + wechat.key
	log.Info(signStr)
	c := md5.New()
	_, err := c.Write([]byte(signStr))
	if err != nil {
		log.Error(err, signStr)
		return "", err
	}
	signByte := c.Sum(nil)
	if err != nil {
		log.Error(err, signStr)
		return "", err
	}
	return strings.ToUpper(fmt.Sprintf("%x", signByte)), nil
}

// CheckSign 检查签名
func (wechat *WechatWebClient) CheckSign(data string, sign string) error {
	signData := data + "&key=" + wechat.key
	c := md5.New()
	_, err := c.Write([]byte(signData))
	if err != nil {
		log.Error(signData, err)
		return err
	}
	signOut := fmt.Sprintf("%x", c.Sum(nil))
	if strings.ToUpper(sign) == strings.ToUpper(signOut) {
		return nil
	}
	log.Error(signOut, sign)
	return errors.New("签名交易错误")
}

// QueryOrder 查询订单
func (wechat *WechatWebClient) QueryOrder(tradeNum string) (*common.WeChatQueryResult, error) {
	var m = make(map[string]string)
	m["appid"] = wechat.appID
	m["mch_id"] = wechat.mchID
	m["out_trade_no"] = tradeNum
	m["nonce_str"] = util.RandomStr()

	sign, err := wechat.GenSign(m)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	m["sign"] = sign

	buf := bytes.NewBufferString("")
	for k, v := range m {
		buf.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	xmlStr := fmt.Sprintf("<xml>%s</xml>", buf.String())

	result, err := HTTPSC.PostData(wechat.queryURL, "text/xml:charset=UTF-8", xmlStr)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var queryResult common.WeChatQueryResult
	err = xml.Unmarshal(result, &queryResult)
	return &queryResult, err
}
