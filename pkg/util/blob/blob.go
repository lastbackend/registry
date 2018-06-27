//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2018] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package blob

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/util/stream"
	"io"
)

const (
	CONTAINER_LOGS_NAME = "logs"
	BUFFER_SIZE         = 1024
)

type blob struct {
	client *storage.BlobStorageClient
}

func NewClient(accountName, accountKey string) *blob {
	client, err := storage.NewBasicClient(accountName, accountKey)
	if err != nil {
		panic(err)
	}
	cli := client.GetBlobService()

	return &blob{client: &cli}
}

func (b *blob) ReadToWriter(containerName, blockBlobName string, writer io.WriteCloser) error {

	cnt := b.client.GetContainerReference(containerName)
	bl := cnt.GetBlobReference(blockBlobName)
	readCloser, err := bl.Get(nil)
	defer readCloser.Close()
	if err != nil {
		return fmt.Errorf("get blob failed: %s", err)
	}

	if _, err := io.Copy(writer, readCloser); err != nil {
		return fmt.Errorf("copy file err: %s", err)
	}

	return nil
}

func (b *blob) ReadToStream(containerName, blockBlobName string, stream *stream.Stream) error {

	var buffer = make([]byte, BUFFER_SIZE)

	cnt := b.client.GetContainerReference(containerName)

	exists, err := cnt.Exists()
	if err != nil {
		return fmt.Errorf("create container err: %s", err)
	}

	if !exists {
		return nil
	}

	bl := cnt.GetBlobReference(blockBlobName)
	readCloser, err := bl.Get(nil)
	if err != nil {
		return fmt.Errorf("get blob failed: %s", err)
	}
	defer readCloser.Close()

	for {
		readBytes, err := readCloser.Read(buffer)
		if err != nil && err != io.EOF {
			log.Warnf("read bytes from reader err: %s", err)
		}
		if readBytes == 0 {
			break
		}

		_, err = func(p []byte) (n int, err error) {
			n, err = stream.Write(p)
			if err != nil {
				log.Errorf("write bytes to stream err: %s", err)
				return n, err
			}
			stream.Flush()
			return n, nil
		}(buffer[0:readBytes])

		if err != nil {
			log.Errorf("written to stream err: %s", err)
			break
		}

		for i := 0; i < readBytes; i++ {
			buffer[i] = 0
		}
	}

	return nil
}

// Creates a container, and performs operations with page blobs, append blobs and block blobs.
func (b *blob) Write(containerName, blobName string, reader io.Reader) error {
	fmt.Println("Create container with private access type...")

	cnt := b.client.GetContainerReference(containerName)
	options := storage.CreateContainerOptions{}
	_, err := cnt.CreateIfNotExists(&options)
	if err != nil {
		return fmt.Errorf("create container err: %s", err)
	}

	fmt.Println("Create an empty append blob...")

	bl := cnt.GetBlobReference(blobName)
	bl.Properties.ContentType = "text/plain"

	opts := &storage.PutBlobOptions{}
	err = bl.CreateBlockBlobFromReader(reader, opts)
	if err != nil {
		log.Errorf("Can not create blob from reader: %s", err)
		return err
	}

	return nil
}
