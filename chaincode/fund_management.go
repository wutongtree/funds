package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("fund_mgm")

type FundManagementChaincode struct {
}

// Init method will be called during deployment.
// The deploy transaction metadata is supposed to contain the administrator cert
func (t *FundManagementChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	myLogger.Debug("Init Chaincode......")
	function, args, _ = dealParam(function, args)

	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}

	//create table
	err := createTable(stub)
	if err != nil {
		return nil, err
	}

	//set the admin
	// the  metadata will contain the certificate of the administrator
	// adminCert, err := stub.GetCallerMetadata()
	// if err != nil {
	// 	myLogger.Debug("Failed getting metadata")
	// 	return nil, errors.New("Failed getting metadata.")
	// }
	// if len(adminCert) == 0 {
	// 	myLogger.Debug("Invalid admin certificate. Empty.")
	// 	return nil, errors.New("Invalid admin certificate. Empty.")
	// }

	// myLogger.Debug("The administrator is [%x]", adminCert)

	// stub.PutState("admin", adminCert)

	myLogger.Debug("Init Chaincode...done")

	return nil, nil
}

// Invoke will be called for every transaction.
func (t *FundManagementChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	myLogger.Debug("Invoke Chaincode......")
	function, args, _ = dealParam(function, args)

	// Handle different functions
	if function == "create" {
		return t.createFund(stub, args)
	} else if function == "setFundNet" {
		return t.setFundNet(stub, args)
	} else if function == "setFundLimit" {
		return t.setFundLimit(stub, args)
	} else if function == "setFundPool" {
		return t.setFundPool(stub, args)
	} else if function == "transferFund" {
		return t.transferFund(stub, args)
	}
	return nil, errors.New("Received unknown function invocation")
}

// Query callback representing the query of a chaincode
// Anyone can invoke this function.
func (t *FundManagementChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	myLogger.Debug("Query Chaincode....")
	function, args, _ = dealParam(function, args)

	if function == "queryFundInfo" {
		return t.queryFundInfo(stub, args)
	} else if function == "queryUserInfo" {
		return t.queryUserInfo(stub, args)
	}

	return nil, errors.New("Received unknown function query")
}

func dealParam(function string, args []string) (string, []string, error) {
	function_b, err := base64.StdEncoding.DecodeString(function)
	if err != nil {
		return "", nil, err
	}
	for k, v := range args {
		arg_b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return "", nil, err
		}
		args[k] = string(arg_b)
	}

	return string(function_b), args, nil
}

