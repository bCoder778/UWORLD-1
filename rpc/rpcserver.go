package rpc

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/common/utils"
	"github.com/uworldao/UWORLD/config"
	"github.com/uworldao/UWORLD/consensus"
	"github.com/uworldao/UWORLD/core/interface"
	"github.com/uworldao/UWORLD/core/runner"
	coreTypes "github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/crypto/certgen"
	log "github.com/uworldao/UWORLD/log/log15"
	"github.com/uworldao/UWORLD/p2p"
	"github.com/uworldao/UWORLD/rpc/rpctypes"
	"github.com/uworldao/UWORLD/services/reqmgr"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	config        *config.RpcConfig
	txPool        _interface.ITxPool
	accountState  _interface.IAccountState
	contractState _interface.IContractState
	runner        *runner.ContractRunner
	consensus     consensus.IConsensus
	chain         _interface.IBlockChain
	grpcServer    *grpc.Server
	httpServer    *http.Server
	peerManager   p2p.IPeerManager
	peers         reqmgr.Peers
}

func NewServer(config *config.RpcConfig, txPool _interface.ITxPool, state _interface.IAccountState, contractState _interface.IContractState,
	runner *runner.ContractRunner, consensus consensus.IConsensus, chain _interface.IBlockChain, peerManager p2p.IPeerManager,
	peers reqmgr.Peers) *Server {
	return &Server{config: config, txPool: txPool, accountState: state, contractState: contractState,
		consensus: consensus, chain: chain, peerManager: peerManager, peers: peers, runner: runner}
}

func (rs *Server) Start() error {
	var err error
	var tlsConfig *tls.Config
	endPoint := "127.0.0.1:" + rs.config.RpcPort
	lis, err := net.Listen("tcp", ":"+rs.config.RpcPort)
	if err != nil {
		return err
	}
	httpListen, err := net.Listen("tcp", ":"+rs.config.HttpPort)
	if err != nil {
		return err
	}
	rs.grpcServer, err = rs.NewServer()
	if err != nil {
		return err
	}

	RegisterGreeterServer(rs.grpcServer, rs)
	reflection.Register(rs.grpcServer)
	go func() {
		if err := rs.grpcServer.Serve(lis); err != nil {
			log.Error("Rpc startup failed!", "err", err)
			os.Exit(1)
			return
		}
	}()

	getWay, err := rs.NewGetWay(endPoint)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", getWay)
	if rs.config.RpcTLS {
		tlsConfig, err = getTLSConfig(rs.config.RpcCert, rs.config.RpcCertKey)
		if err != nil {
			return err
		}
		httpListen = tls.NewListener(httpListen, tlsConfig)
	}

	rs.httpServer = &http.Server{
		Addr:      ":" + rs.config.HttpPort,
		Handler:   grpcHandlerFunc(rs.grpcServer, mux, rs.config),
		TLSConfig: tlsConfig,
	}
	go func() {
		log.Info("HTTP API startup", "port", rs.config.HttpPort)
		if err := rs.httpServer.Serve(httpListen); err != nil {
			log.Error("Rpc startup!", "err", err)
			os.Exit(1)
			return
		}
	}()

	if rs.config.RpcTLS {
		log.Info("Rpc startup", "port", rs.config.RpcPort, "pem", rs.config.RpcCert)
	} else {
		log.Info("Rpc startup", "port", rs.config.RpcPort)
	}
	return nil
}

func getTLSConfig(certPemPath, certKeyPath string) (*tls.Config, error) {
	var certKeyPair *tls.Certificate
	cert, _ := ioutil.ReadFile(certPemPath)
	key, _ := ioutil.ReadFile(certKeyPath)

	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	certKeyPair = &pair

	return &tls.Config{
		Certificates: []tls.Certificate{*certKeyPair},
		NextProtos:   []string{http2.NextProtoTLS},
	}, nil
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler, rpcConfig *config.RpcConfig) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pass, _ := r.BasicAuth()
		if pass != rpcConfig.RpcPass {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf("the token authentication information is invalid: password=%s\n", pass)))
			return
		}
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func (rs *Server) Close() {
	rs.grpcServer.Stop()
	log.Info("GRPC server closed")
}

func (rs *Server) NewServer() (*grpc.Server, error) {
	var opts []grpc.ServerOption
	var interceptor grpc.UnaryServerInterceptor
	interceptor = rs.interceptor
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	// If tls is configured, generate tls certificate
	if rs.config.RpcTLS {
		if err := rs.generateCertFile(); err != nil {
			return nil, err
		}
		transportCredentials, err := credentials.NewServerTLSFromFile(rs.config.RpcCert, rs.config.RpcCertKey)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(transportCredentials))

	}

	// Set the maximum number of bytes received and sent
	opts = append(opts, grpc.MaxRecvMsgSize(reqmgr.MaxRequestBytes))
	opts = append(opts, grpc.MaxSendMsgSize(reqmgr.MaxRequestBytes))
	return grpc.NewServer(opts...), nil
}

