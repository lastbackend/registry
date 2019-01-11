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

package types

import (
	"context"
	"io"
)

type IBuilder interface {
	Start() error
	BuildLogs(ctx context.Context, pid string, stream io.Writer) error
	BuildCancel(ctx context.Context, pid string) error
	ActiveWorkers() int
	SetReserveMemory(memory string) error
	SetWorkerLimits(instances int, ram, cpu string) error
	Shutdown()
	Done() <-chan bool
}

type BuilderManifest struct {
	Limits *BuilderLimits
}

type BuilderLimits struct {
	Workers   uint
	WorkerRAM int64
	WorkerCPU int64
}
