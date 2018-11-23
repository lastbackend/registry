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

package distribution

import (
	"context"
	"fmt"
	"github.com/lastbackend/lastbackend/pkg/distribution/errors"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/registry/pkg/distribution/types"
	"github.com/lastbackend/registry/pkg/storage"
)

const (
	logSystemPrefix = "distribution:system"
)

type ISystem interface {
	Get() (*types.System, error)
	Update(system *types.System, opts *types.SystemUpdateOptions) error
	UpdateController(system *types.System, opts *types.SystemUpdateControllerOptions) error
	UpdateControllerLastEvent(system *types.System, opts *types.SystemUpdateControllerLastEventOptions) error
}

type System struct {
	context context.Context
	storage storage.IStorage
}

func (s System) Get() (*types.System, error) {
	log.V(logLevel).Debugf("%s:get:> get system", logSystemPrefix)

	sys, err := s.storage.System().Get(s.context)
	if err != nil {
		log.V(logLevel).Debugf("%s:get:> get system err: %v", logSystemPrefix, err)
		return nil, err
	}

	return sys, nil
}

func (s System) Update(system *types.System, opts *types.SystemUpdateOptions) error {

	if system == nil {
		return errors.New("invalid argument")
	}

	if opts == nil {
		opts = new(types.SystemUpdateOptions)
	}

	log.V(logLevel).Debugf("%s:update:> update system %#v", logSystemPrefix, opts)

	if opts.AccessToken != nil {
		system.AccessToken = *opts.AccessToken
	}

	if opts.AuthServer != nil {
		system.AuthServer = *opts.AuthServer
	}

	if err := s.storage.System().Update(s.context, system); err != nil {
		log.V(logLevel).Errorf("%s:update:> update image err: %v", logSystemPrefix, err)
		return err
	}

	return nil
}

func (s System) UpdateController(system *types.System, opts *types.SystemUpdateControllerOptions) error {

	if system == nil {
		return errors.New("invalid argument")
	}

	if opts == nil {
		opts = new(types.SystemUpdateControllerOptions)
	}

	if len(opts.Hostname) == 0 {
		log.V(logLevel).Warnf("%s:update_controller:> hostname is empty", logSystemPrefix)
		return nil
	}

	if opts.Pid == 0 {
		log.V(logLevel).Warnf("%s:update_controller:> pid id zero", logSystemPrefix)
		return nil
	}

	log.V(logLevel).Debugf("%s:update_controller:> update controller %s", logSystemPrefix, opts.Hostname)

	system.CtrlMaster = fmt.Sprintf("%d:%s", opts.Pid, opts.Hostname)

	if err := s.storage.System().UpdateControllerMaster(s.context, system); err != nil {
		log.V(logLevel).Errorf("%s:update_controller:> update controller err: %v", logSystemPrefix, err)
		return err
	}

	return nil
}

func (s System) UpdateControllerLastEvent(system *types.System, opts *types.SystemUpdateControllerLastEventOptions) error {

	if system == nil {
		return errors.New("invalid argument")
	}

	system.CtrlLastEvent = &opts.LastEvent

	if err := s.storage.System().UpdateControllerLastEvent(s.context, system); err != nil {
		log.V(logLevel).Errorf("%s:update_controller_last_event:> update controller err: %v", logSystemPrefix, err)
		return err
	}

	return nil
}

func NewSystemModel(ctx context.Context, stg storage.IStorage) ISystem {
	return &System{ctx, stg}
}
