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

package s3

import (
	"fmt"
	"github.com/minio/minio-go"
	"io"
)

type Driver struct {
	client *minio.Client

	bucketName string
}

func New(endpoint, accessKey, secretKey, bucketName string, ssl bool) *Driver {
	d := new(Driver)

	client, err := minio.New(endpoint, accessKey, secretKey, ssl)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	d.bucketName = bucketName
	d.client = client

	return d
}

func (d *Driver) ReadToWriter(name string, writer io.WriteCloser) error {
	o, err := d.client.GetObject(d.bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, o)
	if err != nil {
		return err
	}

	return nil
}

func (d Driver) Write(name string, reader io.Reader) error {
	_, err := d.client.PutObject(d.bucketName, name, reader, -1, minio.PutObjectOptions{})
	return err
}

func (d Driver) WriteFile(name string, filePath string) error {
	_, err := d.client.FPutObject(d.bucketName, name, filePath, minio.PutObjectOptions{
		ContentType: "plain/text",
	})
	return err
}
