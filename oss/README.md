# oss

## 例子
```go
func NewOss() oss.Service {
	return oss.NewOss()
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