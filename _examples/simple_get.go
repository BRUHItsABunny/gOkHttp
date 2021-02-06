package main

import (
	"fmt"
	gokhttp "github.com/BRUHItsABunny/gOkHttp"
	"net/url"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	req, _ := client.MakeGETRequest("http://httpbin.org/get", url.Values{"param1": {"value"}}, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
