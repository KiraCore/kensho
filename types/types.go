package types

const (
	APP_NAME string = "Kensho"

	BOOTSTRAP_SCRIPT       string = "https://raw.githubusercontent.com/KiraCore/sekin/main/scripts/bootstrap.sh"
	SEKIN_EXECUTE_ENDPOINT string = "http://localhost:8282/api/execute"
	SEKIN_STATUS_ENDPOINT  string = "http://localhost:8282/api/status"

	DEFAULT_INTERX_PORT int = 11000
	DEFAULT_P2P_PORT    int = 26656
	DEFAULT_RPC_PORT    int = 26657
	DEFAULT_GRPC_PORT   int = 9090
	DEFAULT_SHIDAI_PORT int = 8282
)

type RequestDeployPayload struct {
	Command string `json:"command"`
	Args    Args   `json:"args"`
}

// Args represents the arguments in the JSON payload.
type Args struct {
	IP         string `json:"ip"`
	InterxPort int    `json:"interx_port"`
	RPCPort    int    `json:"rpc_port"`
	P2PPort    int    `json:"p2p_port"`
	Mnemonic   string `json:"mnemonic"`
	Local      bool   `json:"local"`
}

type Cmd string

const (
	Activate           Cmd = "activate"
	Pause              Cmd = "pause"
	Unpause            Cmd = "unpause"
	ClaimValidatorSeat Cmd = "claim_seat"
)

type RequestTXPayload struct {
	Command string                       `json:"command"`
	Args    ExecSekaiMaintenanceCommands `json:"args"`
}
type ExecSekaiMaintenanceCommands struct {
	TX      Cmd    `json:"tx"` //pause, unpause, activate,
	Moniker string `json:"moniker"`
}

type ExecSekaiCommands struct {
	Command  string   `json:"command"`
	ExecArgs ExecArgs `json:"args"`
}

type ExecArgs struct {
	Exec []string `json:"exec"`
}
