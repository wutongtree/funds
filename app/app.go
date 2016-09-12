package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gocraft/web"
	"github.com/hyperledger/fabric/core/comm"
	"github.com/hyperledger/fabric/core/crypto"
	"github.com/hyperledger/fabric/core/util"
	pb "github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

// restResult defines the response payload for a general REST interface request.
type restResult struct {
	OK    string `json:",omitempty"`
	Error string `json:",omitempty"`
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
	appLogger = logging.MustGetLogger("app")

	restURL = "http://localhost:7050/"

	admin     crypto.Client
	adminCert crypto.CertificateHandler

	chaincodeID = &pb.ChaincodeID{
		Path: "funds/chaincode",
		Name: "fund_managment",
	}
)

func deploy() (err error) {
	appLogger.Debug("---------app deploy----------")

	adminCert, err = admin.GetTCertificateHandlerNext()
	if err != nil {
		appLogger.Errorf("Failed getting admin TCert [%s]", err)
		return
	}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "deploy",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs("init"),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		appLogger.Errorf("Failed marshal request body [%s]", err)
		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		appLogger.Errorf("Failed to request restful api [%s]", err)
		return
	}

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		appLogger.Errorf("Failed to unmarshal rpc response [%s]", err)
		return
	}

	if result.Result.Status != "OK" {
		appLogger.Errorf("deploy error.")

		return
	}

	appLogger.Debugf("Resp [%s]", string(respBody))

	appLogger.Debug("------------- Done!")

	return
}

