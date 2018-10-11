/*var inputFileName = "/Users/naver/Desktop/test.mp4"
var outputFileName = "/Users/naver/Desktop/test2.mp4"
var url = "rtmp://kr004e_input:STNB96A@kr004e.relay.main.live.vlive.tv/kr004e_input/kr004e_input"*/
package main

import (
	"fmt"
	"vodpub/pkg/publish"
	"time"
)

func main() {
	publisher :=
		publish.NewBuilder().
		SetDefaultConf().
		SetSource("/Users/naver/Desktop/test.mp4").
		SetUri("ch212.relay.live.nhn.gscdn.com").
		SetApplication("ch212").
		SetStream("ch212").
		SetCredentials("ch212", "CG084BA").
		Build().Initialize()

	fmt.Println(publisher.PublishWithDuration(5 * time.Second))
}
