package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"log"
	"os"
	"strings"
)

type PersonFeature struct {
	featureId int64
	personId  string
	feature   string
}

const (
	ShortLen = 2
	IntLen   = 4
	LongLen  = 8
)

// Decode 解析数据
func Decode(data []byte, id string) ([]PersonFeature, error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("data of specified file is empty")
	}
	i := 0
	//HEADER 和 Version信息
	i += IntLen * 2
	result := make([]PersonFeature, 0)
	for i < len(data) {
		var personFeature PersonFeature
		err := binary.Read(bytes.NewBuffer(data[:i+LongLen]), binary.BigEndian, &personFeature.featureId)
		if err != nil {
			return nil, err
		}
		i += LongLen
		var personIdLen int16
		err = binary.Read(bytes.NewBuffer(data[i:i+ShortLen]), binary.BigEndian, &personIdLen)
		if err != nil {
			return nil, err
		}
		i += ShortLen
		personFeature.personId = string(data[i : i+int(personIdLen)])
		i += int(personIdLen)

		var orientation int32
		err = binary.Read(bytes.NewBuffer(data[i:i+IntLen]), binary.BigEndian, &orientation)
		if err != nil {
			return nil, err
		}
		i += IntLen

		var imageReliability int32
		err = binary.Read(bytes.NewBuffer(data[i:i+IntLen]), binary.BigEndian, &imageReliability)
		if err != nil {
			return nil, err
		}
		i += IntLen

		var featureLen int16
		err = binary.Read(bytes.NewBuffer(data[i:i+2]), binary.BigEndian, &featureLen)
		if err != nil {
			return nil, err
		}
		i += 2
		personFeature.feature = base64.StdEncoding.EncodeToString(data[i : i+int(featureLen)])
		i += int(featureLen)
		if id == "" || id == personFeature.personId {
			result = append(result, personFeature)
		}
	}
	return result, nil
}

//把结果写入到指定的文件中
func writeResult(result []PersonFeature, resultFile string) {
	if result == nil || len(result) == 0 {
		return
	}
	f, err := os.Create(resultFile)
	if err != nil {
		log.Panic("Create result file failed")
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalln("error while close file")
		}
	}(f)

	var resultString strings.Builder

	for _, r := range result {
		resultString.WriteString(r.personId + "\r\n")
		resultString.WriteString(r.feature + "\r\n")
	}
	_, err = f.WriteString(resultString.String())
	if err != nil {
		log.Fatalln("error while write result")
	}
}

var (
	f string
	p string
	r string
)

func main() {
	flag.StringVar(&f, "f", "", "File to Read")
	flag.StringVar(&p, "p", "", "PersonId to Analyze")
	flag.StringVar(&r, "r", "result.txt", "Result to Save")
	flag.Parse()
	if f == "" {
		flag.Usage()
		return
	}
	if data, err := os.ReadFile(f); err != nil {
		log.Fatalln("error read file", err)
	} else {
		result, _ := Decode(data, strings.TrimSpace(p))
		writeResult(result, "result.txt")
	}
}
