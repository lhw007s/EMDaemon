package main

const BRIDGE_SERVER_PROPERTIES_SRC = "../cfg/bridge_server.properties"

type INetwork interface {
	Load()
	PacketRecvThread()
	PacketSendThread()
	Start()
	Stop()
	UdpSend()
}

type NetworkManager struct {
	BridgeServer  map[string]map[string]string
	MonitorServer map[string]map[string]string
}

func NewNetworkManager() *NetworkManager {
	return &NetworkManager{
		make(map[string]map[string]string),
		make(map[string]map[string]string),
	}
}

//기본 설정파일 로드한다
func (n *NetworkManager) Load() {
	//gGlobalScope.StringTool.loadProperties(BRIDGE_SERVER_PROPERTIES_SRC, &n.BridgeServer)
}

func (n *NetworkManager) Init() {

}

//패킷 수신 고루틴
//EM웹서버에서 UDP로 발송한다.
func (n *NetworkManager) PacketRecvThread() {

}

//패킷 발신 고루틴
//브릿지, 모니터 서버로 UDP발송한다.
func (n *NetworkManager) PacketSendThread() {

}

//Os 시작 시그널 처리
func (n *NetworkManager) Start() {

}

//OS 종료 시그널 처리
func (n *NetworkManager) Stop() {
}

//브릿지, 모니터 서버로 패킷 발송
func (n *NetworkManager) UdpSend() {

}
