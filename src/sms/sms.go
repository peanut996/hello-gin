package sms

import (
	"errors"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"log"
	"os"
	"taurus-backend/constant"
)

var client *Client

func GetSMSClient() *Client {
	return client
}

type Client struct {
	session    *sms.Client
	credential *common.Credential
	profile    *profile.ClientProfile
}

type Task struct {
	Phone     string
	AwardType int
	AwardCode string
}

func CheckSmsEnv() {
	if os.Getenv("SECRET_ID") == "" ||
		os.Getenv("SECRET_KEY") == "" ||
		os.Getenv("SMS_SDK_APP_ID") == "" ||
		os.Getenv("SMS_SIGN_NAME") == "" ||
		os.Getenv("TEMPLATE_ID") == "" {
		log.Fatal("SMS env not correct.")
	}
}

func Init(e *constant.Env) *Client {
	c := &Client{}
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	cpf.NetworkFailureMaxRetries = 3
	cpf.NetworkFailureRetryDuration = profile.ConstantDurationFunc(5)

	credential := common.NewCredential(e.SecretId, e.SecretKey)
	smsClient, err := sms.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		log.Fatalln("InitSMSClient error:", err)
		return nil
	}
	c.profile = cpf
	c.session = smsClient
	c.credential = credential
	client = c
	return c
}

func (c *Client) SendSMS(phone string, awardType int, awardCode string) (smsSerialNo string, err error) {
	req := sms.NewSendSmsRequest()
	req.PhoneNumberSet = common.StringPtrs([]string{phone})
	req.SmsSdkAppId = common.StringPtr(os.Getenv("SMS_SDK_APP_ID"))
	req.SignName = common.StringPtr(os.Getenv("SMS_SIGN_NAME"))
	req.TemplateId = common.StringPtr(os.Getenv("TEMPLATE_ID"))
	req.TemplateParamSet = getTemplateParamSet(awardType, awardCode)

	resp, err := c.session.SendSms(req)
	if _, ok := err.(*terrors.TencentCloudSDKError); ok {
		log.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		return "", err
	}
	sendStatus := resp.Response.SendStatusSet[0]
	if *sendStatus.Code != "Ok" || *resp.Response.SendStatusSet[0].SerialNo == "" {
		return "", errors.New(fmt.Sprintf("send sms fail, code: %v, phone: %v", *sendStatus.Code, phone))
	}
	return *resp.Response.SendStatusSet[0].SerialNo, nil
}

func getTemplateParamSet(awardType int, code string) []*string {
	var paramSet []*string
	switch awardType {
	case constant.MEITUAN:
		paramSet = common.StringPtrs([]string{"?????????????????????", code, "??????APP"})
	case constant.TENCENT:
		paramSet = common.StringPtrs([]string{"??????????????????", code, "????????????App"})
	case constant.DIDI:
		paramSet = common.StringPtrs([]string{"??????????????????", code, "????????????APP/?????????"})
	default:
	}
	return paramSet
}