func createTable(stub shim.ChaincodeStubInterface) error {
	// 1. 基金基本信息：基金名称、管理员、基金池容量、基金池中剩余基金数、系统资金量、参与者资金量、参与者注册时间、认购起点、认购单量、认购总量
	err := stub.CreateTable("FundInfo", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Name", Type: shim.ColumnDefinition_STRING, Key: true},
		// &shim.ColumnDefinition{Name: "Admin", Type: shim.ColumnDefinition_BYTES, Key: false},
		// &shim.ColumnDefinition{Name: "FundPool", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "Funds", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "Assets", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "PartnerAssets", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "PartnerTime", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "BuyStart", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "BuyPer", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "BuyAll", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "Net", Type: shim.ColumnDefinition_INT64, Key: false},
	})
	if err != nil {
		myLogger.Errorf("Failed creating FundInfo table: %s", err)
		return fmt.Errorf("Failed creating FundInfo table: %s", err)
	}

	// 2. 基金净值：基金名、净值、时间(时间戳)
	// err = stub.CreateTable("FundNet", []*shim.ColumnDefinition{
	// 	&shim.ColumnDefinition{Name: "Name", Type: shim.ColumnDefinition_STRING, Key: true},
	// 	&shim.ColumnDefinition{Name: "Net", Type: shim.ColumnDefinition_INT64, Key: false},
	// 	// &shim.ColumnDefinition{Name: "Time", Type: shim.ColumnDefinition_INT64, Key: true},
	// })
	// if err != nil {
	// 	myLogger.Errorf("Failed creating FundNet table: %s", err)
	// 	return errors.New("Failed creating FundNet table.")
	// }

	// 3. 账户资金信息：账户证书、资金量
	// err = stub.CreateTable("AccountAsset", []*shim.ColumnDefinition{
	// 	&shim.ColumnDefinition{Name: "Owner", Type: shim.ColumnDefinition_BYTES, Key: true},
	// 	&shim.ColumnDefinition{Name: "Assets", Type: shim.ColumnDefinition_INT64, Key: false},
	// })
	// if err != nil {
	// 	myLogger.Errorf("Failed creating AccountAsset table: %s", err)
	// 	return errors.New("Failed creating AccountAsset table.")
	// }

	// 4. 用户基金信息：账户证书、基金名、所购基金份额
	err = stub.CreateTable("AccountFund", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Name", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Owner", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Assets", Type: shim.ColumnDefinition_INT64, Key: false},
		&shim.ColumnDefinition{Name: "Fund", Type: shim.ColumnDefinition_INT64, Key: false},
	})
	if err != nil {
		myLogger.Errorf("Failed creating AccountFund table: %s", err)
		return fmt.Errorf("Failed creating AccountFund table: %s", err)
	}

	// 5. 排队信息：交易者证书、基金名、交易额（+认购或-赎回）
	err = stub.CreateTable("Queue", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: "Owner", Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name: "Name", Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: "Assets", Type: shim.ColumnDefinition_INT64, Key: false},
	})
	if err != nil {
		myLogger.Errorf("Failed creating Queue table: %s", err)
		return fmt.Errorf("Failed creating Queue table: %s", err)
	}

	return nil
}

//校验是否管理员
func (t *FundManagementChaincode) isAdmin(stub shim.ChaincodeStubInterface) (bool, error) {
	// Verify the identity of the caller
	// Only an administrator can invoker assign
	adminCertificate, err := stub.GetState("admin")
	if err != nil {
		return false, errors.New("Failed fetching admin identity")
	}

	ok, err := t.isCaller(stub, adminCertificate)
	if err != nil {
		return false, errors.New("Failed checking admin identity")
	}
	if !ok {
		return false, errors.New("The caller is not an administrator")
	}
	return true, nil
}

func (t *FundManagementChaincode) isCaller(stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
	myLogger.Debug("Check caller...")

	// In order to enforce access control, we require that the
	// metadata contains the signature under the signing key corresponding
	// to the verification key inside certificate of
	// the payload of the transaction (namely, function name and args) and
	// the transaction binding (to avoid copying attacks)

	// Verify \sigma=Sign(certificate.sk, tx.Payload||tx.Binding) against certificate.vk
	// \sigma is in the metadata

	sigma, err := stub.GetCallerMetadata()
	if err != nil {
		return false, errors.New("Failed getting metadata")
	}
	payload, err := stub.GetPayload()
	if err != nil {
		return false, errors.New("Failed getting payload")
	}
	binding, err := stub.GetBinding()
	if err != nil {
		return false, errors.New("Failed getting binding")
	}

	myLogger.Debugf("passed certificate [% x]", certificate)
	myLogger.Debugf("passed sigma [% x]", sigma)
	myLogger.Debugf("passed payload [% x]", payload)
	myLogger.Debugf("passed binding [% x]", binding)

	ok, err := stub.VerifySignature(
		certificate,
		sigma,
		append(payload, binding...),
	)
	if err != nil {
		myLogger.Errorf("Failed checking signature [%s]", err)
		return ok, err
	}
	if !ok {
		myLogger.Error("Invalid signature")
	}

	myLogger.Debug("Check caller...Verified!")

	return ok, err
}

