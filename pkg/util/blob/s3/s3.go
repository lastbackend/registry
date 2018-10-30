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
	"bytes"
	"fmt"
	"github.com/lastbackend/registry/pkg/util/blob/config"
	"github.com/lastbackend/registry/pkg/util/generator"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	CONTAINER_LOGS_NAME = "logs"
	BUFFER_SIZE         = 1024
)

type Driver struct {
	objectACL     string
	bucketName    string
	rootDirectory string
	region        string
	client        *s3.S3
}

func New(params config.Config) *Driver {

	opts := aws.NewConfig()
	sess, err := session.NewSession()
	if err != nil {
		return nil
	}

	creds := credentials.NewChainCredentials([]credentials.Provider{
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     params.AccessID,
				SecretAccessKey: params.SecretKey,
				SessionToken:    params.SessionToken,
			},
		},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{},
		&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(sess)},
	})

	if params.Endpoint != "" {
		opts.WithS3ForcePathStyle(true)
		opts.WithEndpoint(params.Endpoint)
	}

	opts.WithCredentials(creds)
	opts.WithRegion(params.Region)
	opts.WithDisableSSL(!params.Secure)

	sess, err = session.NewSession(opts)
	if err != nil {
		return nil
	}
	s3obj := s3.New(sess)
	return &Driver{
		client:        s3obj,
		bucketName:    params.Bucket,
		rootDirectory: params.RootDirectory,
	}
}

func (d *Driver) Read(path string, writer io.WriteCloser) error {
	reader, err := d.reader(d.makePath(path), 0)
	if err != nil {
		return err
	}
	io.Copy(writer, reader)
	return nil
}

func (d *Driver) Write(path string, contents []byte) error {
	return d.write(path, bytes.NewReader(contents))
}

func (d *Driver) WriteFromReader(path string, reader io.Reader) error {

	var buffer = make([]byte, BUFFER_SIZE)
	var u = generator.GetUUIDV4()

	for {
		select {
		default:

			n, err := reader.Read(buffer)
			if err != nil && io.EOF == err {
				return nil
			}
			if err != nil && io.EOF != err {
				return err
			}

			partNumber := aws.Int64(int64(n))

			_, err = func(p []byte) (n int, err error) {
				d.write(path, bytes.NewReader(p))

				d.client.UploadPart(&s3.UploadPartInput{
					Bucket:     aws.String(d.bucketName),
					Key:        aws.String(d.makePath(path)),
					PartNumber: partNumber,
					UploadId:   aws.String(u),
					Body:       bytes.NewReader(p),
				})

				return n, nil
			}(buffer[0:n])

			if err != nil {
				return err
			}

			for i := 0; i < n; i++ {
				buffer[i] = 0
			}
		}
	}

	return nil
}

func (d *Driver) WriteFromFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file %s not exists", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.write(path, f)
}

func (d *Driver) reader(path string, offset int64) (io.ReadCloser, error) {
	resp, err := d.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(d.bucketName),
		Key:    aws.String(d.makePath(path)),
		Range:  aws.String("bytes=" + strconv.FormatInt(offset, 10) + "-"),
	})

	if err != nil {
		if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "InvalidRange" {
			return ioutil.NopCloser(bytes.NewReader(nil)), nil
		}

		return nil, err
	}
	return resp.Body, nil
}

func (d *Driver) makePath(path string) string {
	return strings.TrimLeft(strings.TrimRight(d.rootDirectory, string(os.PathSeparator))+path, string(os.PathSeparator))
}

func (d *Driver) getContentType() *string {
	return aws.String("application/octet-stream")
}

func (d *Driver) getACL() *string {
	return aws.String(d.objectACL)
}

func (d *Driver) write(path string, reader io.ReadSeeker) error {
	_, err := d.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(d.bucketName),
		Key:         aws.String(d.makePath(path)),
		ContentType: d.getContentType(),
		Body:        reader,
	})
	if err != nil {
		return err
	}
	return nil
}
