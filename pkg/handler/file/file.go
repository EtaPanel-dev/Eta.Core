package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/EtaPanel-dev/EtaPanel/core/pkg/handler"
	"github.com/EtaPanel-dev/EtaPanel/core/pkg/models"
	"github.com/gin-gonic/gin"
)

// 检查路径是否受保护
func isProtectedPath(path string) bool {
	for _, protected := range models.ProtectedDirs {
		if strings.HasPrefix(path, protected) {
			return true
		}
	}
	return false
}

// ListFiles 列出目录文件
// @Summary 列出目录文件
// @Description 获取指定目录下的文件和文件夹列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param path query string false "目录路径" default("/home")
// @Success 200 {object} handler.Response{data=object{files=[]models.FileInfo,currentPath=string}} "获取成功"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files [get]
func ListFiles(c *gin.Context) {
	path := c.DefaultQuery("path", "/home")

	if isProtectedPath(path) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	files, err := os.ReadDir(path)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var fileInfos []models.FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(path, file.Name())
		stat := info.Sys().(*syscall.Stat_t)

		fileInfo := models.FileInfo{
			Name:        file.Name(),
			Path:        fullPath,
			Size:        info.Size(),
			Mode:        info.Mode().String(),
			ModTime:     info.ModTime(),
			IsDir:       file.IsDir(),
			Permissions: fmt.Sprintf("%o", info.Mode().Perm()),
			Owner:       fmt.Sprintf("%d", stat.Uid),
			Group:       fmt.Sprintf("%d", stat.Gid),
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	handler.Respond(c, http.StatusOK, nil, gin.H{
		"files":       fileInfos,
		"currentPath": path,
	})
}

// DownloadFile 下载文件
// @Summary 下载文件
// @Description 下载指定路径的文件
// @Tags 文件管理
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param path query string true "文件路径"
// @Success 200 {file} file "文件内容"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 404 {object} handler.Response "文件不存在"
// @Router /auth/files/download [get]
func DownloadFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		handler.Respond(c, http.StatusBadRequest, "参数为空", nil)
		return
	}

	if isProtectedPath(filePath) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		handler.Respond(c, http.StatusNotFound, "文件不存在", nil)
		return
	}

	if info.IsDir() {
		handler.Respond(c, http.StatusBadRequest, "文件夹不允许下载", nil)
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)
}

// UploadFile 上传文件
// @Summary 上传文件
// @Description 上传文件到指定目录
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param path formData string true "目标目录路径"
// @Param file formData file true "上传的文件"
// @Success 200 {object} handler.Response "上传成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/upload [post]
func UploadFile(c *gin.Context) {
	targetDir := c.PostForm("path")
	if targetDir == "" {
		targetDir = "/tmp"
	}

	if isProtectedPath(targetDir) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "获取上传文件失败", nil)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	targetPath := filepath.Join(targetDir, header.Filename)

	out, err := os.Create(targetPath)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "创建文件失败", nil)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {

		}
	}(out)

	_, err = io.Copy(out, file)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "保存文件失败", nil)
		return
	}

	handler.Respond(c, http.StatusOK, "上传成功", nil)
}

// MoveFile 移动文件
// @Summary 移动文件
// @Description 将文件从源路径移动到目标路径
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{source=string,destination=string} true "源路径和目标路径"
// @Success 200 {object} handler.Response "移动成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/move [post]
func MoveFile(c *gin.Context) {
	var req struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, nil, nil)
		return
	}

	if isProtectedPath(req.Source) || isProtectedPath(req.Target) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	err := os.Rename(req.Source, req.Target)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "移动成功", nil)
}

// CopyFile 复制文件
// @Summary 复制文件
// @Description 将文件从源路径复制到目标路径
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{source=string,destination=string} true "源路径和目标路径"
// @Success 200 {object} handler.Response "复制成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/copy [post]
func CopyFile(c *gin.Context) {
	var req struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, nil, nil)
		return
	}

	if isProtectedPath(req.Source) || isProtectedPath(req.Target) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	err := copyFile(req.Source, req.Target)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "复制成功", nil)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(sourceFile *os.File) {
		err := sourceFile.Close()
		if err != nil {

		}
	}(sourceFile)

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {

		}
	}(destFile)

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定路径的文件或目录
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param path query string true "文件路径"
// @Success 200 {object} handler.Response "删除成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files [delete]
func DeleteFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		handler.Respond(c, http.StatusBadRequest, nil, nil)
		return
	}

	if isProtectedPath(filePath) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	err := os.RemoveAll(filePath)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "删除成功", nil)
}

