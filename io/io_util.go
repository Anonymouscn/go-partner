package io_util

import "io"

// CloseReader 关闭 Reader
func CloseReader(reader io.ReadCloser) {
	if reader != nil {
		_ = reader.Close()
	}
}
