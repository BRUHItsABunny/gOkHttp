package main

import (
	"fmt"
	"gokhttp"
	"io"
	"os"
)

func main() {
	client := gokhttp.GetHTTPClient(nil)
	f, _ := os.Open("test.txt")
	mBody, contentType, _ := client.MakePostMultiPart(map[string]string{"field1": "value"}, map[string]io.Reader{"test.txt": f}, []string{"uploaded.txt"})
	req, _ := client.MakeMultiPartPOSTRequest("http://httpbin.org/post", contentType, mBody, map[string]string{})
	resp, _ := client.Do(req)
	body, _ := resp.Bytes()
	fmt.Println(string(body))
}
