package main

import (
	"fmt"
	gokhttp "github.com/BRUHItsABunny/gOkHttp"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	req, _ := client.MakeGETRequest("http://httpbin.org/get", map[string]string{"param1": "value"}, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
