package api

import (
	"cider/config"
	"cider/gossip"
	"cider/log"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
)

// TODO these should be moved to Task.Requirements
const MinCores = 1
const MinRam = 1000000000 // 1 GB
const MaxLoad = 500000

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

// hasSufficientResources: Determines if a node has enough resources to run a given task
// TODO this function should determine this using Task.Requirements
func hasSufficientResources(ip string) bool {
	if member, ok := gossip.Self.MembershipList[ip]; ok {
		return (member.Profile.Cores >= MinCores) && (member.Profile.Ram >= MinRam) && (member.Profile.Load <= MaxLoad)
	}
	return false
}

// findCapableRemoteCiderNode: Find a remote CIDER node that meets the minimum resource requirements for a given task
func findCapableRemoteCiderNode() (string, error) {
	maxScore := -1.0
	suitableNode := ""

	for ip, node := range gossip.Self.MembershipList {
		if hasSufficientResources(ip) {
			cores := float64(node.Profile.Cores)
			// Adding a small delta to avoid potential divide by 0 error
			load := float64(node.Profile.Load) + 0.0000000001
			memory := float64(node.Profile.Ram)
			reputation := float64(node.Profile.Reputation)

			effectiveLoad := load / cores

			// TODO: (potential) We need scaling parameters to adjust
			//   priorities to each of the three factors:
			// 	 		memory, effective load, reputation.
			//   The scaling params should be based on the task that we are
			//   looking to deploy.
			score := (memory / effectiveLoad) + reputation

			if score > maxScore {
				maxScore = score
				suitableNode = ip
			}
		}
	}
	if suitableNode == "" {
		return suitableNode, errors.New("cannot find a capable remote CIDER node")
	}
	return suitableNode, nil
}

// deployTaskRemotely: Try to deploy the task to a remote CIDER node
func deployTaskRemotely(requestToForward *http.Request) (string, int) {
	remoteNodeIp, err := findCapableRemoteCiderNode()
	if err != nil {
		log.Warning(err)
		return "", http.StatusInternalServerError
	}
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
		var task Task
		err = json.Unmarshal(body, &task)
		if err != nil {
			log.Warning(err)
			return "", http.StatusInternalServerError
		}
		log.Info("Redirected request to remote CIDER node:", remoteNodeIp)
		return remoteUrl + "/" + task.Id, http.StatusOK
	}
	return "", (*response).StatusCode
}
