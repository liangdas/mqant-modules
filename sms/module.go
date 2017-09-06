/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/
package sms

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"net/http"
	"io/ioutil"
	"fmt"
	"net/url"
	"time"
	"encoding/json"
	"github.com/liangdas/mqant/utils"
	"github.com/garyburd/redigo/redis"
	//"github.com/liangdas/mqant/log"
)

var Module = func() module.Module {
	user := new(SMS)
	return user
}

type SMS struct {
	basemodule.BaseModule
	RedisUrl	string
	TTL		int64
	SendCloud	map[string]interface{}
	Ailyun		map[string]interface{}

}

func (self *SMS) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "SMS"
}
func (self *SMS) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (self *SMS) OnInit(app module.App, settings *conf.ModuleSettings) {
	self.BaseModule.OnInit(self, app, settings)
	self.RedisUrl=self.GetModuleSettings().Settings["RedisUrl"].(string)
	self.TTL=int64(self.GetModuleSettings().Settings["TTL"].(float64))
	if SendCloud,ok:=self.GetModuleSettings().Settings["SendCloud"];ok{
		self.SendCloud=SendCloud.(map[string]interface {})
	}
	if Ailyun,ok:=self.GetModuleSettings().Settings["Ailyun"];ok{
		self.Ailyun=Ailyun.(map[string]interface {})
	}
	self.GetServer().RegisterGO("SendVerifiycode", self.doSendVerifiycode) //演示后台模块间的rpc调用
}

func (self *SMS) Run(closeSig chan bool) {
}

func (self *SMS) OnDestroy() {
	//一定别忘了关闭RPC
	self.GetServer().OnDestroy()
}

func (self *SMS)aliyun(phone string,smsCode int64)(string){
	param:=map[string]string{
		"Action": "SendSms",
		"Version": "2017-05-25",
		"RegionId": "cn-hangzhou",
		"PhoneNumbers": phone,
		"SignName": self.Ailyun["SignName"].(string),
		"TemplateCode": self.Ailyun["TemplateCode"].(string),
		"TemplateParam": fmt.Sprintf("{\"smsCode\": \"%d\"}",smsCode),
		"OutId": fmt.Sprintf("%d",time.Now().Unix()*1000),
	}
	AliyunPOPSignature("POST",self.Ailyun["AccessKeyId"].(string),self.Ailyun["AccessSecret"].(string),param)
	values:=url.Values{}
	for k,v:=range param{
		values[k]=[]string{v}
	}
	//log.Error(values.Encode())
	req, err := http.PostForm("http://dysmsapi.aliyuncs.com", values)
	if err != nil {
		// handle error
		return err.Error()
	}

	//req, err := http.Get("http://dysmsapi.aliyuncs.com?"+values.Encode())
	//if err != nil {
	//	return err.Error()
	//}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}

	/**
	# 部分成功
		{
		    "message":"部分成功",
		    "info":{
			    "successCount":1,
			    "failedCount":1,
			    "items":[{"phone":"1312222","vars":{},"message":"手机号格式错误"}],
			    "smsIds":["1458113381893_15_3_11_1ainnq$131112345678"]}
			    },
		    "result":true,
		    "statusCode":311
		}
	 */
	ret:=map[string]interface{}{}
	err=json.Unmarshal(body,&ret)
	if err!=nil{
		return err.Error()
	}
	if result,ok:=ret["Code"];ok{
		if result.(string)!="OK"{
			if message,ok:=ret["Message"];ok{
				return message.(string)
			}else{
				return "验证码发送失败"
			}
		}else{
			return ""
		}
	}else{
		return string(body)
	}
}


func (self *SMS)sendcloud(phone string,smsCode int64)(string){
	param:=map[string]string{
		"smsUser": self.SendCloud["SmsUser"].(string),
		"templateId": self.SendCloud["TemplateId"].(string),
		"msgType": "0",
		"phone": phone,
		"vars": fmt.Sprintf("{\"smsCode\": \"%d\"}",smsCode),
		"timestamp": fmt.Sprintf("%d",time.Now().Unix()*1000),
	}
	SendCloudSignature(self.SendCloud["SmsKey"].(string),param)
	values:=url.Values{}
	for k,v:=range param{
		values[k]=[]string{v}
	}
	req, err := http.PostForm("http://www.sendcloud.net/smsapi/send",values)
	if err != nil {
		// handle error
		return err.Error()
	}


	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err.Error()
	}

	/**
	# 部分成功
		{
		    "message":"部分成功",
		    "info":{
			    "successCount":1,
			    "failedCount":1,
			    "items":[{"phone":"1312222","vars":{},"message":"手机号格式错误"}],
			    "smsIds":["1458113381893_15_3_11_1ainnq$131112345678"]}
			    },
		    "result":true,
		    "statusCode":311
		}
	 */
	ret:=map[string]interface{}{}
	err=json.Unmarshal(body,&ret)
	if err!=nil{
		return err.Error()
	}
	if result,ok:=ret["result"];ok{
		if !result.(bool){
			if message,ok:=ret["message"];ok{
				return message.(string)
			}else{
				return "验证码发送失败"
			}
		}else{
			return ""
		}
	}else{
		return string(body)
	}
}

func (self *SMS) doSendVerifiycode(phone string,purpose string,extra map[string]interface{}) (string, string){
	conn:=utils.GetRedisFactory().GetPool(self.RedisUrl).Get()
	defer conn.Close()
	ttl, err := redis.Int64(conn.Do("TTL",fmt.Sprintf(MobileTTLFormat,phone)))
	if err != nil {
		return "",err.Error()
	}
	if ttl>0{
		return "","操作过于频繁，请您稍后再试。"
	}

	smsCode:=RandInt64(100000, 999999)
	if self.Ailyun!=nil{
		errstr:=self.aliyun(phone,smsCode)
		if errstr!=""{
			return "",errstr
		}
	}else if self.SendCloud!=nil{
		errstr:=self.sendcloud(phone,smsCode)
		if errstr!=""{
			return "",errstr
		}
	}else{
		return "","没有可用的短信通道。"
	}
	_, err = conn.Do("SET",fmt.Sprintf(MobileTTLFormat,phone),smsCode)
	if err != nil {
		return "",err.Error()
	}
	_, err = conn.Do("EXPIRE",fmt.Sprintf(MobileTTLFormat,phone),self.TTL)
	if err != nil {
		return "",err.Error()
	}

	savedatas:=map[string]interface{}{
		"purpose":purpose,
		"extra":extra,
	}
	savedatasBytes,err:=json.Marshal(savedatas)
	if err!=nil{
		return "",err.Error()
	}
	_, err = conn.Do("SET",fmt.Sprintf(MobileSmsCodeFormat,phone,smsCode),savedatasBytes)
	if err != nil {
		return "",err.Error()
	}
	_, err = conn.Do("EXPIRE",fmt.Sprintf(MobileSmsCodeFormat,phone,smsCode),self.TTL)
	if err != nil {
		return "",err.Error()
	}
	return "验证码发送成功",""
}