func (rs *Server) NewGetWay(endPoint string) (*runtime.ServeMux, error) {
	dopts := []grpc.DialOption{}
	dopts = append(dopts, grpc.WithPerRPCCredentials(&customCredential{
		Password: rs.config.RpcPass,
		OpenTLS:  rs.config.RpcTLS,
	}))
	ctx := context.Background()
	if rs.config.RpcTLS {
		creds, err := credentials.NewClientTLSFromFile(rs.config.RpcCert, "")
		if err != nil {
			return nil, err
		}
		dopts = append(dopts, grpc.WithTransportCredentials(creds))
	} else {
		dopts = append(dopts, grpc.WithInsecure())
	}

	gwmux := runtime.NewServeMux()
	if err := RegisterGreeterHandlerFromEndpoint(ctx, gwmux, endPoint, dopts); err != nil {
		return nil, err
	}
	gwmux.GetForwardResponseOptions()
	return gwmux, nil
}

func (rs *Server) SendTransaction(_ context.Context, req *Bytes) (*Response, error) {
	var rpcTx *coreTypes.RpcTransaction
	if err := json.Unmarshal(req.Bytes, &rpcTx); err != nil {
		return NewResponse(rpctypes.RpcErrParam, nil, err.Error()), nil
	}
	tx, err := coreTypes.TranslateRpcTxToTx(rpcTx)
	if err != nil {
		return NewResponse(rpctypes.RpcErrParam, nil, ""), nil
	}
	if err := rs.txPool.Add(tx, false); err != nil {
		return NewResponse(rpctypes.RpcErrTxPool, nil, err.Error()), nil
	}
	return NewResponse(rpctypes.RpcSuccess, []byte(fmt.Sprintf("send transaction %s success", tx.Hash().String())), ""), nil
}

func (rs *Server) GetAccount(_ context.Context, req *Address) (*Response, error) {
	addr := hasharry.StringToAddress(req.Address)
	account := rs.accountState.GetAccountState(addr)
	rpcAccount := rpctypes.TranslateAccountToRpcAccount(account.(*coreTypes.Account))
	bytes, err := json.Marshal(rpcAccount)
	if err != nil {
		return NewResponse(rpctypes.RpcErrMarshal, nil, fmt.Sprintf("%s address not exsit", req.Address)), nil
	}
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil
}

func (rs *Server) GetTransaction(ctx context.Context, req *Hash) (*Response, error) {
	hash, err := hasharry.StringToHash(req.Hash)
	if err != nil {
		return NewResponse(rpctypes.RpcErrParam, nil, "hash error"), nil
	}
	tx, err := rs.chain.GetTransaction(hash)
	if err != nil {
		return NewResponse(rpctypes.RpcErrBlockChain, nil, err.Error()), nil
	}
	confirmed := rs.chain.GetConfirmedHeight()
	index, err := rs.chain.GetTransactionIndex(hash)
	if err != nil {
		return NewResponse(rpctypes.RpcErrBlockChain, nil, fmt.Sprintf("%s is not exist", hash.String())), nil
	}
	height := index.GetHeight()
	var rpcTx *coreTypes.RpcTransaction
	state, _ := rs.chain.GetContractState(hash)
	if state != nil {
		rpcTx, _ = coreTypes.TranslateContractV2TxToRpcTx(tx.(*coreTypes.Transaction), state)
	} else {
		rpcTx, _ = coreTypes.TranslateTxToRpcTx(tx.(*coreTypes.Transaction))
	}
	rsMsg := &coreTypes.RpcTransactionConfirmed{
		TxHead:    rpcTx.TxHead,
		TxBody:    rpcTx.TxBody,
		Height:    height,
		Confirmed: confirmed >= height,
	}
	bytes, _ := json.Marshal(rsMsg)
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil
}

func (rs *Server) GetBlockByHash(ctx context.Context, req *Hash) (*Response, error) {
	hash, err := hasharry.StringToHash(req.Hash)
	if err != nil {
		return NewResponse(rpctypes.RpcErrParam, nil, "hash error"), nil
	}
	block, err := rs.chain.GetBlockByHash(hash)
	if err != nil {
		return NewResponse(rpctypes.RpcErrBlockChain, nil, err.Error()), nil
	}
	rpcBlock, _ := coreTypes.TranslateBlockToRpcBlock(block, rs.chain.GetConfirmedHeight(), rs.chain.GetContractState)
	bytes, err := json.Marshal(rpcBlock)
	if err != nil {
		return NewResponse(rpctypes.RpcErrMarshal, nil, err.Error()), nil
	}
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil

}

