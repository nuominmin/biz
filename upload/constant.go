package upload

const (
	// 默认上传路径
	DefaultUploadDir = "uploads"
)

var (
	// 图片格式
	ImgTypes = []string{
		".jpg",
		".jpeg",
		".png",
		".gif",
		".webp",
		".bmp",
		".svg",
	}

	// 模型格式
	ModelTypes = []string{
		".glb",
		".gltf",
		".obj",
		".fbx",
		".dae",
		".3ds",
		".ply",
		".stl",
	}

	// 视频格式
	VideoTypes = []string{
		".mp4",
		".avi",
		".mov",
		".wmv",
		".flv",
		".webm",
		".mkv",
		".m4v",
	}

	// 压缩文件格式
	ZipTypes = []string{
		".zip",
	}

	// 导入文件格式
	ImportTypes = []string{
		".xlsx",
	}
)
