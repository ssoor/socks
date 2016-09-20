package compiler_test

import (
	"errors"
	"fmt"
	"github.com/ssoor/socks/compiler"
)

func test() {
	var smatch SMatch
	var rules compiler.SCompiler

	rules.Add("www.hao123.com", "s@^(http[s]?)://www.hao123.com/*/\\?.*$@$1://www.hao123.com/?tn=13087099_4_hao_pg@i")
	rules.Add("www.baidu.com", "s@^(http[s]?)://www.baidu.com/*/s\\?(?:(.*)&)?(?:tn=[^&]*)(.*)$@$1://www.baidu.com/s?$2&tn=13087099_4_hao_pg$3@i")

	rules.Add("www.sogou.com", "s@^(http[s]?)://www.sogou.com/*/sogou\\?(?:(.*)&)?(?:pid=[^&]*)(.*)$@$1://www.sogou.com/sogou?$2&pid=sogou-netb-3be0214185d6177a-4012$3@i")

	dsturl, err := rules.Replace("www.baidu.com", "http://www.baidu.com/s?word=dfgdfg&tn=10018800_hao_pg&ie=utf-8&ssl_sample=normal")
	fmt.Printf("%s - %s\n", err, dsturl)

	dsturl, err = rules.Replace("www.sogou.com", "http://www.sogou.com/sogou?query=dfgdfg&_asf=www.sogou.com&_ast=1452842311&w=&p=40040702&pid=sogou-netb-51be2fed6c55f5aa-7749%00&sut=935&sst0=1452842311151&lkt=6%2C1452842310216%2C1452842310489")
	fmt.Printf("%s - %s\n", err, dsturl)

	dsturl, err = rules.Replace("www.hao123.com", "http://www.hao123.com/?tn=130asd9_4_hao_pg")
	fmt.Printf("%s - %s\n", err, dsturl)

	dsturl, err = rules.Replace("www.hao123.com", "http://www.hao123.com/api/newforecast?callback=jQuery17208796808742918074_1452839147843&t=1&_=1452839148069")
	fmt.Printf("%s - %s\n", err, dsturl)
}
