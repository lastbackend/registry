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
	"github.com/lastbackend/registry/pkg/controller/envs"
	"github.com/lastbackend/registry/pkg/distribution"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/events"
	"github.com/lastbackend/registry/pkg/log"
	"github.com/lastbackend/registry/pkg/rpc"
)

// SetBuilderOnlineHandler - handler set builder online
func SetBuilderOnlineHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {
	log.Debug("Build Controller: SetBuilderOnlineHandler: set builder online handler")

	event := types.BuilderOnlineEvent{}

	if err := json.Unmarshal(payload, &event); err != nil {
		log.Errorf("Build Controller: SetBuilderOnlineHandler: parse incoming data: %s", err)
		return err
	}

	// TODO: save status for builder to storage

	bm := distribution.NewBuildModel(ctx, envs.Get().GetStorage())
	builds, err := bm.ListActiveByBuilder(event.Payload.Builder)
	if err != nil {
		log.Errorf("Build Controller: SetBuilderOnlineHandler: can-not get pod by ID: %s", err.Error())
		return err
	}

	for _, build := range builds {
		job, err := bm.CreateJob(build)
		if err != nil {
			log.Errorf("Build Controller: SetBuilderOnlineHandler: create job err: %s", err.Error())
			return err
		}

		events.BuildExecuteRequest(envs.Get().GetRPC(), job)
	}

	return nil
}

// SetBuilderOfflineHandler - handler set builder online
func SetBuilderOfflineHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {
	log.Debug("Build Controller: SetBuilderOfflineHandler: set builder offline handler")

	event := types.BuilderOfflineEvent{}

	if err := json.Unmarshal(payload, &event); err != nil {
		log.Errorf("Build Controller: SetBuilderOnlineHandler: parse incoming data: %s", err)
		return err
	}

	// TODO: save status for builder to storage

	return nil
}
