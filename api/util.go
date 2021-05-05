package api

import (
	"cider/config"
	"cider/exportapi"
	"cider/exportgossip"
	"cider/functions"
	"cider/handle"
	"cider/log"
	"cider/util"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// writeMessage: Write headers + formatted message to the response
func writeMessage(response *http.ResponseWriter, status int, format string, args ...interface{}) {
	message := []byte(fmt.Sprintf(format, args...))

	// changing the Header after calling WriteHeader(statusCode) has no effect,
	// so these lines must remain in this order
	(*response).Header().Add("Content-Type", "text/plain")
	(*response).WriteHeader(status)

	// HTTP response body is terminated with CRLF
	(*response).Write(append(message, '\r', '\n'))
}

// writeStruct: Write headers + struct (as json) to the response
func writeStruct(response *http.ResponseWriter, body interface{}) {
	jsonBody, err := json.Marshal(body)
	handle.Fatal(err)

	// changing the Header after calling WriteHeader(statusCode) has no effect,
	// so these lines must remain in this order
	(*response).Header().Add("Content-Type", "text/json")
	(*response).WriteHeader(http.StatusOK)

	// HTTP response body is terminated with CRLF
	(*response).Write(append(jsonBody, '\r', '\n'))
}

// generateUUID: Generate UUID using time, local ip
func generateUUID() (string, error) {
	input, err := time.Now().MarshalBinary()
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().UnixNano())
	nonce := make([]byte, config.NonceLength)
	rand.Read(nonce)
	input = append(input, nonce...)

	nodeIpAddress, err := util.GetIpAddress()
	if err != nil {
		return "", err
	}
	input = append(input, ([]byte(nodeIpAddress))...)

	return fmt.Sprintf("%x", sha256.Sum256(input)), nil
}

// isLocalIp: Check if IP is loopback or local IP
func isLocalIp(ip string) bool {
	if !net.ParseIP(ip).IsLoopback() {
		nodeIpAddress, err := util.GetIpAddress()
		if err != nil {
			log.Error(err)
			return false
		}
		if nodeIpAddress != ip {
			return false
		}
	}
	return true
}

// isValidRemote: Check if the ip is in membership list from gossip
func isValidRemote(ip string) bool {
	membershipList := exportgossip.GetMembershipList()
	_, ok := membershipList[ip]
	return ok
}

// findSuitableComputeNode: Returns a suitable compute node if available
// based on the task request, else it returns empty string
func findSuitableComputeNode(taskRequest exportapi.TaskRequest) string {
	membershipList := exportgossip.GetMembershipList()
	maxScore := -1.0
	suitableNode := ""
	for ip, node := range membershipList {
		cores := float64(node.NodeProfile.Cores)
		// Adding a small delta to avoid potential divide by 0 error
		load := float64(node.NodeProfile.Load) + 0.0000000001
		memory := float64(node.NodeProfile.Ram)
		reputation := float64(node.NodeProfile.Reputation)

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
	return suitableNode
}

// updateNodeReputation: Update the reputation
func updateNodeReputation() {
	nodeIpAddress, err := util.GetIpAddress()
	if err != nil {
		log.Error(err)
	} else {
		membershipList := exportgossip.GetMembershipList()
		node := membershipList[nodeIpAddress]
		node.NodeProfile.Reputation++
		membershipList[nodeIpAddress] = node
		log.Info("Node reputation updated to", node.NodeProfile.Reputation)
	}
}

// completeTask: Routine to run async for task completion
func completeTask(taskId string) {
	task := exportapi.Tasks[taskId]
	task.Status = exportapi.Running
	exportapi.Tasks[taskId] = task

	taskResult, taskErr := functions.Map[task.Function](task.Data, task.Abort)
	task.Status = exportapi.Stopped
	task.Result = taskResult

	task.Metrics.EndTime = time.Now().Format("15:04:05.000000")

	updateNodeReputation()

	MetricsLog(task.Metrics)

	if taskErr != nil {
		task.Error = taskErr.Error()
	} else {
		task.Error = ""
	}

	exportapi.Tasks[taskId] = task
}

// MetricsLog: Log metrics
func MetricsLog(metricsIn exportapi.TaskMetrics) {
	metrics, _ := json.Marshal(metricsIn)
	log.Output("METRIC ", 3, string(metrics))
}
