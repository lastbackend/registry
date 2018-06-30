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

package http

import (
	"context"
	"github.com/lastbackend/registry/pkg/distribution/types"
)

type EventClient struct {
	client RestClient
}

func (ec EventClient) SendTaskStatus(ctx context.Context, payload *types.TaskStatusBuilderEvent) error {
	return nil
}

func (ec EventClient) SendImageInfo(ctx context.Context, payload *types.ImageInfoBuilderEvent) error {
	return nil
}


func newEventClient(req RestClient) EventClient {
	return EventClient{client: req}
}
