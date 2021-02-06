package main

import (
	"fmt"
	gokhttp "github.com/BRUHItsABunny/gOkHttp"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	req, _ := client.MakeGETRequest("http://httpbin.org/get", nil, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
