package compiler_test

import (
	"errors"
	"fmt"
	"github.com/ssoor/socks/compiler"
)

func main() {
	var smatch SMatch

	err := smatch.Init("s@^.*@http://1.sogoulp.com/index5883_1.html@i")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(smatch.Replace("https://golang.org/pkg/regexp/#example_MatchString"))
}
