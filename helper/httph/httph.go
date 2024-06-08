package httph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	interxendpoint "github.com/KiraCore/kensho/types/endpoint/interx"
	sekaiendpoint "github.com/KiraCore/kensho/types/endpoint/sekai"
	shidaiendpoint "github.com/KiraCore/kensho/types/endpoint/shidai"
	"golang.org/x/crypto/ssh"
)

func MakeHttpRequest(url, method string) ([]byte, error) {
	client := http.DefaultClient
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetInterxStatus(nodeIP, interxPort string) (*interxendpoint.Status, error) {
	url := fmt.Sprintf("http://%v:%v/api/status", nodeIP, interxPort)
	b, err := MakeHttpRequest(url, "GET")
	if err != nil {
		return nil, err
	}
	var info *interxendpoint.Status
	err = json.Unmarshal(b, &info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func GetSekaiStatus(nodeIP, port string) (*sekaiendpoint.Status, error) {
	url := fmt.Sprintf("http://%v:%v/status", nodeIP, port)
	b, err := MakeHttpRequest(url, "GET")
	if err != nil {
		return nil, err
	}
	var info *sekaiendpoint.Status
	err = json.Unmarshal(b, &info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func GetShidaiStatus(sshClient *ssh.Client, shidaiPort int) (shidaiendpoint.Status, error) {
	valid := ValidatePortRange(strconv.Itoa(shidaiPort))
	if !valid {
		return shidaiendpoint.Status{}, fmt.Errorf("<%v> is not valid", shidaiPort)
	}
	o, err := ExecHttpRequestBySSHTunnel(sshClient, fmt.Sprintf("http://localhost:%v/status", shidaiPort), "GET", nil)
	if err != nil {
		return shidaiendpoint.Status{}, err
	}
	var data shidaiendpoint.Status
	err = json.Unmarshal(o, &data)
	if err != nil {
		return shidaiendpoint.Status{}, err
	}
	return data, err
}

func GetSekaiABCI_Info(nodeIP, port string) (*sekaiendpoint.ABCI_Info, error) {
	url := fmt.Sprintf("http://%v:%v/abci_info", nodeIP, port)
	b, err := MakeHttpRequest(url, "GET")
	if err != nil {
		return nil, err
	}
	var data sekaiendpoint.ABCI_Info
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func GetBinariesVersionsFromTrustedNode(trustedIP, sekaiRPC_Port, interxPort string) (sekaiVersion, interxVersion string, err error) {
	log.Printf("Getting sekai and interx version from <%v>", trustedIP)
	abci, err := GetSekaiABCI_Info(trustedIP, sekaiRPC_Port)
	if err != nil {
		return "", "", fmt.Errorf("error getting abci info from sekai: %w", err)
	}
	sekaiVersion = abci.ABCI_result.Response.Version

	interxStatus, err := GetInterxStatus(trustedIP, interxPort)
	if err != nil {
		return "", "", fmt.Errorf("error getting interx status: %w", err)
	}
	interxVersion = interxStatus.InterxInfo.Version
	log.Printf("Sekai version: <%v>  Interx version: <%v>", sekaiVersion, interxVersion)
	return sekaiVersion, interxVersion, nil
}

func GetValidatorStatus(sshClient *ssh.Client, shidaiPort int) (*shidaiendpoint.Validator, error) {
	valid := ValidatePortRange(strconv.Itoa(shidaiPort))
	if !valid {
		return nil, fmt.Errorf("<%v> is not valid", shidaiPort)
	}
	o, err := ExecHttpRequestBySSHTunnel(sshClient, fmt.Sprintf("http://localhost:%v/validator", shidaiPort), "GET", nil)
	if err != nil {
		return nil, err
	}
	var data shidaiendpoint.Validator
	err = json.Unmarshal(o, &data)
	if err != nil {
		return nil, err
	}
	return &data, err
}

func ValidatePortRange(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false // Not an integer
	}
	if port < 1 || port > 65535 {
		return false // Out of valid port range
	}
	return true
}

func ValidateIP(input string) bool {
	ipCheck := net.ParseIP(input)
	return ipCheck != nil
}

func ExecHttpRequestBySSHTunnel(sshClient *ssh.Client, address, method string, payload []byte) ([]byte, error) {
	log.Printf("requesting <%v>\nPayload: %+v", address, payload)
	dialer := func(network, addr string) (net.Conn, error) {
		conn, err := sshClient.Dial(network, addr)
		if err != nil {
			log.Printf("Failed to establish SSH tunnel: %v", err)
		}
		return conn, err
	}

	httpTransport := &http.Transport{
		Dial: dialer,
	}

	httpClient := &http.Client{
		Transport: httpTransport,
	}

	var req *http.Request
	var err error

	if len(payload) == 0 {
		req, err = http.NewRequest(method, address, nil)
	} else {
		req, err = http.NewRequest(method, address, bytes.NewBuffer(payload))
	}
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to send HTTP request: %v", err)
		return nil, err
	}

	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func CreateTunnelForSSEConnection(sshClient *ssh.Client, address string) (*http.Response, error) {
	dialer := func(network, addr string) (net.Conn, error) {
		return sshClient.Dial(network, addr)
	}

	httpTransport := &http.Transport{
		Dial: dialer,
	}

	httpClient := &http.Client{
		Transport: httpTransport,
	}

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request: %v", err)
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to send HTTP request: %v", err)
		return nil, err

	}

	return resp, nil
}

// TODO:
type Dashboard struct {
	RoleIDs []string `json:"role_ids"`

	Date                string `json:"date"`
	ValidatorStatus     string `json:"val_status"`
	Blocks              string `json:"blocks"`
	Top                 string `json:"top"`
	Streak              string `json:"streak"`
	Mischance           string `json:"mischance"`
	MischanceConfidence string `json:"mischance_confidence"`
	StartHeight         string `json:"start_height"`
	LastProducedBlock   string `json:"last_present_block"`
	ProducedBlocks      string `json:"produced_blocks_counter"`
	Moniker             string `json:"moniker"`
	ValidatorAddress    string `json:"address"`
	ChainID             string `json:"chain_id"`
	NodeID              string `json:"node_id"`
	GenesisChecksum     string `json:"genesis_checksum"`

	ActiveValidators   int `json:"active_validators"`
	PausedValidators   int `json:"paused_validators"`
	InactiveValidators int `json:"inactive_validators"`
	JailedValidators   int `json:"jailed_validatore"`
	WaitingValidators  int `json:"waiting_validators"`

	SeatClaimAvailable bool `json:"seat_claim_available"`
	Waiting            bool `json:"seat_claim_pending"`
	CatchingUp         bool `json:"catching_up"`
}

func GetDashboardInfo(sshClient *ssh.Client, shidaiPort int) (*Dashboard, error) {
	url := fmt.Sprintf("http://localhost:%v/dashboard", shidaiPort)
	o, err := ExecHttpRequestBySSHTunnel(sshClient, url, "GET", nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR getting request from <%v>, reason: %w", url, err)
	}
	var data *Dashboard
	err = json.Unmarshal(o, &data)
	if err != nil {
		return nil, fmt.Errorf("ERROR when unmarshaling <%v>\nReason: %w", string(o), err)
	}
	return data, nil
}