// CreateDirectory 创建目录
// @Summary 创建目录
// @Description 在指定路径创建新目录
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{path=string} true "目录路径"
// @Success 200 {object} handler.Response "创建成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/mkdir [post]
func CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "拒绝访问", nil)
		return
	}

	fullPath := filepath.Join(req.Path, req.Name)

	if isProtectedPath(fullPath) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "创建成功", nil)
}

// CompressFiles 压缩文件
// @Summary 压缩文件
// @Description 将多个文件压缩为指定格式的压缩包
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{files=[]string,output=string,format=string} true "文件列表、输出路径和压缩格式"
// @Success 200 {object} handler.Response "压缩成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/compress [post]
func CompressFiles(c *gin.Context) {
	var req struct {
		Files      []string `json:"files"`
		OutputPath string   `json:"outputPath"`
		Format     string   `json:"format"` // zip, tar, tar.gz
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, nil, nil)
		return
	}

	if isProtectedPath(req.OutputPath) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	var err error
	switch req.Format {
	case "zip":
		err = createZip(req.Files, req.OutputPath)
	case "tar":
		err = createTar(req.Files, req.OutputPath)
	case "tar.gz":
		err = createTarGz(req.Files, req.OutputPath)
	default:
		handler.Respond(c, http.StatusBadRequest, "格式不支持", nil)
		return
	}

	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "压缩完毕", nil)
}

