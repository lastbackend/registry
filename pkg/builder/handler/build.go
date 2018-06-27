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

package handler

import (
	"context"
	"encoding/json"

	"github.com/lastbackend/registry/pkg/builder/envs"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/rpc"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
)

// BuildTaskExecuteHandler - handler called build task execute
func BuildTaskExecuteHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {

	log.Info("Handler: BuildTaskExecuteHandler: execute task for create image build")

	event := &types.BuildTaskExecuteEvent{}

	if err := json.Unmarshal(payload, event); err != nil {
		log.Errorf("Handler: BuildTaskExecuteHandler: parse incoming data: %s", err.Error())
		return err
	}

	if event.Payload.Job == nil {
		err := errors.New("error: job can't be nil")
		log.Errorf("Handler: BuildTaskExecuteHandler: check data payload: %s", err.Error())
		return err
	}

	if err := envs.Get().GetBuilder().NewTask(ctx, event.Payload.Job); err != nil {
		log.Errorf("Handler: BuildTaskExecuteHandler: create new task err: %s", err.Error())
		return err
	}

	return nil
}

// BuildTaskCancelHandler - handler called build task cancel
func BuildTaskCancelHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {

	log.Info("Handler: BuildTaskCancelHandler: cancel execute task")

	event := &types.BuildTaskCancelEvent{}

	if err := json.Unmarshal(payload, event); err != nil {
		log.Errorf("Handler: BuildTaskExecuteHandler: parse incoming data: %s", err.Error())
		return err
	}

	log.Debugf("Handler: BuildTaskExecuteHandler: cancel task [%s]", event.Payload.Task)

	if err := envs.Get().GetBuilder().CancelTask(ctx, event.Payload.Task); err != nil {
		log.Errorf("Handler: BuildTaskExecuteHandler: cancel task err: %s", err.Error())
		return err
	}

	return nil
}

// BuildLogsCancelHandler - handler for get logs stream
func BuildLogsHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {

	log.Info("Handler: BuildLogsHandler: get task logs stream")

	event := new(types.BuildTaskLogsEvent)

	if err := json.Unmarshal(payload, &event); err != nil {
		log.Errorf("Handler: BuildTaskExecuteHandler: parse incoming data: %s", err.Error())
		return err
	}

	log.Debugf("Handler: BuildTaskExecuteHandler: task [%s] task [%s] log to endpoint [%s] ", event.Payload.Task, event.Payload.Task, event.Payload.URI)

	if err := envs.Get().GetBuilder().LogsTask(ctx, event.Payload.Task, event.Payload.URI); err != nil {
		log.Errorf("Handler: BuildTaskExecuteHandler: get logs task err: %s", err.Error())
		return err
	}

	return nil
}
