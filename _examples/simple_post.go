package main

import (
	"fmt"
	"gokhttp"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	req, _ := client.MakePOSTRequest("http://httpbin.org/post", map[string]string{"field1": "value"}, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
