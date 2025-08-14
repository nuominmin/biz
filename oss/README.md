# oss

## 例子
```go
func NewOss(data *conf.Data) (oss.Service, error) {
    return oss.NewOss(
        oss.WithEndpoint(data.Oss.Endpoint),
        oss.WithAccessKeyId(data.Oss.AccessKeyId),
        oss.WithAccessKeySecret(data.Oss.AccessKeySecret),
        oss.WithBucketName(data.Oss.BucketName),
        oss.WithBaseUrl(data.Oss.BaseUrl),
    )
}


// 上传文件
filename := ossSvc.GenerateUniqueFilepath("goods", handler.Filename)
fileURL, err = ossSvc.UploadFile(file, filename)
if err != nil {
    return status.Errorf(codes.Internal, "Failed to upload file to OSS: %v", err)
}

// 下载
reader, _ := ossSvc.DownloadFile(safeFilename)

```