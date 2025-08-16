# oss

## 例子
```go
func NewOss(data *conf.Data) (upload.OssService, error) {
    return upload.NewOssService(
        upload.WithOssConfig(
            data.Oss.Endpoint,
            data.Oss.AccessKeyId,
            data.Oss.AccessKeySecret,
            data.Oss.BucketName,
            data.Oss.BaseUrl,
        ),
    )
}


// 上传文件
filename := ossSvc.GenerateUniqueFilepath("goods", handler.Filename)
fileURL, err = ossSvc.UploadFile(file, filename)
if err != nil {
    return status.Errorf(codes.Internal, "Failed to upload file to OSS: %v", err)
}

// 下载
reader, ossErr := ossSvc.DownloadFile(safeFilename)
if ossErr == nil {
    defer reader.Close()
    data, err = io.ReadAll(reader)
    if err != nil {
        log.Errorf("read file from OSS error (%+v), filename: %s", err, filename)
        ctx.Response().WriteHeader(500)
        _, _ = ctx.Response().Write([]byte("500 Internal Server Error"))
        return nil
	}
} else {
    log.Warnf("Failed to download from OSS, trying local file: %v", ossErr)
}
```