package main

import (
	"fmt"
	"gokhttp"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	req, _ := client.MakeGETRequest("http://httpbin.org/get", map[string]string{}, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