//创建基金
func (s *FundManageAPP) create(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}
	appLogger.Infof("create fund Request: %v", fund)

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund name may not be blank."})
		appLogger.Error("Error: fund name may not be blank.")

		return
	}

	if fund.Assets <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund Assets maust be > 0"})
		appLogger.Error("Error: fund Assets maust be > 0")

		return
	}

	if fund.Funds <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund funds maust be > 0"})
		appLogger.Error("Error: fund Assets maust be > 0")

		return
	}

	if fund.PartnerAssets < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund PartnerAssets maust be >= 0"})
		appLogger.Error("Error: fund PartnerAssets maust be >= 0")

		return
	}

	if fund.PartnerTime < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund PartnerTime maust be >= 0"})
		appLogger.Error("Error: fund PartnerTime maust be >= 0")

		return
	}

	if fund.BuyStart < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund BuyStart maust be >= 0"})
		appLogger.Error("Error: fund BuyStart maust be >= 0")

		return
	}

	if fund.BuyPer < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund eBuyPernt maust be >= 0"})
		appLogger.Error("Error: fund BuyPer maust be >= 0")

		return
	}

	if fund.BuyAll < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund BuyAll maust be >= 0"})
		appLogger.Error("Error: fund BuyAll maust be >= 0")

		return
	}

	if fund.Net <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund net maust be > 0"})
		appLogger.Error("Error: fund net maust be > 0")

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

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "invoke",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("create fund Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "create fund Error"})
		appLogger.Errorf("create fund Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: "OK"})
	appLogger.Errorf("create func OK...")

	return
}

//设置基金净值
func (s *FundManageAPP) setNet(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}
	appLogger.Infof("create fund Request: %v", fund)

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund name may not be blank."})
		appLogger.Error("Error: fund name may not be blank.")

		return
	}

	if fund.Net <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund ent maust be > 0"})
		appLogger.Error("Error: fund net maust be > 0")

		return
	}

	args := []string{"setFundNet",
		fund.Name,
		strconv.FormatInt(fund.Net, 10)}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "invoke",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("set fund net Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "set fund net Error"})
		appLogger.Errorf("set fund net Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: "OK"})
	appLogger.Errorf("set fund net OK...")

	return
}

//设置基金限制
func (s *FundManageAPP) setLimit(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}
	appLogger.Infof("create fund Request: %v", fund)

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund name may not be blank."})
		appLogger.Error("Error: fund name may not be blank.")

		return
	}

	if fund.PartnerAssets < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund PartnerAssets maust be >= 0"})
		appLogger.Error("Error: fund PartnerAssets maust be >= 0")

		return
	}

	if fund.PartnerTime < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund PartnerTime maust be >= 0"})
		appLogger.Error("Error: fund PartnerTime maust be >= 0")

		return
	}

	if fund.BuyStart < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund BuyStart maust be >= 0"})
		appLogger.Error("Error: fund BuyStart maust be >= 0")

		return
	}

	if fund.BuyPer < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund eBuyPernt maust be >= 0"})
		appLogger.Error("Error: fund BuyPer maust be >= 0")

		return
	}

	if fund.BuyAll < 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund BuyAll maust be >= 0"})
		appLogger.Error("Error: fund BuyAll maust be >= 0")

		return
	}

	args := []string{"setFundLimit",
		fund.Name,
		strconv.FormatInt(fund.PartnerAssets, 10),
		strconv.FormatInt(fund.PartnerTime, 10),
		strconv.FormatInt(fund.BuyStart, 10),
		strconv.FormatInt(fund.BuyPer, 10),
		strconv.FormatInt(fund.BuyAll, 10)}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "invoke",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("set fund net Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "set fund net Error"})
		appLogger.Errorf("set fund net Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: "OK"})
	appLogger.Errorf("set fund net OK...")

	return
}

//扩股回购
func (s *FundManageAPP) setPool(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}
	appLogger.Infof("create fund Request: %v", fund)

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund name may not be blank."})
		appLogger.Error("Error: fund name may not be blank.")

		return
	}

	if fund.Funds <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund funds maust be > 0"})
		appLogger.Error("Error: fund Assets maust be > 0")

		return
	}

	args := []string{"setFoundPool",
		fund.Name,
		strconv.FormatInt(fund.Funds, 10)}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "invoke",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("create fund Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "create fund Error"})
		appLogger.Errorf("create fund Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: "OK"})
	appLogger.Errorf("create func OK...")

	return
}

//认购赎回
func (s *FundManageAPP) transfer(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}
	appLogger.Infof("create fund Request: %v", fund)

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund name may not be blank."})
		appLogger.Error("Error: fund name may not be blank.")

		return
	}

	if fund.Assets <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund Assets maust be > 0"})
		appLogger.Error("Error: fund Assets maust be > 0")

		return
	}

	if fund.Funds <= 0 {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund funds maust be > 0"})
		appLogger.Error("Error: fund Assets maust be > 0")

		return
	}

	args := []string{"transferFound",
		fund.Name,
		strconv.FormatInt(fund.Funds, 10)}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "invoke",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("create fund Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "create fund Error"})
		appLogger.Errorf("create fund Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: "OK"})
	appLogger.Errorf("create func OK...")

	return
}

//查询基金
func (s *FundManageAPP) getFund(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}
	appLogger.Infof("create fund Request: %v", fund)

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund name may not be blank."})
		appLogger.Error("Error: fund name may not be blank.")

		return
	}

	args := []string{"queryFundInfo",
		"one",
		fund.Name,
	}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "query",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("create fund Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "create fund Error"})
		appLogger.Errorf("create fund Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: result.Result.Message})
	appLogger.Errorf("create func OK...")

	return
}

//查询基金列表
func (s *FundManageAPP) getFundList(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	args := []string{"queryFundInfo",
		"list",
	}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "query",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("create fund Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "create fund Error"})
		appLogger.Errorf("create fund Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: result.Result.Message})
	appLogger.Errorf("create func OK...")

	return
}

//查询用户自己信息
func (s *FundManageAPP) getUser(rw web.ResponseWriter, req *web.Request) {
	encoder := json.NewEncoder(rw)

	// Decode the incoming JSON payload
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}

	var fund fundInfo
	err = json.Unmarshal(body, &fund)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("Error: %s", err)

		return
	}
	appLogger.Infof("create fund Request: %v", fund)

	// Check that the name,fund,assets are not left blank.
	if fund.Name == "" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "fund name may not be blank."})
		appLogger.Error("Error: fund name may not be blank.")

		return
	}

	args := []string{"queryUserInfo",
		fund.Name,
	}

	request := &rpcRequest{
		Jsonrpc: "2.0",
		Method:  "query",
		Params: &pb.ChaincodeSpec{
			Type:        pb.ChaincodeSpec_GOLANG,
			ChaincodeID: chaincodeID,
			CtorMsg: &pb.ChaincodeInput{
				Args: util.ToChaincodeArgs(args...),
			},
			//Timeout:1,
			SecureContext:        "lukas",
			ConfidentialityLevel: pb.ConfidentialityLevel_CONFIDENTIAL,
			Metadata:             adminCert.GetCertificate(),
			//Attributes:[]string{},
		},
		// ID: &rpcID{
		// 	StringValue: "123",
		// 	IntValue:    int64(123),
		// },
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "Marshal error."})
		appLogger.Errorf("Error: Marshal error: %v", err)

		return
	}

	respBody, err := doHTTPPost(restURL+"chaincode", reqBody)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "get data error."})
		appLogger.Error("Error: get data error.")

		return
	}
	appLogger.Debugf("url response: %v", string(respBody))

	result := new(rpcResponse)
	err = json.Unmarshal(respBody, result)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: err.Error()})
		appLogger.Errorf("query user Error: %s", err)
		return
	}

	if result.Result.Status != "OK" {
		rw.WriteHeader(http.StatusBadRequest)
		encoder.Encode(restResult{Error: "query user Error"})
		appLogger.Errorf("query user Error")

		return
	}

	rw.WriteHeader(http.StatusOK)
	encoder.Encode(restResult{OK: result.Result.Message})
	appLogger.Errorf("query user OK...")

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
	json.NewEncoder(rw).Encode(restResult{Error: "Openchain endpoint not found."})
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
	router.Post("/create", (*FundManageAPP).create)
	router.Post("/setnet", (*FundManageAPP).setNet)
	router.Post("/setLimit", (*FundManageAPP).setLimit)
	router.Post("/setPool", (*FundManageAPP).setPool)
	router.Post("/transfer", (*FundManageAPP).transfer)
	router.Get("/getFund", (*FundManageAPP).getFund)
	router.Get("/chgetFundList", (*FundManageAPP).getFundList)
	router.Get("/getUser", (*FundManageAPP).getUser)

	// Add not found page
	router.NotFound((*FundManageAPP).NotFound)

	return router
}

func main() {
	deploy()

	router := buildRESTRouter()

	// Start server
	if comm.TLSEnabled() {
		err := http.ListenAndServeTLS(viper.GetString("rest.address"), viper.GetString("peer.tls.cert.file"), viper.GetString("peer.tls.key.file"), router)
		if err != nil {
			appLogger.Errorf("ListenAndServeTLS: %s", err)
		}
	} else {
		err := http.ListenAndServe(viper.GetString("rest.address"), router)
		if err != nil {
			appLogger.Errorf("ListenAndServe: %s", err)
		}
	}
}
