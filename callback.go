package gopay

import (
	"gopay/common"
	"io/ioutil"
	"net/http"
	"sort"
)

func AliWebCallback(w http.ResponseWriter, r *http.Request) {
	var m = make(map[string]string)
	var signSlice []string
	r.ParseForm()
	for k, v := range r.Form {
		// k不会有多个值的情况
		m[k] = v[0]
		if k == "sign" || k == "sign_type" {
			continue
		}
		signSlice = append(signSlice, fmt.Sprintf("%s=%s", k, v[0]))
	}

	sort.Strings(signSlice)
	signData := strings.Join(signSlice, "&")
	if m["sign_type"] != "RSA" {
		log.Error(m)
		return
	}

	err := client.DefaultAliWebClient().CheckSign(signData, m["sign"])
	if err != nil {
		log.Error("签名验证失败：", err, signData, m)
		return
	}

	var aliPay common.AliWebPayResult
	err = util.MapStringToStruct(m, &aliPay)
	if err != nil {
		log.Error(err)
		w.Write([]byte("error"))
		return
	}

	err = biz.AliWebCallBack(&aliPay)
	if err != nil {
		log.Error(err)
		w.Write([]byte("error"))
		return
	}

	w.Write([]byte("success"))
}

// 支付宝app支付回调
func AliAppCallback(w http.ResponseWriter, r *http.Request) {
	var m = make(map[string]string)
	var signSlice []string
	r.ParseForm()
	for k, v := range c.Form {
		m[k] = v[0]
		if k == "sign" || k == "sign_type" {
			continue
		}
		signSlice = append(signSlice, fmt.Sprintf("%s=%s", k, v[0]))
	}
	sort.Strings(signSlice)
	signData := strings.Join(signSlice, "&")
	if m["sign_type"] != "RSA" {
		log.Error(m)
		return
	}

	err := client.DefaultAliAppClient().CheckSign(signData, m["sign"])
	if err != nil {
		log.Error(err, m, signData)
		w.Write([]byte("error"))
		return
	}

	var aliPay common.AliWebPayResult
	err = util.MapStringToStruct(m, &aliPay)
	if err != nil {
		log.Error(err)
		w.Write([]byte("error"))
		return
	}

	err = biz.AliAppCallBack(&aliPay)
	if err != nil {
		log.Error(err)
		w.Write([]byte("error"))
	}

	w.Write([]byte("success"))
}

// // WeChatCallback 微信公众号支付
// func WeChatCallback(w http.ResponseWriter, r *http.Request) {
// 	var returnCode = "FAIL"
// 	var returnMsg = ""
// 	defer func() {
// 		formatStr := `<xml><return_code><![CDATA[%s]]></return_code>
//                   <return_msg>![CDATA[%s]]</return_msg></xml>`
// 		returnBody := fmt.Sprintf(formatStr, returnCode, returnMsg)
// 		w.Write([]byte(returnBody))
// 	}()

// 	var reXML common.WeChatPayResult
// 	body := cb.Ctx.Input.RequestBody
// 	err := xml.Unmarshal(body, &reXML)
// 	if err != nil {
// 		log.Error(err, string(body))
// 		returnMsg = "参数错误"
// 		returnCode = "FAIL"
// 		return
// 	}

// 	if reXML.ReturnCode != constant.WECHAT_PAY_SUCCEED {
// 		log.Error(reXML)
// 		returnCode = "FAIL"
// 		return
// 	}
// 	m, err := util.XmlToMap(body)
// 	if err != nil {
// 		log.Error(err, body)
// 		returnMsg = "参数错误"
// 		returnCode = "FAIL"
// 		return
// 	}
// 	log.Info(m)
// 	var signData []string
// 	for k, v := range m {
// 		if k == "sign" {
// 			continue
// 		}
// 		signData = append(signData, fmt.Sprintf("%v=%v", k, v))
// 	}
// 	sort.Strings(signData)
// 	signData2 := strings.Join(signData, "&")
// 	err = client.DefaultWechatWebClient().CheckSign(signData2, m["sign"])
// 	if err != nil {
// 		returnCode = "FAIL"
// 		return
// 	}

// 	err = biz.WechatWebCallback(&reXML)
// 	if err != nil {
// 		returnCode = "FAIL"
// 	}
// 	returnCode = "SUCCESS"
// }

// WeChatCallback 微信app支付
func WeChatAppCallback(w http.ResponseWriter, r *http.Request) {
	var returnCode = "FAIL"
	var returnMsg = ""
	defer func() {
		formatStr := `<xml><return_code><![CDATA[%s]]></return_code>
                  <return_msg>![CDATA[%s]]</return_msg></xml>`
		returnBody := fmt.Sprintf(formatStr, returnCode, returnMsg)
		w.Write([]byte(returnBody))
	}()
	var reXML common.WeChatPayResult
	//body := cb.Ctx.Input.RequestBody
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(string(body))
		returnCode = "FAIL"
		returnMsg = "Bodyerror"
		return
	}
	err = xml.Unmarshal(body, &reXML)
	if err != nil {
		log.Error(err, string(body))
		returnMsg = "参数错误"
		returnCode = "FAIL"
		return
	}

	if reXML.ReturnCode != "SUCCESS" {
		log.Error(reXML)
		returnCode = "FAIL"
		return
	}
	m, err := util.XmlToMap(body)
	if err != nil {
		log.Error(err, body)
		returnMsg = "参数错误"
		returnCode = "FAIL"
		return
	}
	log.Info(m)
	var signData []string
	for k, v := range m {
		if k == "sign" {
			continue
		}
		signData = append(signData, fmt.Sprintf("%v=%v", k, v))
	}
	sort.Strings(signData)
	signData2 := strings.Join(signData, "&")
	err = client.DefaultWechatAppClient().CheckSign(signData2, m["sign"])
	if err != nil {
		returnCode = "FAIL"
		return
	}

	err = biz.WechatAppCallback(&reXML)
	if err != nil {
		returnCode = "FAIL"
	}
	returnCode = "SUCCESS"
}
