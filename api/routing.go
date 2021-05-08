package api

import (
	"cider/config"
	"cider/gossip"
	"cider/log"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
)

// isLocalIp: Check if IP refers to this node itself
func isLocalIp(ip string) bool {
	return (net.ParseIP(ip).IsLoopback() || (ip == gossip.Self.IpAddress))
}

// isTrustedRemote: Return whether or not we trust this remote node
func isTrustedRemote(ip string) bool {
	// TODO Add authentication/authorization logic
	// Currently, a trusted node is just one that's in our membership list
	_, ok := gossip.Self.MembershipList[ip]
	return ok 
}

func isUntrustedSource(request *http.Request, response *http.ResponseWriter) bool {
	requestSourceIp, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		log.Warning(err)
		(*response).WriteHeader(http.StatusBadRequest)
		return true
	}
	trusted := isLocalIp(requestSourceIp) || isTrustedRemote(requestSourceIp)
	if !trusted {
		log.Warning("Denied", request.Method, request.URL.Path, "from untrusted source", requestSourceIp)
		(*response).WriteHeader(http.StatusNotFound)
	}
	return !trusted
}

func insufficientLocalResources() bool {
	// FIXME look at cores and RAM and load
	return (gossip.Self.IpAddress != "172.22.94.255")
}

// TODO should take task's expected resource requirements into consideration
func findRemoteCiderNode() string {
	// FIXME 
	return "172.22.94.255"
}

// deployTaskRemotely: Try to deploy the task to a remote CIDER node
func deployTaskRemotely(requestToForward *http.Request) (string, int) {
	remoteNodeIp := findRemoteCiderNode()
	remoteUrl := "http://" + remoteNodeIp + ":" + strconv.Itoa(config.ApiPort) + "/tasks"

	// PUT the task to the remote CIDER node
	request, err := http.NewRequest(http.MethodPut, remoteUrl, requestToForward.Body)
	if err != nil {
		log.Warning(err)
		return "", http.StatusInternalServerError
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Warning(err)
		return "", http.StatusInternalServerError
	}

	// if successful PUT, return a redirect URL for the remote task deployment
	if (*response).StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll((*response).Body)
		if err != nil {
			log.Warning(err)
			return "", http.StatusInternalServerError
		}
		log.Debug(string(body))
		var task Task
		err = json.Unmarshal(body, &task)
		if err != nil {
			log.Warning(err)
			return "", http.StatusInternalServerError
		}
		return remoteUrl + "/" + task.Id, http.StatusOK
	} 
	return "", (*response).StatusCode	
}
