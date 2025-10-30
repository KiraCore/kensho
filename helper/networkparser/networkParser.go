package networkparser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	interxendpoint "github.com/KiraCore/kensho/types/endpoint/interx"
)

var mu sync.Mutex

type Node struct {
	IP      string
	ID      string
	Peers   []Node
	NCPeers int
}

type BlacklistedNode struct {
	IP    string
	Error []error
}

// get nodes that are available by the specified port
func GetAllNodesV3(ctx context.Context, firstNode string, port int, depth int, ignoreDepth bool) (map[string]Node, map[string]BlacklistedNode, error) {
	nodesPool := make(map[string]Node)
	blacklist := make(map[string]BlacklistedNode)
	processed := make(map[string]string)
	client := http.DefaultClient
	node, err := GetNetInfoFromInterx(ctx, client, firstNode, port)
	if err != nil {
		return nil, nil, err
	}

	var wg sync.WaitGroup
	for _, n := range node.Peers {
		wg.Add(1)
		go loopFunc(ctx, &wg, client, nodesPool, blacklist, processed, n.RemoteIP, port, 0, depth, ignoreDepth)
	}

	wg.Wait()
	fmt.Println()
	log.Printf("\nTotal saved peers:%v\nOriginal node peer count: %v\nBlacklisted nodes(not reachable): %v\n", len(nodesPool), len(node.Peers), len(blacklist))
	// log.Printf("BlackListed: %+v ", blacklist)

	return nodesPool, blacklist, nil
}

func loopFunc(ctx context.Context, wg *sync.WaitGroup, client *http.Client, pool map[string]Node, blacklist map[string]BlacklistedNode, processed map[string]string, ip string, port int, currentDepth, totalDepth int, ignoreDepth bool) {

	defer wg.Done()
	if !ignoreDepth {
		if currentDepth >= totalDepth {
			// log.Printf("DEPTH LIMIT REACHED")
			return
		}
	}

	// log.Printf("Current depth: %v, IP: %v", currentDepth, ip)

	mu.Lock()
	if _, exist := blacklist[ip]; exist {
		mu.Unlock()
		// log.Printf("BLACKLISTED: %v", ip)
		return
	}
	if _, exist := pool[ip]; exist {
		mu.Unlock()
		// log.Printf("ALREADY EXIST: %v", ip)
		return
	}
	if _, exist := processed[ip]; exist {
		mu.Unlock()
		// log.Printf("ALREADY PROCESSED: %v", ip)
		return
	} else {
		processed[ip] = ip
	}
	mu.Unlock()

	currentDepth++

	var nodeInfo *interxendpoint.NetInfo
	var status *interxendpoint.Status
	var errNetInfo error
	var errStatus error

	//local wait group
	var localWaitGroup sync.WaitGroup
	localWaitGroup.Add(2)
	go func() {
		defer localWaitGroup.Done()
		nodeInfo, errNetInfo = GetNetInfoFromInterx(ctx, client, ip, port)
	}()
	go func() {
		defer localWaitGroup.Done()
		status, errStatus = GetStatusFromInterx(ctx, client, ip, port)
	}()
	localWaitGroup.Wait()
	// nodeInfo, errNetInfo = GetNetInfoFromInterx(ctx, client, ip)
	// status, errStatus = GetStatusFromInterx(ctx, client, ip)

	if errNetInfo != nil || errStatus != nil {
		// log.Printf("%v", err.Error())
		mu.Lock()
		log.Printf("adding <%v> to blacklist", ip)
		blacklist[ip] = BlacklistedNode{IP: ip, Error: []error{errNetInfo, errStatus}}
		cleanValue(processed, ip)
		mu.Unlock()
		// defer localWaitGroup.Done()
		return
	}

	mu.Lock()
	log.Printf("adding <%v> to the pool, nPeers: %v", ip, nodeInfo.NPeers)

	node := Node{
		IP:      ip,
		NCPeers: nodeInfo.NPeers,
		ID:      status.NodeInfo.ID,
	}

	for _, nn := range nodeInfo.Peers {
		ip, port, err := extractIP(nn.NodeInfo.ListenAddr)
		if err != nil {
			continue
		}
		node.Peers = append(node.Peers, Node{IP: fmt.Sprintf("%v:%v", ip, port), ID: nn.NodeInfo.ID})
	}

	pool[ip] = node
	cleanValue(processed, ip)
	mu.Unlock()

	for _, p := range nodeInfo.Peers {
		wg.Add(1)
		go loopFunc(ctx, wg, client, pool, blacklist, processed, p.RemoteIP, port, currentDepth, totalDepth, ignoreDepth)

		listenAddr, _, err := extractIP(p.NodeInfo.ListenAddr)
		if err != nil {
			continue
		} else {
			if listenAddr != p.RemoteIP {
				log.Printf("listen addr (%v) and remoteIp (%v) are not the same, creating new goroutine for listen addr", listenAddr, p.RemoteIP)
				wg.Add(1)
				go loopFunc(ctx, wg, client, pool, blacklist, processed, listenAddr, port, currentDepth, totalDepth, ignoreDepth)
			}
		}

	}

}

func cleanValue(toClean map[string]string, key string) {
	delete(toClean, key)
}

const TimeOutDelay time.Duration = time.Second * 5

func GetNetInfoFromInterx(ctx context.Context, client *http.Client, ip string, port int) (*interxendpoint.NetInfo, error) {

	ctxWithTO, c := context.WithTimeout(ctx, TimeOutDelay)
	defer c()
	// log.Printf("Getting net_info from: %v", ip)
	url := fmt.Sprintf("http://%v:%d/api/net_info", ip, port)
	req, err := http.NewRequestWithContext(ctxWithTO, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeInfo interxendpoint.NetInfo
	err = json.Unmarshal(b, &nodeInfo)
	if err != nil {
		return nil, err
	}
	return &nodeInfo, nil
}

func GetStatusFromInterx(ctx context.Context, client *http.Client, ip string, port int) (*interxendpoint.Status, error) {
	ctxWithTO, c := context.WithTimeout(ctx, TimeOutDelay)
	defer c()
	// log.Printf("Getting net_info from: %v", ip)
	url := fmt.Sprintf("http://%v:%d/api/status", ip, port)
	req, err := http.NewRequestWithContext(ctxWithTO, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeStatus interxendpoint.Status
	err = json.Unmarshal(b, &nodeStatus)
	if err != nil {
		return nil, err
	}
	return &nodeStatus, nil
}

func extractIP(input string) (ip string, port string, err error) {
	// Regular expression to match IP addresses
	re := regexp.MustCompile(`tcp://([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+):([0-9]+)`)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("no IP address or port found in input")
	}
	return matches[1], matches[2], nil
}
