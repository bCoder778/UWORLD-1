package config

type RpcConfig struct {
	DataDir    string
	RpcIp      string
	RpcPort    string
	HttpPort   string
	RpcTLS     bool
	RpcCert    string
	RpcCertKey string
	RpcPass    string
}
