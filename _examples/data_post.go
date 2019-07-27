package main

import (
	"fmt"
	gokhttp "github.com/BRUHItsABunny/gOkHttp"
	"strings"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	req, _ := client.MakeRawPOSTRequest("http://httpbin.org/post", strings.NewReader("BUNBunbun!"), map[string]string{}, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
