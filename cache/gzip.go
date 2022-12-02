package cache

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func WriteCompressedResponseOrDecompress(w http.ResponseWriter, r *http.Request, compressedOutput []byte) {
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		_, err := w.Write(compressedOutput)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failed to write gzip output: %v", err)
		}
		return
	}

	decompressed, err := Decompress(compressedOutput)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to decompress output: %v", err)
		return
	}

	_, err = w.Write(decompressed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to write decompressed output: %v", err)
	}
}

func Decompress(input []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(input)
	reader, err := gzip.NewReader(buffer)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create gzip reader: %v", err.Error())
	}
	defer reader.Close()

	decompressed, err := ioutil.ReadAll(reader)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to decompress gzip payload: %v", err.Error())
	}

	return decompressed, nil
}