func createZip(files []string, output string) error {
	zipFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(zipFile *os.File) {
		err := zipFile.Close()
		if err != nil {

		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		err := zipWriter.Close()
		if err != nil {

		}
	}(zipWriter)

	for _, file := range files {
		if isProtectedPath(file) {
			continue
		}

		err := addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func createTar(files []string, output string) error {
	tarFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(tarFile *os.File) {
		err := tarFile.Close()
		if err != nil {

		}
	}(tarFile)

	tarWriter := tar.NewWriter(tarFile)
	defer func(tarWriter *tar.Writer) {
		err := tarWriter.Close()
		if err != nil {

		}
	}(tarWriter)

	for _, file := range files {
		if isProtectedPath(file) {
			continue
		}

		err := addFileToTar(tarWriter, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToTar(tarWriter *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	return err
}

func createTarGz(files []string, output string) error {
	tarFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func(tarFile *os.File) {
		err := tarFile.Close()
		if err != nil {

		}
	}(tarFile)

	gzWriter := gzip.NewWriter(tarFile)
	defer func(gzWriter *gzip.Writer) {
		err := gzWriter.Close()
		if err != nil {

		}
	}(gzWriter)

	tarWriter := tar.NewWriter(gzWriter)
	defer func(tarWriter *tar.Writer) {
		err := tarWriter.Close()
		if err != nil {

		}
	}(tarWriter)

	for _, file := range files {
		if isProtectedPath(file) {
			continue
		}

		err := addFileToTar(tarWriter, file)
		if err != nil {
			return err
		}
	}

	return nil
}

// ExtractFiles 解压文件
// @Summary 解压文件
// @Description 解压压缩文件到指定目录
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{filePath=string,destination=string} true "压缩文件路径和解压目标目录"
// @Success 200 {object} handler.Response "解压成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/extract [post]
func ExtractFiles(c *gin.Context) {
	var req struct {
		FilePath   string `json:"filePath"`
		OutputPath string `json:"outputPath"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, nil, nil)
		return
	}

	if isProtectedPath(req.FilePath) || isProtectedPath(req.OutputPath) {
		handler.Respond(c, http.StatusForbidden, "拒绝访问", nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(req.FilePath))
	var err error

	switch ext {
	case ".zip":
		err = extractZip(req.FilePath, req.OutputPath)
	case ".tar":
		err = extractTar(req.FilePath, req.OutputPath)
	case ".gz":
		if strings.HasSuffix(strings.ToLower(req.FilePath), ".tar.gz") {
			err = extractTarGz(req.FilePath, req.OutputPath)
		} else {
			err = fmt.Errorf("不支持的文件格式")
		}
	default:
		err = fmt.Errorf("不支持的文件格式")
	}

	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, nil, nil)
}

func extractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func(reader *zip.ReadCloser) {
		err := reader.Close()
		if err != nil {

		}
	}(reader)

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, file.FileInfo().Mode())
			if err != nil {
				return err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer func(fileReader io.ReadCloser) {
			err := fileReader.Close()
			if err != nil {

			}
		}(fileReader)

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer func(targetFile *os.File) {
			err := targetFile.Close()
			if err != nil {

			}
		}(targetFile)

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractTar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return err
			}
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer func(outFile *os.File) {
				err := outFile.Close()
				if err != nil {

				}
			}(outFile)

			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func(gzReader *gzip.Reader) {
		err := gzReader.Close()
		if err != nil {

		}
	}(gzReader)

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return err
			}
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer func(outFile *os.File) {
				err := outFile.Close()
				if err != nil {
					fmt.Printf("Error closing file %s: %v\n", outFile.Name(), err)
				}
			}(outFile)

			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetPermissions 获取文件权限
// @Summary 获取文件权限
// @Description 获取指定文件或目录的权限信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param path query string true "文件路径"
// @Success 200 {object} handler.Response{data=object{permissions=string,owner=string,group=string}} "获取成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/permissions [get]
func GetPermissions(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		handler.Respond(c, http.StatusBadRequest, "文件路径不能为空", nil)
		return
	}

	if isProtectedPath(filePath) {
		handler.Respond(c, http.StatusForbidden, "无法查看受保护文件的权限", nil)
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		handler.Respond(c, http.StatusNotFound, "文件不错哎", nil)
		return
	}

	stat := info.Sys().(*syscall.Stat_t)

	handler.Respond(c, http.StatusOK, gin.H{
		"permissions": fmt.Sprintf("%o", info.Mode().Perm()),
		"mode":        info.Mode().String(),
		"owner":       fmt.Sprintf("%d", stat.Uid),
		"group":       fmt.Sprintf("%d", stat.Gid),
		"size":        info.Size(),
		"modTime":     info.ModTime(),
	}, nil)
}

// SetPermissions 设置文件权限
// @Summary 设置文件权限
// @Description 设置指定文件或目录的权限
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{path=string,permissions=string,recursive=bool} true "文件路径、权限和是否递归"
// @Success 200 {object} handler.Response "设置成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/permissions [post]
func SetPermissions(c *gin.Context) {
	var req struct {
		Path        string `json:"path"`
		Permissions string `json:"permissions"`
		Owner       string `json:"owner,omitempty"`
		Group       string `json:"group,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误", nil)
		return
	}

	if isProtectedPath(req.Path) {
		handler.Respond(c, http.StatusForbidden, "无法修改受保护文件的权限", nil)
		return
	}

	// 设置文件权限
	perm, err := strconv.ParseUint(req.Permissions, 8, 32)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "权限格式错误", nil)
		return
	}

	err = os.Chmod(req.Path, os.FileMode(perm))
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, "设置权限失败: "+err.Error(), nil)
		return
	}

	// 设置所有者和组（如果提供）
	if req.Owner != "" || req.Group != "" {
		uid := -1
		gid := -1

		if req.Owner != "" {
			if u, err := strconv.Atoi(req.Owner); err == nil {
				uid = u
			}
		}

		if req.Group != "" {
			if g, err := strconv.Atoi(req.Group); err == nil {
				gid = g
			}
		}

		err = os.Chown(req.Path, uid, gid)
		if err != nil {
			handler.Respond(c, http.StatusInternalServerError, "设置所有者失败: "+err.Error(), nil)
			return
		}
	}

	handler.Respond(c, http.StatusOK, "权限设置成功", nil)
}

// GetFileContent 获取文件内容（用于编辑文本文件）
// @Summary 获取文件内容
// @Description 获取文本文件的内容用于编辑
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param path query string true "文件路径"
// @Success 200 {object} handler.Response{data=object{content=string}} "获取成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/content [get]
func GetFileContent(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		handler.Respond(c, http.StatusBadRequest, "文件路径不能为空", nil)
		return
	}
	decodedPath, err := url.QueryUnescape(filePath)
	if err != nil {
		handler.Respond(c, http.StatusBadRequest, "文件路径解码失败", nil)
		return
	}
	filePath = decodedPath

	if isProtectedPath(filePath) {
		handler.Respond(c, http.StatusForbidden, "无法查看受保护文件的内容", nil)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err, nil)
		return
	}
	fmt.Println(err)

	handler.Respond(c, http.StatusOK, "", gin.H{
		"content": string(content),
		"path":    filePath,
	})
}

// SaveFileContent 保存文件内容
// @Summary 保存文件内容
// @Description 保存编辑后的文本文件内容
// @Tags 文件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{path=string,content=string} true "文件路径和内容"
// @Success 200 {object} handler.Response "保存成功"
// @Failure 400 {object} handler.Response "请求参数错误"
// @Failure 401 {object} handler.Response "未授权"
// @Failure 403 {object} handler.Response "拒绝访问"
// @Failure 500 {object} handler.Response "服务器内部错误"
// @Router /auth/files/content [post]
func SaveFileContent(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Respond(c, http.StatusBadRequest, "请求参数错误", nil)
		return
	}

	if isProtectedPath(req.Path) {
		handler.Respond(c, http.StatusForbidden, "无法修改受保护的文件", nil)
		return
	}

	err := os.WriteFile(req.Path, []byte(req.Content), 0644)
	if err != nil {
		handler.Respond(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	handler.Respond(c, http.StatusOK, "文件保存成功", nil)
}