//创建基金
func (t *FundManagementChaincode) createFund(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debug("createFund......")

	if len(args) != 9 {
		return nil, errors.New("Incorrect number of arguments. Expecting 9")
	}

	// ok, err := t.isAdmin(stub)
	// if !ok {
	// 	return nil, err
	// }

	name := args[0]
	// admin, err := base64.StdEncoding.DecodeString(args[1])
	// if err != nil {
	// 	return nil, errors.New("Failed fetching admin identify")
	// }
	// fundPool, err := strconv.ParseInt(args[2], 10, 64)
	// if err != nil {
	// 	return nil, errors.New("Fund pool is not int64")
	// }

	funds, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("Fund is not int64")
	}
	assets, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, errors.New("assets is not int64")
	}
	partnerAssets, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return nil, errors.New("partner assets is not int64")
	}
	partnerTime, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return nil, errors.New("partner time is not int64")
	}
	buyStart, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return nil, errors.New("buy start is not int64")
	}
	buyPer, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		return nil, errors.New("buy per is not int64")
	}
	buyAll, err := strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		return nil, errors.New("buy all is not int64")
	}
	net, err := strconv.ParseInt(args[8], 10, 64)
	if err != nil {
		return nil, errors.New("fund net is not int64")
	}

	ok, err := stub.InsertRow("FundInfo", shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: name}},
			// &shim.Column{Value: &shim.Column_Bytes{Bytes: admin}},
			// &shim.Column{Value: &shim.Column_Int64{Int64: fundPool}},
			&shim.Column{Value: &shim.Column_Int64{Int64: funds}},
			&shim.Column{Value: &shim.Column_Int64{Int64: assets}},
			&shim.Column{Value: &shim.Column_Int64{Int64: partnerAssets}},
			&shim.Column{Value: &shim.Column_Int64{Int64: partnerTime}},
			&shim.Column{Value: &shim.Column_Int64{Int64: buyStart}},
			&shim.Column{Value: &shim.Column_Int64{Int64: buyPer}},
			&shim.Column{Value: &shim.Column_Int64{Int64: buyAll}},
			&shim.Column{Value: &shim.Column_Int64{Int64: net}}},
	})
	if !ok && err == nil {
		return nil, errors.New("the fund info was already existed")
	}

	if err != nil {
		myLogger.Errorf("insert fund info failed:%s", err)
		return nil, fmt.Errorf("insert fund info failed:%s", err)
	}

	myLogger.Debug("createFund done.")
	return nil, nil
}

//设置基金净值
func (t *FundManagementChaincode) setFundNet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debug("setFundNet.....")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	// ok, err := t.isAdmin(stub)
	// if !ok {
	// 	return nil, err
	// }

	fundName := args[0]
	fundNet, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("fund net is not int64")
	}

	_, row, err := getFundInfoByName(stub, fundName)
	if err != nil {
		return nil, err
	}

	row.Columns[8].Value = &shim.Column_Int64{Int64: fundNet}

	_, err = stub.ReplaceRow("FundInfo", *row)
	if err != nil {
		myLogger.Errorf("update fund net failed:%s", err)
		return nil, fmt.Errorf("update fund net failed:%s", err)
	}

	myLogger.Debug("setFundNetc done.")
	return nil, nil
}

