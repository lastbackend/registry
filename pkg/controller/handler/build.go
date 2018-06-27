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

// BuildProvisionHandler - handler provision state event from api
func BuildProvisionHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {
	log.Debug("Build Controller: BuildProvisionHandler: provision state handler")

	event := types.BuildProvisionEvent{}

	if err := json.Unmarshal(payload, &event); err != nil {
		log.Errorf("Parse incoming data: %s", err)
		return err
	}

	bm := distribution.NewBuildModel(ctx, envs.Get().GetStorage())

	build, err := bm.SetProvision(event.Payload.RepoID, event.Payload.Tag)
	if err != nil {
		log.Errorf("Build provision: %s", err)
		return err
	}

	if build == nil {
		log.Debugf("Build: build provision for %s/%s not found", event.Payload.RepoID, event.Payload.Tag)
		return nil
	}

	job, err := bm.CreateJob(build)
	if err != nil {
		log.Errorf("Build provision: create job %s", err)
		return err
	}

	if job == nil {
		log.Debugf("Build: job not found for %s:%s", event.Payload.RepoID, event.Payload.Tag)
		return nil
	}

	return events.BuildExecuteRequest(envs.Get().GetRPC(), job)
}

// LogStreamRequestHandler - handler request for log streaming
func BuildLogStreamHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {

	log.Debug("Build Controller: log stream request handler")

	event := types.BuildLogsEvent{}

	if err := json.Unmarshal(payload, &event); err != nil {
		log.Errorf("Parse incoming data: %s", err)
		return err
	}

	bm := distribution.NewBuildModel(ctx, envs.Get().GetStorage())

	b, err := bm.Get(event.Payload.Build)
	if err != nil {
		log.Errorf("Build Controller: log request: get build by id err: %s", err)
		return err
	}
	if b == nil {
		log.Errorf("Build Controller: can not find requested build: %s", event.Payload.Build)
		return nil
	}

	e := types.BuildTaskLogsEvent{
		Payload: types.BuildTaskLogsEventPayload{
			Task:  b.Meta.Task,
			Token: event.Payload.Token,
			URI:   event.Payload.URI,
		},
	}

	// Send build to builder for building image
	endpoint := rpc.Destination{Name: types.KindBuilder, UUID: b.Meta.Builder, Handler: types.BuildTaskLogsEventName}

	if err := envs.Get().GetRPC().Call(endpoint, e); err != nil {
		log.Errorf("Build logs: %s", err)
		return err
	}

	return nil
}

// LogStreamRequestHandler - handler request for log streaming
func BuildCancelHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {

	log.Debug("Build Controller: log stream request handler")

	var event = types.BuildCancelEvent{}
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Errorf("Parse incoming data: %s", err)
		return err
	}

	bm := distribution.NewBuildModel(ctx, envs.Get().GetStorage())

	b, err := bm.Get(event.Payload.Build)
	if err != nil {
		log.Errorf("Build Controller: log request: get build by id err: %s", err)
		return err
	}
	if b == nil {
		log.Errorf("Build Controller: can not find requested build: %s", event.Payload.Build)
		return nil
	}

	if b.State.Done {
		log.Warn("Build Controller: the build process is already complete: %s", event.Payload.Build)
		return nil
	}

	e := types.BuildTaskCancelEvent{
		Payload: types.BuildTaskCancelEventPayload{
			Task: b.Meta.Task,
		},
	}
	// Send cancel event to builder
	endpoint := rpc.Destination{Name: types.KindBuilder, UUID: b.Meta.Builder, Handler: types.BuildTaskCancelEventName}

	return envs.Get().GetRPC().Call(endpoint, e)
}

