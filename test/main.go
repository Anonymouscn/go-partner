package main

import (
	"fmt"
	"github.com/Anonymouscn/go-tools/restful"
)

func main() {
	data := &struct {
		UserName string `json:"UserName"`
		Password string `json:"Password"`
	}{
		UserName: "root",
		Password: "Pgl15218185426*",
	}
	client := restful.NewRestClient().
		//SetURL("https://bing.com").
		//SetPath(restful.Path{"s"}).
		SetURL("https://192.168.2.23/redfish/v1/SessionService/Sessions").
		//SetPath(restful.Path{"s"}).
		//SetHeaders(restful.Data{"Accept": "application/json; charset=utf-8"}).
		//SetQuery(restful.Data{"wd": "文心一言"}).
		DisableCertAuth().
		SetBody(data).
		Post()
	fmt.Println(client.TimesOfRetry())
	resp, err := client.Stringify()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
}