//设置基金限制参数
func (t *FundManagementChaincode) setFundLimit(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debug("setFundLimit.....")

	if len(args) != 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}

	// ok, err := t.isAdmin(stub)
	// if !ok {
	// 	return nil, err
	// }

	fundName := args[0]
	partnerAssets, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("partner assets is not int64")
	}
	partnerTime, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, errors.New("partner time is not int64")
	}
	buyStart, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return nil, errors.New("buy start is not int64")
	}
	buyPer, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return nil, errors.New("buy per is not int64")
	}
	buyAll, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return nil, errors.New("buy all is not int64")
	}

	_, row, err := getFundInfoByName(stub, fundName)
	if err != nil {
		return nil, err
	}

	row.Columns[3].Value = &shim.Column_Int64{Int64: partnerAssets}
	row.Columns[4].Value = &shim.Column_Int64{Int64: partnerTime}
	row.Columns[5].Value = &shim.Column_Int64{Int64: buyStart}
	row.Columns[6].Value = &shim.Column_Int64{Int64: buyPer}
	row.Columns[7].Value = &shim.Column_Int64{Int64: buyAll}

	_, err = stub.ReplaceRow("FundInfo", *row)
	if err != nil {
		return nil, errors.New("update fund limit failed:" + err.Error())
	}

	myLogger.Debug("setFundLimit done.")
	return nil, nil
}

//设置基金池（扩股回购）
func (t *FundManagementChaincode) setFundPool(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debug("setFundPool.....")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	// ok, err := t.isAdmin(stub)
	// if !ok {
	// 	return nil, err
	// }

	fundName := args[0]
	fundCount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("Fund is not int64")
	}

	_, row, err := getFundInfoByName(stub, fundName)
	if err != nil {
		return nil, err
	}

	funds := row.Columns[1].GetInt64() + fundCount
	if funds < 0 {
		//回购不足
		return nil, errors.New("回购失败，可回购数不足")
	}

	row.Columns[1].Value = &shim.Column_Int64{Int64: funds}
	_, err = stub.ReplaceRow("FundInfo", *row)
	if err != nil {
		myLogger.Errorf("update fund pool failed:%s", err)
		return nil, fmt.Errorf("update fund pool failed:%s", err)
	}

	myLogger.Debug("setFundPool done.")
	return nil, nil
}

//交易基金（认购赎回）
func (t *FundManagementChaincode) transferFund(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debug("transferFund.....")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	fundName := args[0]
	fundCount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("Fund count is not int64")
	}

	_, fundInfRow, err := getFundInfoByName(stub, fundName)
	if err != nil {
		return nil, err
	}

	owner, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, fmt.Errorf("Get caller certificate failed:%s", err)
	}

	_, _, err = getUserInfo(stub, fundName, owner)
	if err != nil {
		return nil, err
	}

	//验证限制是否满足

	sysFunds := fundInfRow.Columns[1].GetInt64() - fundCount
	sysAsset := fundInfRow.Columns[2].GetInt64() + fundCount*fundInfRow.Columns[8].GetInt64()

	userFunds := 0 //userRow.Columns[3].GetInt64() + fundCount
	userAsset := 0 //userRow.Columns[2].GetInt64() - fundCount*fundInfRow.Columns[8].GetInt64()
	if fundCount > 0 {
		//认购
		if sysFunds < 0 || userAsset < 0 {
			return nil, errors.New("认购失败，系统基金不租或者用户资金不足")
		}
	} else {
		if sysAsset < 0 || userFunds < 0 {
			return nil, errors.New("赎回失败，系统资金不足或者赎回数量超出用户基金数")
		}
	}

	fundInfRow.Columns[1].Value = &shim.Column_Int64{Int64: sysFunds}
	fundInfRow.Columns[2].Value = &shim.Column_Int64{Int64: sysAsset}
	_, err = stub.ReplaceRow("FundInfo", *fundInfRow)
	if err != nil {
		myLogger.Errorf("failed update fundinfo:%s", err)
		return nil, fmt.Errorf("failed update fundinfo:%s", err)
	}

	// userRow.Columns[2].Value = &shim.Column_Int64{Int64: userAsset}
	// userRow.Columns[3].Value = &shim.Column_Int64{Int64: userFunds}
	// _, err = stub.ReplaceRow("AccountFund", *userRow)
	// if err != nil {
	// 	myLogger.Errorf("failed update user fund info:%s", err)
	// 	return nil, fmt.Errorf("failed update user fund info:%s", err)
	// }

	myLogger.Debug("transferFund done.")
	return nil, nil
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

