package api

import (
	"fmt"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/ssoor/fundadore/common"

	"bytes"
	"net/http"
)

const APIEnkey = "890161F37139989CFA9433BAF32BDAFB"

func Decrypt(key string, base64Code []byte) (decode []byte, err error) {
	for i := 0; i < len(base64Code); i++ {
		base64Code[i] = base64Code[i] - 0x90
	}

	iv := base64Code[:16]
	encode := base64Code[16:]

	var block cipher.Block
	if block, err = aes.NewCipher([]byte(key)); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encode, encode)

	var zipReader io.ReadCloser
	if zipReader, err = zlib.NewReader(bytes.NewBuffer(encode)); nil != err {
		return nil, err
	}
	defer zipReader.Close()

	decodeBuf := bytes.NewBuffer(nil)
	if _, err := io.Copy(decodeBuf, zipReader); nil != err {
		return nil, err
	}

	return decodeBuf.Bytes(), nil
}

func GetURL(url string) (decodeData string, err error) {
	var data []byte

	fmt.Println(url)

	var resp *http.Response
	for i := 0; i < 3; i++ {
		if resp, err = http.Get(url); nil != err {
			continue
		}

		break
	}
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var bodyBuf bytes.Buffer

	bodyBuf.ReadFrom(resp.Body)

	data, err = Decrypt(APIEnkey, bodyBuf.Bytes())
	if err != nil {
		return "", err
	}

	//log.Info("<", url, ">", common.GetValidString(data))

	return common.GetValidString(data), nil
}