func (rs *Server) GetBlockByHeight(_ context.Context, req *Height) (*Response, error) {
	block, err := rs.chain.GetBlockByHeight(req.Height)
	if err != nil {
		return NewResponse(rpctypes.RpcErrBlockChain, nil, err.Error()), nil
	}
	rpcBlock, _ := coreTypes.TranslateBlockToRpcBlock(block, rs.chain.GetConfirmedHeight(), rs.chain.GetContractState)
	bytes, err := json.Marshal(rpcBlock)
	if err != nil {
		return NewResponse(rpctypes.RpcErrMarshal, nil, err.Error()), nil
	}
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil
}

func (rs *Server) GetPoolTxs(context.Context, *Null) (*Response, error) {
	preparedTxs, futureTxs := rs.txPool.GetAll()
	txPoolTxs, _ := coreTypes.TranslateTxsToRpcTxPool(preparedTxs, futureTxs)
	bytes, err := json.Marshal(txPoolTxs)
	if err != nil {
		return NewResponse(rpctypes.RpcErrMarshal, nil, err.Error()), nil
	}
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil
}

func (rs *Server) GetCandidates(context.Context, *Null) (*Response, error) {
	candidates := rs.consensus.GetCandidates(rs.chain)
	if candidates == nil || len(candidates) == 0 {
		return NewResponse(rpctypes.RpcErrDPos, nil, "no candidates"), nil
	}
	bytes, err := json.Marshal(coreTypes.TranslateCandidatesToRpcCandidates(candidates))
	if err != nil {
		return NewResponse(rpctypes.RpcErrMarshal, nil, err.Error()), nil
	}
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil
}

func (rs *Server) GetLastHeight(context.Context, *Null) (*Response, error) {
	height := rs.chain.GetLastHeight()
	sHeight := strconv.FormatUint(height, 10)
	return NewResponse(rpctypes.RpcSuccess, []byte(sHeight), ""), nil
}

func (rs *Server) GetContract(ctx context.Context, req *Address) (*Response, error) {
	contract := rs.contractState.GetContract(req.Address)
	if contract == nil {
		return NewResponse(rpctypes.RpcErrContract, nil, fmt.Sprintf("contract address %s is not exist", req.Address)), nil
	}
	bytes, err := json.Marshal(coreTypes.TranslateContractToRpcContract(contract))
	if err != nil {
		return NewResponse(rpctypes.RpcErrMarshal, nil, err.Error()), nil
	}
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil
}

func (rs *Server) GetConfirmedHeight(context.Context, *Null) (*Response, error) {
	height := rs.chain.GetConfirmedHeight()
	sHeight := strconv.FormatUint(height, 10)
	return NewResponse(rpctypes.RpcSuccess, []byte(sHeight), ""), nil
}

func (rs *Server) Peers(context.Context, *Null) (*Response, error) {
	peers := rs.peers.PeersInfo()
	peersJson, _ := json.Marshal(peers)
	return NewResponse(rpctypes.RpcSuccess, peersJson, ""), nil
}

func (rs *Server) NodeInfo(context.Context, *Null) (*Response, error) {
	node := rs.peers.NodeInfo()
	nodeJson, _ := json.Marshal(node)
	return NewResponse(rpctypes.RpcSuccess, nodeJson, ""), nil
}

func (rs *Server) GetExchangePairs(ctx context.Context, req *Address) (*Response, error) {
	pairs, err := rs.runner.ExchangePair(hasharry.StringToAddress(req.Address))
	if err != nil {
		return NewResponse(rpctypes.RpcErrContract, nil, err.Error()), nil
	}
	bytes, _ := json.Marshal(pairs)
	return NewResponse(rpctypes.RpcSuccess, bytes, ""), nil
}

func NewResponse(code int32, result []byte, err string) *Response {
	return &Response{Code: code, Result: result, Err: err}
}

// Authenticate rpc users
func (rs *Server) auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("no token authentication information")
	}
	var (
		password string
	)

	if val, ok := md["password"]; ok {
		password = val[0]
	}

	if password != rs.config.RpcPass {
		return fmt.Errorf("the token authentication information is invalid:password=%s", password)
	}
	return nil
}

func (rs *Server) interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	err = rs.auth(ctx)
	if err != nil {
		return
	}
	return handler(ctx, req)
}

func (rs *Server) generateCertFile() error {
	if rs.config.RpcCert == "" {
		rs.config.RpcCert = rs.config.DataDir + "/server.pem"
	}
	if rs.config.RpcCertKey == "" {
		rs.config.RpcCertKey = rs.config.DataDir + "/server.key"
	}
	if !utils.IsExist(rs.config.RpcCert) || !utils.IsExist(rs.config.RpcCertKey) {
		return certgen.GenCertPair(rs.config.RpcCert, rs.config.RpcCertKey)
	}
	return nil
}
