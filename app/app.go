package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gocraft/web"
	"github.com/hyperledger/fabric/core/crypto"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"github.com/hyperledger/fabric/core/util"
	pb "github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// restResult defines the response payload for a general REST interface request.
type restResult struct {
	Status string `json:status",omitempty"`
	Msg    string `json:msg",omitempty"`
}

// rpcRequest defines the JSON RPC 2.0 request payload for the /chaincode endpoint.
type rpcRequest struct {
	Jsonrpc string            `json:"jsonrpc,omitempty"`
	Method  string            `json:"method,omitempty"`
	Params  *pb.ChaincodeSpec `json:"params,omitempty"`
	ID      *rpcID            `json:"id,omitempty"`
}

type rpcID struct {
	StringValue string
	IntValue    int64
}

// rpcResponse defines the JSON RPC 2.0 response payload for the /chaincode endpoint.
type rpcResponse struct {
	Jsonrpc string     `json:"jsonrpc,omitempty"`
	Result  *rpcResult `json:"result,omitempty"`
	Error   *rpcError  `json:"error,omitempty"`
	ID      *rpcID     `json:"id"`
}

// rpcResult defines the structure for an rpc sucess/error result message.
type rpcResult struct {
	Status  string    `json:"status,omitempty"`
	Message string    `json:"message,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

// rpcError defines the structure for an rpc error.
type rpcError struct {
	// A Number that indicates the error type that occurred. This MUST be an integer.
	Code int64 `json:"code,omitempty"`
	// A String providing a short description of the error. The message SHOULD be
	// limited to a concise single sentence.
	Message string `json:"message,omitempty"`
	// A Primitive or Structured value that contains additional information about
	// the error. This may be omitted. The value of this member is defined by the
	// Server (e.g. detailed error information, nested errors etc.).
	Data string `json:"data,omitempty"`
}

type FundManageAPP struct {
}

type fundInfo struct {
	EnrollID      string `json:"enrollID,omitempty"`
	Name          string `json:"name,omitempty"`
	Funds         int64  `json:"funds,omitempty"`
	Assets        int64  `json:"assets,omitempty"`
	PartnerAssets int64  `json:"partnerAssets,omitempty"`
	PartnerTime   int64  `json:"partnerTime,omitempty"`
	BuyStart      int64  `json:"buyStart,omitempty"`
	BuyPer        int64  `json:"buyPer,omitempty"`
	BuyAll        int64  `json:"buyAll,omitempty"`
	Net           int64  `json:"net,omitempty"`
}

var (
	// Logging
	appLogger = logging.MustGetLogger("app")

	// NVP related objects
	peerClientConn *grpc.ClientConn
	serverClient   pb.PeerClient

	// Alice is the deployer
	admin crypto.Client
)

func deploy() (err error) {
	appLogger.Debug("------------- deploy...")

	resp, err := deployInternal()
	if err != nil {
		appLogger.Errorf("Failed deploying [%s]", err)
		return
	}
	if resp.Status != pb.Response_SUCCESS {
		appLogger.Errorf("Failed deploying [%s]", string(resp.Msg))
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())
	appLogger.Debugf("Chaincode NAME: [%s]-[%s]", chaincodeName, string(resp.Msg))

	appLogger.Debug("------------- deploy Done!")
	return

	// adminCert, err = admin.GetTCertificateHandlerNext()
	// if err != nil {
	// 	appLogger.Errorf("Failed getting admin TCert [%s]", err)
	// 	return
	// }

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "deploy",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Path: chaincodePath,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs("init"),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	appLogger.Errorf("Failed marshal request body [%s]", err)
	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	appLogger.Errorf("Failed to request restful api [%s]", err)
	// 	return
	// }

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	appLogger.Errorf("Failed to unmarshal rpc response [%s]", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	appLogger.Errorf("deploy error.")

	// 	return
	// }

	// appLogger.Debugf("Resp [%s]", string(respBody))

	// chaincodeName = result.Result.Message

	// appLogger.Debug("------------- Done!")

	// return
}

//创建基金
func (s *FundManageAPP) create(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- create ...")

	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed createFund: [%s]", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed createFund: [%s]", err)

		return
	}
	appLogger.Debugf("create fund Request: %v", fund)

	invoker, err := setCryptoClient(fund.EnrollID, "")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed createFund: [%s]", err)

		return
	}

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund name may not be blank."})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund name may not be blank"))

		return
	}

	if fund.Assets <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund Assets maust be > 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund Assets maust be > 0"))

		return
	}

	if fund.Funds <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund funds maust be > 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund funds maust be > 0"))

		return
	}

	if fund.PartnerAssets < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund PartnerAssets maust be >= 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund PartnerAssets maust be >= 0"))

		return
	}

	if fund.PartnerTime < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund PartnerTime maust be >= 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund PartnerTime maust be >= 0"))

		return
	}

	if fund.BuyStart < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund BuyStart maust be >= 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund BuyStart maust be >= 0"))

		return
	}

	if fund.BuyPer < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund eBuyPernt maust be >= 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund eBuyPernt maust be >= 0"))

		return
	}

	if fund.BuyAll < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund BuyAll maust be >= 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund BuyAll maust be >= 0"))

		return
	}

	if fund.Net <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund net maust be > 0"})
		appLogger.Errorf("Failed createFund: [%s]", errors.New("fund net maust be > 0"))

		return
	}

	args := []string{"createFund",
		fund.Name,
		strconv.FormatInt(fund.Funds, 10),
		strconv.FormatInt(fund.Assets, 10),
		strconv.FormatInt(fund.PartnerAssets, 10),
		strconv.FormatInt(fund.PartnerTime, 10),
		strconv.FormatInt(fund.BuyStart, 10),
		strconv.FormatInt(fund.BuyPer, 10),
		strconv.FormatInt(fund.BuyAll, 10),
		strconv.FormatInt(fund.Net, 10)}

	invokerCert, err := invoker.GetTCertificateHandlerNext()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed createFund [%s]", err)
		return
	}

	resp, err := invokeInternal(invoker, invokerCert, &pb.ChaincodeInput{Args: util.ToChaincodeArgs(args...)})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed createFund [%s]", err)
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())

	appLogger.Debug("------------- create Done!")

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "invoke",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Name: chaincodeName,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs(args...),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "Marshal error."})
	// 	appLogger.Errorf("Error: Marshal error: %v", err)

	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "get data error."})
	// 	appLogger.Error("Error: get data error.")

	// 	return
	// }
	// appLogger.Debugf("url response: %v", string(respBody))

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
	// 	appLogger.Errorf("create fund Error: %s", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "create fund Error"})
	// 	appLogger.Errorf("create fund Error")

	// 	return
	// }

	// rw.WriteHeader(http.StatusOK)
	// encoder.Encode(restResult{Status: "OK"})
	// appLogger.Errorf("create func OK...")
	if resp.Status != pb.Response_SUCCESS {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: string(resp.Msg)})
		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: "successful create fund"})

	return
}

//设置基金净值
func (s *FundManageAPP) setNet(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- setNet ...")

	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set net: [%s]", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set net: [%s]", err)

		return
	}
	appLogger.Debugf("set net Request: %v", fund)

	invoker, err := setCryptoClient(fund.EnrollID, "")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set net: [%s]", err)

		return
	}

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund name may not be blank"})
		appLogger.Errorf("Failed set net: [%s]", errors.New("fund name may not be blank"))

		return
	}

	if fund.Net <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund ent maust be > 0"})
		appLogger.Errorf("Failed set net: [%s]", errors.New("fund ent maust be > 0"))

		return
	}

	args := []string{"setFundNet",
		fund.Name,
		strconv.FormatInt(fund.Net, 10)}

	invokerCert, err := invoker.GetTCertificateHandlerNext()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set net: [%s]", err)
		return
	}

	resp, err := invokeInternal(invoker, invokerCert, &pb.ChaincodeInput{Args: util.ToChaincodeArgs(args...)})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set net: [%s]", err)
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "invoke",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Name: chaincodeName,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs(args...),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "Marshal error."})
	// 	appLogger.Errorf("Error: Marshal error: %v", err)

	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "get data error."})
	// 	appLogger.Error("Error: get data error.")

	// 	return
	// }
	// appLogger.Debugf("url response: %v", string(respBody))

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
	// 	appLogger.Errorf("set fund net Error: %s", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "set fund net Error"})
	// 	appLogger.Errorf("set fund net Error")

	// 	return
	// }

	// rw.WriteHeader(http.StatusOK)
	// encoder.Encode(restResult{Status: "OK"})
	// appLogger.Errorf("set fund net OK...")
	if resp.Status != pb.Response_SUCCESS {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: string(resp.Msg)})
		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: "successful set net"})

	appLogger.Debug("------------- setNet Done!")

	return
}

//设置基金限制
func (s *FundManageAPP) setLimit(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- setLimit ...")

	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set limit: [%s]", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set limit: [%s]", err)

		return
	}
	appLogger.Debugf("set limit Request: %v", fund)

	invoker, err := setCryptoClient(fund.EnrollID, "")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set limit: [%s]", err)

		return
	}

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund name may not be blank"})
		appLogger.Errorf("Failed set limit: [%s]", errors.New("fund name may not be blank"))

		return
	}

	if fund.PartnerAssets < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund PartnerAssets maust be >= 0"})
		appLogger.Errorf("Failed set limit: [%s]", errors.New("fund PartnerAssets maust be >= 0"))

		return
	}

	if fund.PartnerTime < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund PartnerTime maust be >= 0"})
		appLogger.Errorf("Failed set limit: [%s]", errors.New("fund PartnerTime maust be >= 0"))

		return
	}

	if fund.BuyStart < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund BuyStart maust be >= 0"})
		appLogger.Errorf("Failed set limit: [%s]", errors.New("fund BuyStart maust be >= 0"))

		return
	}

	if fund.BuyPer < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund eBuyPernt maust be >= 0"})
		appLogger.Errorf("Failed set limit: [%s]", errors.New("fund eBuyPernt maust be >= 0"))

		return
	}

	if fund.BuyAll < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund BuyAll maust be >= 0"})
		appLogger.Errorf("Failed set limit: [%s]", errors.New("fund BuyAll maust be >= 0"))

		return
	}

	args := []string{"setFundLimit",
		fund.Name,
		strconv.FormatInt(fund.PartnerAssets, 10),
		strconv.FormatInt(fund.PartnerTime, 10),
		strconv.FormatInt(fund.BuyStart, 10),
		strconv.FormatInt(fund.BuyPer, 10),
		strconv.FormatInt(fund.BuyAll, 10)}

	invokerCert, err := invoker.GetTCertificateHandlerNext()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set limit: [%s]", err)
		return
	}

	resp, err := invokeInternal(invoker, invokerCert, &pb.ChaincodeInput{Args: util.ToChaincodeArgs(args...)})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set limit: [%s]", err)
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "invoke",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Name: chaincodeName,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs(args...),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "Marshal error."})
	// 	appLogger.Errorf("Error: Marshal error: %v", err)

	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "get data error."})
	// 	appLogger.Error("Error: get data error.")

	// 	return
	// }
	// appLogger.Debugf("url response: %v", string(respBody))

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
	// 	appLogger.Errorf("set fund net Error: %s", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "set fund net Error"})
	// 	appLogger.Errorf("set fund net Error")

	// 	return
	// }

	// rw.WriteHeader(http.StatusOK)
	// encoder.Encode(restResult{Status: "OK"})
	// appLogger.Errorf("set fund net OK...")
	if resp.Status != pb.Response_SUCCESS {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: string(resp.Msg)})
		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: "successful set limit"})

	appLogger.Debug("------------- setLimit Done!")

	return
}

//扩股回购
func (s *FundManageAPP) setPool(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- setPool ...")

	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set pool: [%s]", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set pool: [%s]", err)

		return
	}
	appLogger.Debugf("set pool Request: %v", fund)

	invoker, err := setCryptoClient(fund.EnrollID, "")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set pool: [%s]", err)

		return
	}

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund name may not be blank"})
		appLogger.Errorf("Failed set pool: [%s]", errors.New("fund name may not be blank"))

		return
	}

	if fund.Funds <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund funds maust be > 0"})
		appLogger.Errorf("Failed set pool: [%s]", errors.New("fund funds maust be > 0"))

		return
	}

	args := []string{"setFoundPool",
		fund.Name,
		strconv.FormatInt(fund.Funds, 10)}

	invokerCert, err := invoker.GetTCertificateHandlerNext()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set pool: [%s]", err)
		return
	}

	resp, err := invokeInternal(invoker, invokerCert, &pb.ChaincodeInput{Args: util.ToChaincodeArgs(args...)})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed set pool: [%s]", err)
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "invoke",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Name: chaincodeName,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs(args...),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "Marshal error."})
	// 	appLogger.Errorf("Error: Marshal error: %v", err)

	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "get data error."})
	// 	appLogger.Error("Error: get data error.")

	// 	return
	// }
	// appLogger.Debugf("url response: %v", string(respBody))

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
	// 	appLogger.Errorf("create fund Error: %s", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "create fund Error"})
	// 	appLogger.Errorf("create fund Error")

	// 	return
	// }

	// rw.WriteHeader(http.StatusOK)
	// encoder.Encode(restResult{Status: "OK"})
	// appLogger.Errorf("create func OK...")
	if resp.Status != pb.Response_SUCCESS {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: string(resp.Msg)})
		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: "successful set pool"})

	appLogger.Debug("------------- setPool Done")

	return
}

//认购赎回
func (s *FundManageAPP) transfer(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- transfer...")

	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed transfer: [%s]", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed transfer: [%s]", err)

		return
	}
	appLogger.Debugf("transfer fund Request: %v", fund)

	invoker, err := setCryptoClient(fund.EnrollID, "")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed transfer: [%s]", err)

		return
	}

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund name may not be blank"})
		appLogger.Errorf("Failed transfer: [%s]", errors.New("fund name may not be blank"))

		return
	}

	if fund.Assets <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund Assets maust be > 0"})
		appLogger.Errorf("Failed transfer: [%s]", errors.New("fund Assets maust be > 0"))

		return
	}

	if fund.Funds <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund funds maust be > 0"})
		appLogger.Errorf("Failed transfer: [%s]", errors.New("fund funds maust be > 0"))

		return
	}

	invokerCert, err := invoker.GetTCertificateHandlerNext()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed transfer: [%s]", err)
		return
	}

	args := []string{"transferFound",
		string(invokerCert.GetCertificate()),
		fund.Name,
		strconv.FormatInt(fund.Funds, 10)}

	resp, err := invokeInternal(invoker, invokerCert, &pb.ChaincodeInput{Args: util.ToChaincodeArgs(args...)})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed transfer: [%s]", err)
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "invoke",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Name: chaincodeName,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs(args...),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "Marshal error."})
	// 	appLogger.Errorf("Error: Marshal error: %v", err)

	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "get data error."})
	// 	appLogger.Error("Error: get data error.")

	// 	return
	// }
	// appLogger.Debugf("url response: %v", string(respBody))

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
	// 	appLogger.Errorf("create fund Error: %s", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "create fund Error"})
	// 	appLogger.Errorf("create fund Error")

	// 	return
	// }

	// rw.WriteHeader(http.StatusOK)
	// encoder.Encode(restResult{Status: "OK"})
	// appLogger.Errorf("create func OK...")
	if resp.Status != pb.Response_SUCCESS {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: string(resp.Msg)})
		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: "successful transfer fund"})

	appLogger.Debug("------------- transfer Done")

	return
}

//查询基金
func (s *FundManageAPP) getFund(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- query fund...")

	encoder := json.NewEncoder(rw)

	fundName := req.PathParams["name"]

	// Check that the name,fund,assets are not left blank.
	if fundName == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund name may not be blank"})
		appLogger.Errorf("Failed query fund: [%s]", errors.New("fund name may not be blank"))

		return
	}

	args := []string{"queryFundInfo",
		"one",
		fundName,
	}
	resp, err := queryInternal(admin, &pb.ChaincodeInput{Args: util.ToChaincodeArgs(args...)})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed query fund: [%s]", err)
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "query",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Name: chaincodeName,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs(args...),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "Marshal error."})
	// 	appLogger.Errorf("Error: Marshal error: %v", err)

	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "get data error."})
	// 	appLogger.Error("Error: get data error.")

	// 	return
	// }
	// appLogger.Debugf("url response: %v", string(respBody))

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
	// 	appLogger.Errorf("create fund Error: %s", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "create fund Error"})
	// 	appLogger.Errorf("create fund Error")

	// 	return
	// }

	// rw.WriteHeader(http.StatusOK)
	// encoder.Encode(restResult{Status: "OK", Msg: result.Result.Message})
	// appLogger.Errorf("create func OK...")
	if resp.Status != pb.Response_SUCCESS {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: string(resp.Msg)})
		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: string(resp.Msg)})
	appLogger.Debug("------------- query fund Done")

	return
}

//查询用户自己信息
func (s *FundManageAPP) getUser(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- query user ...")

	encoder := json.NewEncoder(rw)

	fundName := req.PathParams["fundName"]
	enrollID := req.PathParams["enrollID"]

	invoker, err := setCryptoClient(enrollID, "")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed query user: [%s]", err)

		return
	}

	// Check that the name,fund,assets are not left blank.
	if fundName == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "fund name may not be blank"})
		appLogger.Errorf("Failed query user: [%s]", errors.New("fund name may not be blank"))

		return
	}

	invokerCert, err := invoker.GetTCertificateHandlerNext()
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed query user: [%s]", err)
		return
	}

	args := []string{"queryUserInfo",
		string(invokerCert.GetCertificate()),
		fundName,
	}

	resp, err := queryInternal(invoker, &pb.ChaincodeInput{Args: util.ToChaincodeArgs(args...)})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed query user: [%s]", err)
		return
	}
	appLogger.Debugf("Resp [%s]", resp.String())

	// request := &rpcRequest{
	// 	Jsonrpc: "2.0",
	// 	Method:  "query",
	// 	Params: &pb.ChaincodeSpec{
	// 		Type: pb.ChaincodeSpec_GOLANG,
	// 		ChaincodeID: &pb.ChaincodeID{
	// 			Name: chaincodeName,
	// 		},
	// 		CtorMsg: &pb.ChaincodeInput{
	// 			Args: util.ToChaincodeArgs(args...),
	// 		},
	// 		//Timeout:1,
	// 		SecureContext:        "lukas",
	// 		ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
	// 		// Metadata:             adminCert.GetCertificate(),
	// 		//Attributes:[]string{},
	// 	},
	// 	// ID: &rpcID{
	// 	// 	StringValue: "123",
	// 	// 	IntValue:    int64(123),
	// 	// },
	// }

	// reqBody, err := json.Marshal(request)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "Marshal error."})
	// 	appLogger.Errorf("Error: Marshal error: %v", err)

	// 	return
	// }

	// respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "get data error."})
	// 	appLogger.Error("Error: get data error.")

	// 	return
	// }
	// appLogger.Debugf("url response: %v", string(respBody))

	// result := new(rpcResponse)
	// err = json.Unmarshal(respBody, result)
	// if err != nil {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
	// 	appLogger.Errorf("query user Error: %s", err)
	// 	return
	// }

	// if result.Result.Status != "OK" {
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	encoder.Encode(restResult{Status: "Err", Msg: "query user Error"})
	// 	appLogger.Errorf("query user Error")

	// 	return
	// }

	// rw.WriteHeader(http.StatusOK)
	// encoder.Encode(restResult{Status: "OK", Msg: result.Result.Message})
	// appLogger.Errorf("query user OK...")
	if resp.Status != pb.Response_SUCCESS {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: string(resp.Msg)})
		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: string(resp.Msg)})

	appLogger.Debug("------------- query user Done")

	return
}

// login confirms the account and secret password of the client with the
// CA and stores the enrollment certificate and key in the Devops server.
func (s *FundManageAPP) login(rw web.ResponseWriter, req *web.Request) {
	appLogger.Debug("------------- login...")

	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed login: [%s]", err)

		return
	}

	loginRequest := struct {
		EnrollID     string `protobuf:"bytes,1,opt,name=enrollId" json:"enrollId,omitempty"`
		EnrollSecret string `protobuf:"bytes,2,opt,name=enrollSecret" json:"enrollSecret,omitempty"`
	}{}

	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: err.Error()})
		appLogger.Errorf("Failed login: [%s]", err)

		return
	}

	// Check that the enrollId and enrollSecret are not left blank.
	if (loginRequest.EnrollID == "") || (loginRequest.EnrollSecret == "") {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: "enrollId and enrollSecret may not be blank."})
		appLogger.Errorf("Failed login: [%s]", errors.New("enrollId and enrollSecret may not be blank"))

		return
	}

	_, err = setCryptoClient(loginRequest.EnrollID, loginRequest.EnrollSecret)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Status: "Err", Msg: fmt.Sprintf("Login error: %v", err)})
		appLogger.Errorf("Failed login: [%s]", err)

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{Status: "OK", Msg: fmt.Sprintf("Login successful for user '%s'.", loginRequest.EnrollID)})
	appLogger.Debugf("Login successful for user '%s'.\n", loginRequest.EnrollID)

	appLogger.Debug("------------- login Done")

	return
}

func doHTTPPost(url string, reqBody []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json;charset=utf-8", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// NotFound returns a custom landing page when a given hyperledger end point
// had not been defined.
func (s *FundManageAPP) NotFound(rw web.ResponseWriter, r *web.Request) {
	rw.WriteHeader(http.StatusNotFound)
	json.NewEncoder(rw).Encode(restResult{Status: "Err", Msg: "Openchain endpoint not found."})
}

// SetResponseType is a middleware function that sets the appropriate response
// headers. Currently, it is setting the "Content-Type" to "application/json" as
// well as the necessary headers in order to enable CORS for Swagger usage.
func (s *FundManageAPP) SetResponseType(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	rw.Header().Set("Content-Type", "application/json")

	// Enable CORS
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "accept, content-type")

	next(rw, req)
}

func buildRESTRouter() *web.Router {
	router := web.New(FundManageAPP{})

	// Add middleware
	router.Middleware((*FundManageAPP).SetResponseType)

	// Add routes
	router.Post("/login", (*FundManageAPP).login)
	router.Post("/create", (*FundManageAPP).create)
	router.Post("/setnet", (*FundManageAPP).setNet)
	router.Post("/setLimit", (*FundManageAPP).setLimit)
	router.Post("/setPool", (*FundManageAPP).setPool)
	router.Post("/transfer", (*FundManageAPP).transfer)
	router.Get("/getFund", (*FundManageAPP).getFund)
	router.Get("/getUser", (*FundManageAPP).getUser)

	// Add not found page
	router.NotFound((*FundManageAPP).NotFound)

	return router
}

func main() {
	logging.SetLevel(logging.DEBUG, "app")

	primitives.SetSecurityLevel("SHA3", 256)
	// Initialize a non-validating peer whose role is to submit
	// transactions to the fabric network.
	// A 'core.yaml' file is assumed to be available in the working directory.
	if err := initNVP(); err != nil {
		appLogger.Debugf("Failed initiliazing NVP [%s]", err)
		os.Exit(-1)
	}

	crypto.Init()

	// Enable fabric 'confidentiality'
	confidentiality(true)

	// Deploy
	if err := deploy(); err != nil {
		appLogger.Errorf("Failed deploying [%s]", err)
		os.Exit(-1)
	}

	router := buildRESTRouter()
	address := viper.GetString("app.address")
	appLogger.Debugf("App server start [%s]", address)

	// Start server
	if viper.GetBool("app.tls.enabled") {
		err := http.ListenAndServeTLS(address, viper.GetString("app.tls.cert.file"), viper.GetString("app.tls.key.file"), router)
		if err != nil {
			appLogger.Errorf("ListenAndServeTLS: %s", err)
			os.Exit(-1)
		}
	} else {
		err := http.ListenAndServe(address, router)
		if err != nil {
			appLogger.Errorf("ListenAndServe: %s", err)
			os.Exit(-1)
		}
	}
}
