package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lxc/lxd/shared"
	"github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/version"
)

func containersGet(d *Daemon, r *http.Request) Response {
	for i := 0; i < 100; i++ {
		result, err := doContainersGet(d, d.isRecursionRequest(r))
		if err == nil {
			return SyncResponse(true, result)
		}
		if !isDbLockedError(err) {
			shared.LogDebugf("DBERR: containersGet: error %q", err)
			return InternalError(err)
		}
		// 1 s may seem drastic, but we really don't want to thrash
		// perhaps we should use a random amount
		time.Sleep(100 * time.Millisecond)
	}

	shared.LogDebugf("DBERR: containersGet, db is locked")
	shared.PrintStack()
	return InternalError(fmt.Errorf("DB is locked"))
}

func doContainersGet(d *Daemon, recursion bool) (interface{}, error) {
	result, err := dbContainersList(d.db, cTypeRegular)
	if err != nil {
		return nil, err
	}

	resultString := []string{}
	resultList := []*api.Container{}
	if err != nil {
		return []string{}, err
	}

	for _, container := range result {
		if !recursion {
			url := fmt.Sprintf("/%s/containers/%s", version.APIVersion, container)
			resultString = append(resultString, url)
		} else {
			c, err := doContainerGet(d, container)
			if err != nil {
				c = &api.Container{
					Name:       container,
					Status:     api.Error.String(),
					StatusCode: api.Error}
			}
			resultList = append(resultList, c)
		}
	}

	if !recursion {
		return resultString, nil
	}

	return resultList, nil
}

func doContainerGet(d *Daemon, cname string) (*api.Container, error) {
	c, err := containerLoadByName(d, cname)
	if err != nil {
		return nil, err
	}

	cts, _, err := c.Render()
	if err != nil {
		return nil, err
	}

	return cts.(*api.Container), nil
}
