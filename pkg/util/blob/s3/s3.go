//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2019] Last.Backend LLC
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
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/lastbackend/registry/pkg/util/blob/config"
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
	resp, err := d.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(d.bucketName),
		Key:    aws.String(d.makePath(path)),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(writer, resp.Body)
	return nil
}

func (d *Driver) Write(path string, contents []byte) error {
	return d.write(path, bytes.NewReader(contents))
}

func (d *Driver) WriteFromFile(path, filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("file %s not exists", path)
	}
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return d.write(path, file)
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
	resp, err := d.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(d.bucketName),
		Key:         aws.String(d.makePath(path)),
		ContentType: d.getContentType(),
		Body:        reader,
	})
	if err != nil {
		return err
	}
	fmt.Printf("response %s", awsutil.StringValue(resp))
	return nil
}
