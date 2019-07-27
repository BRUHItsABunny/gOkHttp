package main

import (
	"fmt"
	gokhttp "github.com/BRUHItsABunny/gOkHttp"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	mBody, contentType, _ := client.MakePostMultiPart(map[string]string{"field1": "value"}, nil, nil)
	req, _ := client.MakeMultiPartPOSTRequest("http://httpbin.org/post", contentType, mBody, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