func getFundInfoByName(stub shim.ChaincodeStubInterface, fundName string) (*fundInfo, *shim.Row, error) {
	columns := []shim.Column{shim.Column{Value: &shim.Column_String_{String_: fundName}}}
	row, err := stub.GetRow("FundInfo", columns)
	if err != nil {
		myLogger.Errorf("Failed retrieving fundInfo [%s]: [%s]", fundName, err)
		return nil, nil, fmt.Errorf("Failed retrieving fundInfo [%s]: [%s]", fundName, err)
	}

	fundInfo := new(fundInfo)
	fundInfo.Name = row.Columns[0].GetString_()
	fundInfo.Funds = row.Columns[1].GetInt64()
	fundInfo.Assets = row.Columns[2].GetInt64()
	fundInfo.PartnerAssets = row.Columns[3].GetInt64()
	fundInfo.PartnerTime = row.Columns[4].GetInt64()
	fundInfo.BuyStart = row.Columns[5].GetInt64()
	fundInfo.BuyPer = row.Columns[6].GetInt64()
	fundInfo.BuyAll = row.Columns[7].GetInt64()
	fundInfo.Net = row.Columns[8].GetInt64()

	return fundInfo, &row, nil
}

func getFundInfoList(stub shim.ChaincodeStubInterface) ([]fundInfo, error) {
	return nil, nil
}

type userInfo struct {
	Name   string `json:"name,omitempty"`
	Owner  []byte `json:"owner,omitempty"`
	Assets int64  `json:"assets,omitempty"`
	Fund   int64  `json:"fund,omitempty"`
}

func getUserInfo(stub shim.ChaincodeStubInterface, fundName string, userCert []byte) (*userInfo, *shim.Row, error) {
	columns := []shim.Column{
		shim.Column{Value: &shim.Column_String_{String_: fundName}},
		shim.Column{Value: &shim.Column_Bytes{Bytes: userCert}},
	}

	row, err := stub.GetRow("AccountFund", columns)
	if err != nil {
		myLogger.Errorf("Failed retrieving account fundInfo [%s]: [%s]", fundName, err)
		return nil, nil, fmt.Errorf("Failed retrieving account fundInfo [%s]: [%s]", fundName, err)
	}

	userInfo := new(userInfo)
	userInfo.Name = row.Columns[0].GetString_()
	userInfo.Owner = row.Columns[1].GetBytes()
	userInfo.Assets = row.Columns[2].GetInt64()
	userInfo.Fund = row.Columns[3].GetInt64()

	return userInfo, &row, nil
}

//查询基金信息
func (t *FundManagementChaincode) queryFundInfo(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debug("query fund info....")

	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. ")
	}

	queryType := args[0]
	if queryType == "one" {
		fundName := args[1]
		info, _, err := getFundInfoByName(stub, fundName)
		if err != nil {
			return nil, err
		}
		return json.Marshal(info)
	} else if queryType == "list" {
		infos, err := getFundInfoList(stub)
		if err != nil {
			return nil, err
		}

		list := struct {
			List []fundInfo `json:"list,omitempty"`
		}{List: infos}

		return json.Marshal(&list)
	}
	return nil, nil
}

//查询用户信息
func (t *FundManagementChaincode) queryUserInfo(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	myLogger.Debug("query user info....")

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	user, err := stub.GetCallerCertificate()
	if err != nil {
		myLogger.Errorf("Get caller certificate failed:%s", err)
		return nil, fmt.Errorf("Get caller certificate failed:%s", err)
	}

	fundName := args[0]
	info, _, err := getUserInfo(stub, fundName, user)
	if err != nil {
		return nil, err
	}
	return json.Marshal(info)
}

func main() {
	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(FundManagementChaincode))
	if err != nil {
		fmt.Printf("Error starting FundManagementChaincode: %s", err)
	}
}
