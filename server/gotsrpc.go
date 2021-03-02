// Code generated by gotsrpc https://github.com/foomo/gotsrpc  - DO NOT EDIT.

package server

import (
	io "io"
	ioutil "io/ioutil"
	http "net/http"
	time "time"

	gotsrpc "github.com/foomo/gotsrpc"
	github_com_foomo_webgrapple_vo "github.com/foomo/webgrapple/vo"
)

type ServiceGoTSRPCProxy struct {
	EndPoint    string
	allowOrigin []string
	service     Service
}

func NewDefaultServiceGoTSRPCProxy(service Service, allowOrigin []string) *ServiceGoTSRPCProxy {
	return &ServiceGoTSRPCProxy{
		EndPoint:    "/___webgrapple-service",
		allowOrigin: allowOrigin,
		service:     service,
	}
}

func NewServiceGoTSRPCProxy(service Service, endpoint string, allowOrigin []string) *ServiceGoTSRPCProxy {
	return &ServiceGoTSRPCProxy{
		EndPoint:    endpoint,
		allowOrigin: allowOrigin,
		service:     service,
	}
}

// ServeHTTP exposes your service
func (p *ServiceGoTSRPCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for _, origin := range p.allowOrigin {
		// todo we have to compare this with the referer ... and only send one
		w.Header().Add("Access-Control-Allow-Origin", origin)
	}
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if r.Method != http.MethodPost {
		if r.Method == http.MethodOptions {
			return
		}
		gotsrpc.ErrorMethodNotAllowed(w)
		return
	}
	defer io.Copy(ioutil.Discard, r.Body) // Drain Request Body

	var args []interface{}
	funcName := gotsrpc.GetCalledFunc(r, p.EndPoint)
	callStats := gotsrpc.GetStatsForRequest(r)
	if callStats != nil {
		callStats.Func = funcName
		callStats.Package = "github.com/foomo/webgrapple/server"
		callStats.Service = "Service"
	}
	switch funcName {
	case "Remove":
		var (
			arg_serviceIDs []github_com_foomo_webgrapple_vo.ServiceID
		)
		args = []interface{}{&arg_serviceIDs}
		err := gotsrpc.LoadArgs(&args, callStats, r)
		if err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		removeErr := p.service.Remove(arg_serviceIDs)
		if callStats != nil {
			callStats.Execution = time.Now().Sub(executionStart)
		}
		gotsrpc.Reply([]interface{}{removeErr}, callStats, r, w)
		return
	case "Upsert":
		var (
			arg_services []*github_com_foomo_webgrapple_vo.Service
		)
		args = []interface{}{&arg_services}
		err := gotsrpc.LoadArgs(&args, callStats, r)
		if err != nil {
			gotsrpc.ErrorCouldNotLoadArgs(w)
			return
		}
		executionStart := time.Now()
		upsertErr := p.service.Upsert(arg_services)
		if callStats != nil {
			callStats.Execution = time.Now().Sub(executionStart)
		}
		gotsrpc.Reply([]interface{}{upsertErr}, callStats, r, w)
		return
	default:
		gotsrpc.ClearStats(r)
		http.Error(w, "404 - not found "+r.URL.Path, http.StatusNotFound)
	}
}
