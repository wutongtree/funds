package main

import (
	"fmt"

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

	myLogger.Debug("Init Chaincode...done")

	return nil, nil
}

// Invoke will be called for every transaction.
func (t *FundManagementChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

// Query callback representing the query of a chaincode
// Anyone can invoke this function.
func (t *FundManagementChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	return nil, nil
}

func main() {
	primitives.SetSecurityLevel("SHA3", 256)
	err := shim.Start(new(FundManagementChaincode))
	if err != nil {
		fmt.Printf("Error starting FundManagementChaincode: %s", err)
	}
}
