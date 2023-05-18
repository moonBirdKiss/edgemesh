package tunnel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Response struct {
	Status string              `json:"status"`
	Path   map[string][]string `json:"path"`
}

type RequestData struct {
	Host string `json:"hostname"`
}

type RouteTable struct {
	// table contains the path from current node to target node
	table map[string]string
}

const (
	TableSize  = 10
	url_suffix = ":8000/sat-route-update"
	url_prefix = "http://"
)

var (
	url string
)

// todo: 这里不应该采用硬编码的方式进行编码，这样做首先无法做到可以添加新的节点，同时也不够优雅
// the direct use of hardcoding is employed to initialize the information
// of cluster,however, it should be replaced by the information from
// the k8s cluster or from external service
func (r *RouteTable) initTable() {
	hostname, err := os.Hostname()
	if err != nil {
		return
	}
	klog.Infof("[Router]: the hostname is: %s", hostname)
	klog.Infof("[Router]: the NodeHostIP is: %+v", NodeHost)

	klog.Info("[Router]: route table is starting to init")
	r.table["edge-0"] = hostname + "," + "edge-0"
	r.table["edge-1"] = hostname + "," + "edge-1"
	r.table["edge-2"] = hostname + "," + "edge-2"
	r.table["edge-3"] = hostname + "," + "edge-3"

	// current, we only use these nodes
	r.table["edge-4"] = hostname + "," + "edge-4"
	r.table["edge-5"] = hostname + "," + "edge-5"
	r.table["edge-6"] = hostname + "," + "edge-6"

	r.table["master"] = hostname + "," + "master"

	// init the host path
	r.table[hostname] = hostname

	// init url
	hostURL, err := getHostIP(NodeHost[0].String())
	url = url_prefix + hostURL + url_suffix
	klog.Infof("[Router]: the url is: %s", url)
}

// updateTable is used to get external info to update the table filed
func (r *RouteTable) updateTable() error {
	res, err := sendPostRequest()
	if err != nil {
		klog.Error("[updateTable]: fail to query the table", err)
		return err
	}

	path := res.Path

	for key, value := range path {
		r.table[key] = strings.Join(value, ",")
	}

	// only display when debugging
	klog.Infof("[updateTable]: the route table is: %v", r.table)
	return nil
}

func (r *RouteTable) query(dst string) ([]string, error) {
	path, ok := r.table[dst]
	if !ok {
		klog.Info("[router]: dst ID is not in the route table", dst)
		return nil, errors.New("dst ID is not in the route table")
	}
	ret := strings.Split(path, ",")
	klog.Info("[router]: the route path is: ", ret)
	return ret, nil
}

func sendPostRequest() (*Response, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	data := RequestData{
		Host: hostname,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		table: make(map[string]string, TableSize),
	}
}

func getHostIP(input string) (string, error) {
	re := regexp.MustCompile(`/ip4/(\d+\.\d+\.\d+\.\d+)/tcp/\d+`)
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		ip := match[1]
		fmt.Println(ip)
		return ip, nil
	} else {
		fmt.Println("No match found")
		return "", errors.New("no match found")
	}
}
