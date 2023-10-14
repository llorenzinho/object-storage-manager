package storage

// Errors that says that the file does not exist
type FileDoesNotExistError struct {
	bucket string
	file   string
}

func (e *FileDoesNotExistError) Error() string {
	return e.file + " does not exist in bucket" + e.bucket
}

// check if the error is a FileDoesNotExistError
func IsFileDoesNotExistError(err error) bool {
	_, ok := err.(*FileDoesNotExistError)
	return ok
}

// New file does not exist error
func NewFileDoesNotExistError(bucket string, file string) *FileDoesNotExistError {
	return &FileDoesNotExistError{file: file, bucket: bucket}
}

// File already exists error
type FileAlreadyExistsError struct {
	bucket string
	file   string
}

func (e *FileAlreadyExistsError) Error() string {
	return e.file + " already exists in bucket" + e.bucket
}

// check if the error is a FileAlreadyExistsError
func IsFileAlreadyExistsError(err error) bool {
	_, ok := err.(*FileAlreadyExistsError)
	return ok
}

// New file already exists error
func NewFileAlreadyExistsError(bucket string, file string) *FileAlreadyExistsError {
	return &FileAlreadyExistsError{file: file, bucket: bucket}
}
