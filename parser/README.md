# parser

## 例子
```go
    // 1. 创建FBX解析器实例
    fbxParser := NewFBXParser()
    
    // 2. 解析FBX文件中的贴图引用
    fbxPath := "./uploads/model.fbx"
    textureRefs, err := fbxParser.ParseTextureReferences(fbxPath)
    if err != nil {
        log.Printf("Failed to parse FBX texture references: %v", err)
        return
    }
    
    fmt.Printf("Found %d texture references:\n", len(textureRefs))
    for i, texture := range textureRefs {
        fmt.Printf("  %d. %s\n", i+1, texture)
    }
    
    // 3. 获取FBX文件的详细信息
    fbxInfo, err := fbxParser.GetFBXInfo(fbxPath)
    if err != nil {
        log.Printf("Failed to get FBX info: %v", err)
        return
    }
    
    fmt.Printf("\nFBX文件信息：\n")
    fmt.Printf("  版本: %s\n", fbxInfo.Version)
    fmt.Printf("  创建者: %s\n", fbxInfo.Creator)
    fmt.Printf("  格式: %s\n", func() string {
        if fbxInfo.IsBinary {
            return "二进制"
        }
        return "ASCII"
    }())
    fmt.Printf("  贴图引用数量: %d\n", len(fbxInfo.TextureRefs))
    
    // 4. 创建贴图映射示例
    textureMapping := []TextureMapping{
        {Source: "znzmoModel-1112543272-0077.jpg", Target: "/uploads/ce7d52265d2140c880cf45b7f041962d.jpg"},
        {Source: "diffuse.jpg", Target: "/uploads/abc123def456.jpg"},
        {Source: "normal.png", Target: "/uploads/xyz789uvw456.png"},
    }
    
    fmt.Printf("\n贴图映射示例：\n")
    for i, mapping := range textureMapping {
        fmt.Printf("  %d. 原文件: %s -> 目标URL: %s\n", i+1, mapping.Source, mapping.Target)
    }
    
    // 5. 获取支持的文件格式
    modelExts := fbxParser.GetSupportedModelExtensions()
    textureExts := fbxParser.GetSupportedTextureExtensions()
    
    fmt.Printf("\n支持的模型格式：\n")
    for ext := range modelExts {
        fmt.Printf("  %s\n", ext)
    }
    
    fmt.Printf("\n支持的贴图格式：\n")
    for ext := range textureExts {
        fmt.Printf("  %s\n", ext)
    }

```