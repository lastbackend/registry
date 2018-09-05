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

package azure

import (
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/lastbackend/lastbackend/pkg/log"
)

const (
	CONTAINER_LOGS_NAME = "logs"
	BUFFER_SIZE         = 1024
)

type Driver struct {
	containerName string
	client        *storage.BlobStorageClient
}

func New(endpoint, accountName, accountKey, containerName string, ssl bool) *Driver {
	client, err := storage.NewClient(accountName, accountKey, endpoint, "", ssl)
	if err != nil {
		panic(err)
	}
	cli := client.GetBlobService()

	return &Driver{client: &cli, containerName: containerName}
}

func (b *Driver) ReadToWriter(name string, writer io.WriteCloser) error {

	cnt := b.client.GetContainerReference(b.containerName)
	bl := cnt.GetBlobReference(name)
	readCloser, err := bl.Get(nil)
	defer readCloser.Close()
	if err != nil {
		return fmt.Errorf("get Driver failed: %s", err)
	}

	if _, err := io.Copy(writer, readCloser); err != nil {
		return fmt.Errorf("copy file err: %s", err)
	}

	return nil
}

// Creates a container, and performs operations with page blobs, append blobs and block blobs.
func (b *Driver) Write(name string, reader io.Reader) error {
	fmt.Println("create container with private access type...")

	cnt := b.client.GetContainerReference(b.containerName)
	options := storage.CreateContainerOptions{}
	_, err := cnt.CreateIfNotExists(&options)
	if err != nil {
		return fmt.Errorf("create container err: %s", err)
	}

	fmt.Println("create an empty append Driver...")

	bl := cnt.GetBlobReference(name)
	bl.Properties.ContentType = "text/plain"

	opts := &storage.PutBlobOptions{}
	err = bl.CreateBlockBlobFromReader(reader, opts)
	if err != nil {
		log.Errorf("can not create Driver from reader: %s", err)
		return err
	}

	return nil
}

func (b *Driver) WriteFile(name string, filePath string) error {
	return nil
}
