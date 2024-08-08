package main

import (
	"crypto/tls"
	"fmt"
	"github.com/Anonymouscn/go-tools/restful"
	"net/http"
	"time"
	"unsafe"
)

// X-Auth-Token :  [a824cba6e8d57ddd1d694271259a250e]

func main() {
	//data := &struct {
	//	UserName string `json:"UserName"`
	//	Password string `json:"Password"`
	//}{
	//	UserName: "root",
	//	Password: "Pgl15218185426*",
	//}
	client := restful.NewRestClient().
		ApplyConfig(&restful.RestClientConfig{
			EnableRetry:    true,
			MaxRetry:       3,
			RequestTimeout: 2 * time.Minute,
			RetryDelay:     500 * time.Millisecond,
		}).

		//SetURL("https://bing.com").
		//SetPath(restful.Path{"s"}).

		//SetURL("https://192.168.2.23/redfish/v1/SessionService/Sessions").

		SetURL("https://192.168.2.23/redfish/v1/Chassis/1/Thermal#Fans").
		SetHeaders(restful.Data{"X-Auth-Token": "a824cba6e8d57ddd1d694271259a250e"}).

		//SetURL("http://demo.ip-api.com/json/").
		//SetPath(restful.Path{"219.133.188.223"}).
		//SetQuery(restful.Data{"lang": "en"}).

		//SetHeaders(restful.Data{"Accept": "application/json; charset=utf-8"}).

		//SetQuery(restful.Data{"fields": "66842623"}).

		ApplyTransPort(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, DisableKeepAlives: true}).

		//DisableCertAuth().
		//SetBody(data).
		Get()
	fmt.Println("http client size: ", unsafe.Sizeof(http.Client{}))
	fmt.Println("take size: ", unsafe.Sizeof(*client))
	fmt.Println("retry: ", client.TimesOfRetry())
	headers := client.ResponseHeaders()
	if headers != nil {
		fmt.Println("headers:")
		for k, v := range *headers {
			fmt.Println(k, ": ", v)
		}
	}
	resp, err := client.Stringify()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("response: ", resp)
}
