package model

import "os"

type UploadFile struct {
	FLocation string
	FType     string
	FData     *os.File
}
