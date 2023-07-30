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
	"time"
)

type Response struct {
	Status string              `json:"status"`
	Path   map[string][]string `json:"path"`
}

type RequestData struct {
	Index      int `json:"index"`
	SatSize    int `json:"sat_size"`
	GroundSize int `json:"ground_size"`
	Time       int `json:"time"`
}

type RouteTable struct {
	// table contains the all paths from current node to target node
	table map[string][]string

	// url 是用来更新路由表的url
	url string

	// index 表示当前的node是NodeList中的第几号节点
	index    int
	hostname string
	time     time.Time
}

const (
	TableSize = 10
	//urlSuffix  = ":8000/bent-pipe"
	//urlSuffix  = ":8000/bent-pipe"
	urlSuffix  = ":8000/sat-route-update"
	urlPrefix  = "http://"
	SatSize    = 8
	GroundSize = 1
)

// NodeList 这里采用了hard-coding的方案，默认此时只有一个地球基站
// 如果时间充足会考虑部署多个地球基站
var (
	NodeList = []string{"edge-0", "edge-1", "edge-2", "edge-3", "edge-4", "edge-5", "edge-6", "edge-7", "master"}
)

// todo: 这里不应该采用硬编码的方式进行编码，这样做首先无法做到可以添加新的节点，同时也不够优雅
// the direct use of hard-coding is employed to initialize the information
// of cluster,however, it should be replaced by the information from
// the k8s cluster or from external service
func (r *RouteTable) initTable(hostname string, hostURL string) {
	klog.Info("[Router::initTable]: route table is starting to init")

	r.index = 0
	// 初始化路由表
	for _, key := range NodeList {
		r.table[key] = []string{}
	}
	r.table[hostname] = []string{hostname}

	// 初始化 index
	r.hostname = hostname
	for i, key := range NodeList {
		if key == hostname {
			r.index = i
			klog.Info("[Router::initTable]: the index is: ", r.index)
		}
	}

	// init url
	//hostURL, _ := getHostIP(NodeHost[0].String())
	r.url = urlPrefix + hostURL + urlSuffix
	klog.Infof("[Router]: the url is: %s", r.url)

	// 首先就要更新一次
	// todo: 这里只更新了一次，后续需要让其不断的更新
	go func() {
		for {
			err := r.UpdateTable()
			if err != nil {
				klog.Info("[router::NewRouteTable]: failed to update table")
			}
			time.Sleep(30 * time.Second)
		}
	}()
}

// UpdateTable 用来发起请求，然后更新路由表
func (r *RouteTable) UpdateTable() error {
	req := RequestData{
		Index:      r.index,
		SatSize:    SatSize,
		GroundSize: GroundSize,
		Time:       int(time.Now().Sub(r.time).Seconds()),
	}
	body, err := r.SendPostRequest(req)
	if err != nil {
		klog.Info("[router::UpdateTable]: failed to send post request")
		return err
	}

	// clear 掉当前的路由表
	// 初始化路由表
	for _, key := range NodeList {
		r.table[key] = []string{}
	}
	r.table[r.hostname] = []string{r.hostname}

	// fmt.Println("response Body:", string(body))
	err = json.Unmarshal(body, &r.table)
	if err != nil {
		return err
	}
	klog.Infof("[router::UpdateTable]: the route table is: %+v", r.table)
	return nil
}

// query 用来提供查询
// 现在时间紧急，来不及实现多路的版本了，首先就提供一个最初的就行了
func (r *RouteTable) Query(dst string) ([]string, error) {
	path, ok := r.table[dst]
	if !ok {
		klog.Info("[router::query]: the dst is not in the table")
		return nil, errors.New("the dst is not in the table")
	}

	if len(path) == 0 {
		klog.Info("[router::query]: there is no route in the table")
		return nil, errors.New("No route")
	}

	return path, nil
}

func (r *RouteTable) SendPostRequest(reqData RequestData) ([]byte, error) {
	jsonValue, _ := json.Marshal(reqData)
	response, err := http.Post(r.url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 将初始化都放在NewRouteTable里面
func NewRouteTable() *RouteTable {

	r := &RouteTable{
		table: make(map[string][]string, TableSize),
	}

	r.time = time.Now()

	// 初始化hostname
	hostname, err := os.Hostname()
	if err != nil {
		klog.Info("[router::NewRouteTable]: failed to get hostname")
		return nil
	}
	// 初始化hostURL
	//hostIP, err := getHostIP(NodeHost[0].String())

	// todo: the hostIP is currently hard-coding
	hostIP := "192.168.1.29"
	r.initTable(hostname, hostIP)
	return r
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