// SetBuildStateHandler - handle request for update build state
func BuildStateHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {
	log.Debug("Build Controller: BuildStateHandler: set finish state handler")

	var e = new(types.BuildStateBuilderEvent)
	if err := json.Unmarshal(payload, &e); err != nil {
		log.Errorf("Parse incoming data: %s", err)
		return err
	}

	log.Debugf("Update build state: %s", e.ID)
	bm := distribution.NewBuildModel(ctx, envs.Get().GetStorage())

	build, err := bm.Get(e.Payload.Build)
	if err != nil {
		log.Errorf("Build Controller: BuildStateHandler: get build by id err: %s", err)
		return err
	}
	if build == nil {
		log.Warn("Build Controller: BuildStateHandler: build  not found")
		return nil
	}
	if e.Payload.State.Step != types.BuildStepBuild && len(build.Meta.Task) != 0 && build.Meta.Task != e.Payload.Task {
		return nil
	}

	switch true {
	case e.Payload.State.Canceled:
		build.MarkAsCanceled(e.Payload.State.Step, e.Payload.State.Message)
	case e.Payload.State.Error:
		build.MarkAsError(e.Payload.State.Step, e.Payload.State.Message)
		// BuildStepFetch -  not supported
	case e.Payload.State.Step == types.BuildStepFetch:
		build.MarkAsFetching(e.Payload.State.Step, e.Payload.State.Message)
	case e.Payload.State.Step == types.BuildStepBuild:
		if err := bm.UpdateTaskInfo(e.Payload.Build, e.Payload.Builder, e.Payload.Task); err != nil {
			log.Errorf("Build Controller: BuildStateHandler: update task info for build: %s", err)
			return err
		}
		build.MarkAsBuilding(e.Payload.State.Step, e.Payload.State.Message)
	case e.Payload.State.Step == types.BuildStepUpload:
		build.MarkAsUploading(e.Payload.State.Step, e.Payload.State.Message)
	case e.Payload.State.Step == types.BuildStepDone && !e.Payload.State.Error:
		build.MarkAsDone(e.Payload.State.Step, e.Payload.State.Message)
	}

	if err := bm.UpdateState(build); err != nil {
		log.Errorf("Build Controller: BuildStateHandler: update state build: %s", err)
		return err
	}

	if build.State.Done || build.State.Canceled || build.State.Error {

		log.Debug("Build Controller: BuildStateHandler: update repo sources for services and call redeploy for it")

		if build.State.Done {

			// TODO: update service image and send events for redeploy

			//sm := distribution.NewServiceModel(ctx, envs.Get().GetStorage())

			r := new(types.ServiceSourcesRepo)
			r.ID = build.Repo.ID
			r.Tag = build.Image.Tag
			r.Build = build.Meta.ID

			//if err := sm.UpdateRepoSources(r); err != nil {
			//	log.Errorf("Build Controller: UpdateState: update repo sources for services: %s", err)
			//	return err
			//}

			//services, err := sm.ListByRepo(r.ID)
			//if err != nil {
			//	log.Errorf("Build Controller: UpdateState:  get services list err: %s", err)
			//	return err
			//}

			//log.Debugf("Build Controller: UpdateState:  send services to controller for redeploy")
			//for _, srv := range services {
			//	events.DeploymentCreateRequest(envs.Get().GetRPC(), srv.Meta.ID)
			//}
		}

		if err := events.BuildProvisionRequest(envs.Get().GetRPC(), build.Repo.ID, build.Image.Tag); err != nil {
			log.Errorf("Build Controller: UpdateState:  send event for provision build err: %s", err.Error())
			return err
		}
	}

	return nil
}

// BuildImageInfoHandler - handle request for update image info
func BuildImageInfoHandler(ctx context.Context, _ rpc.Sender, payload []byte) error {

	log.Debug("Build Controller: BuildImageInfoHandler: set image info handler")

	var e = new(types.BuildImageInfoBuilderEvent)
	if err := json.Unmarshal(payload, &e); err != nil {
		log.Errorf("Parse incoming data: %s", err)
		return err
	}

	info := e.Payload.Info

	stats := new(types.BuildInfo)
	stats.ImageHash = info.ID
	stats.Size = info.VirtualSize

	bm := distribution.NewBuildModel(ctx, envs.Get().GetStorage())

	if err := bm.UpdateInfo(e.Payload.Build, stats); err != nil {
		log.Errorf("Build Controller: UpdateInfo: update image info: %s", err)
		return err
	}

	return nil
}
