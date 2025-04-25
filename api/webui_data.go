package api

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type webuiData struct {
	body     []byte
	fileInfo os.FileInfo
}

var (
	lock         = sync.RWMutex
	webuiDataMap = make(map[string]webuiData)
)

func initWebui(webuiPath string) {
	for name := range _bindata {
		if err := initWebuiData(webuiPath, name); err != nil {
			log.Panic("init webui data error:" + name)
		}
	}
}

func initWebuiData(webuiPath, name string) error {
	path := filepath.Join(webuiPath, name)
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	data, err := compressText(body)
	if err != nil {
		return err
	}
	lock.Lock()
	webuiDataMap[name] = webuiData{
		body:     data,
		fileInfo: fileInfo,
	}
	lock.Unlock()
	return nil
}

func compressText(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	// 创建一个 gzip 写入器
	writer := gzip.NewWriter(&buf)
	// 写入要压缩的内容
	if _, err := writer.Write(body); err != nil {
		return nil, err
	}
	// 关闭写入器，确保所有数据都被写入
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
