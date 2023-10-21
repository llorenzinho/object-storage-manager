package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llorenzinho/object-storage-manager/utils/storage"
)

// Gin router group for files
func FilesRouter(r gin.IRouter) {
	g := r.Group("/files/api")
	v := g.Group("/v1")

	// No auth required
	v.GET("/:bucket", ListFiles)
	v.GET("/:bucket/:file", GetFile)

	// Auth required
	v.GET("download/:bucket/:file", DownloadFile)
	v.POST("/:bucket", UploadFile)

	// Admin required
	v.DELETE("/:bucket/:file", DeleteFile)
	v.GET("verify/:bucket/:file", VerifyFile)
	v.GET("unverify/:bucket/:file", UnverifyFiles)

}

// Gin function handler to list all files in a bucket
func ListFiles(c *gin.Context) {
	// get storage instance
	f := storage.Instance()
	bucket := c.Param("bucket")
	files, err := f.GetAll(bucket, c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

// Gin function handler to get a file from a bucket
func GetFile(c *gin.Context) {
	// get storage instance
	f := storage.Instance()
	bucket := c.Param("bucket")
	file := c.Param("file")
	fileData, err := f.Get(bucket, file, c.Request.Context())
	if err != nil {
		if storage.IsFileDoesNotExistError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fileData)
}

// Gin function handler to upload a file to a bucket
func UploadFile(c *gin.Context) {
	// get storage instance
	f := storage.Instance()
	bucket := c.Param("bucket")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fileData, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer fileData.Close()
	b := make([]byte, file.Size)
	_, err = fileData.Read(b)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	retFile, err := f.Upload(bucket, file.Filename, b, file.Header.Get("Content-Type"), c.Request.Context())
	if err != nil {
		if storage.IsFileAlreadyExistsError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, retFile)
}

// Gin function handler to delete a file from a bucket
func DeleteFile(c *gin.Context) {
	// get storage instance
	f := storage.Instance()
	bucket := c.Param("bucket")
	file := c.Param("file")
	del, err := f.Delete(bucket, file, c.Request.Context())
	if err != nil {
		if storage.IsFileDoesNotExistError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "file deleted", "file": del})
}

// Gin function handler to verify a file from a bucket
func VerifyFile(c *gin.Context) {
	// get storage instance
	f := storage.Instance()
	bucket := c.Param("bucket")
	file := c.Param("file")
	fileData, err := f.Verify(bucket, file, c.Request.Context())
	if err != nil {
		if storage.IsFileDoesNotExistError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fileData)
}

// Gin function handler Unverify files from a bucket
func UnverifyFiles(c *gin.Context) {
	// get storage instance
	f := storage.Instance()
	bucket := c.Param("bucket")
	file := c.Param("file")
	files, err := f.Unverify(bucket, file, c.Request.Context())
	if err != nil {
		if storage.IsFileDoesNotExistError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

// Gin function handler to download a file from a bucket
func DownloadFile(c *gin.Context) {
	// get storage instance
	f := storage.Instance()
	bucket := c.Param("bucket")
	file := c.Param("file")
	minioObj, fileInstance, err := f.Download(bucket, file, c.Request.Context())
	if err != nil {
		if storage.IsFileDoesNotExistError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.DataFromReader(http.StatusOK, fileInstance.Size, fileInstance.ContentType, minioObj, map[string]string{})
}