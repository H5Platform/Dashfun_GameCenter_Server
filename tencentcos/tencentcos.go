package tencentcos

import (
	"bytes"
	"context"
	"dashfun_gamecenter/config"
	"dashfun_gamecenter/datasource/data"
	"dashfun_gamecenter/snowflake"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var once sync.Once
var instance *TencentCos

type TencentCos struct {
	client  *cos.Client
	nameGen *snowflake.Worker
}

func Get() *TencentCos {
	once.Do(func() {
		instance = &TencentCos{}
		instance.Init()
	})
	return instance
}

func (t *TencentCos) Init() {
	u, _ := url.Parse(config.GetConfig().TencentCosCfg.BucketUrl)
	b := &cos.BaseURL{BucketURL: u}

	t.client = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.GetConfig().TencentCosCfg.SecretId,
			SecretKey: config.GetConfig().TencentCosCfg.SecretKey,
		},
	})
	t.nameGen = snowflake.Must(snowflake.GetWorker(data.WorkerTencentCosName))
}

func (t *TencentCos) UploadData(key string, data []byte, contentType string) (*cos.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: contentType},
	}
	put, err := t.client.Object.Put(ctx, key, bytes.NewReader(data), opt)
	if err != nil {
		return nil, err
	}

	return put, nil
}

func (t *TencentCos) NextName() string {
	id := t.nameGen.NextId()
	return strconv.FormatInt(id, 36)
}

func (t *TencentCos) UploadFile(key string, filePath string) (*cos.CompleteMultipartUploadResult, *cos.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return t.client.Object.Upload(ctx, key, filePath, nil)
}

func init() {

}
