package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

var timeInSeconds = map[string]int{
	"SECOND": 1,
	"MINUTE": 1 * 60,
	"HOUR":   1 * 60 * 60,
	"DAY":    1 * 60 * 60 * 24,
	"WEEK":   1 * 60 * 60 * 24 * 7,
	"MONTH":  1 * 60 * 60 * 24 * 7 * 30,
}

type SmartContract struct {
	contractapi.Contract
}

type Party struct {
	Name  string `json:"name"`
	MSPID string `json:"mspid"`
	Aware bool   `json:"aware"`
}

type Parties struct {
	Application Party `json:"application"`
	Process     Party `json:"process"`
}

type Interval struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Timeout struct {
	Increase int `json:"increase"`
	End      int `json:"end"`
}

type MaxNumberOfOperation struct {
	Max      int    `json:"max"`
	Used     int    `json:"used"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	TimeUnit string `json:"timeUnit"`
}

type RightRequestPayment0 struct {
	MaxNumberOfOperation0 MaxNumberOfOperation `json:"maxNumberOfOperation0"`

	MessageContent0 string `json:"messageContent0"`

	MessageContent1 string `json:"messageContent1"`
}

type ObligationResponsePayment1 struct {
	Timeout0 Timeout `json:"timeout0"`
}

type Asset struct {
	Parties     Parties `json:"parties"`
	IsActivated bool    `json:"isActivated"`

	RightRequestPayment0 RightRequestPayment0 `json:"rightRequestPayment0"`

	ObligationResponsePayment1 ObligationResponsePayment1 `json:"obligationResponsePayment1"`
}

func (s *SmartContract) IsParty(MSPID string, asset *Asset) bool {
	return MSPID == asset.Parties.Process.MSPID || MSPID == asset.Parties.Application.MSPID
}

func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface, parties Parties) (string, error) {

	if parties.Application.MSPID == "" {
		return "", fmt.Errorf("the MSPID from Application is required")
	}

	if parties.Process.MSPID == "" {
		return "", fmt.Errorf("the MSPID from process is required")
	}

	parties.Application.Aware = false
	parties.Process.Aware = false

	contractId := uuid.New()

	contract := Asset{}
	contract.Parties = parties

	contract.RightRequestPayment0.MaxNumberOfOperation0.Used = 0
	contract.RightRequestPayment0.MaxNumberOfOperation0.Start = 0
	contract.RightRequestPayment0.MaxNumberOfOperation0.End = 0

	contractAsBytes, err := json.Marshal(contract)

	if err != nil {
		return "", fmt.Errorf("marshal error: %s", err.Error())
	}

	ctx.GetStub().PutState(contractId.String(), contractAsBytes)

	return contractId.String(), nil
}

func (s *SmartContract) Sign(ctx contractapi.TransactionContextInterface, contractId string) error {

	contractAsBytes, err := ctx.GetStub().GetState(contractId)

	if err != nil {
		return fmt.Errorf("failed to read from state: %s", err.Error())
	}

	if contractAsBytes == nil {
		return fmt.Errorf("contract %s does not exist", contractId)
	}

	contract := new(Asset)

	MSPID, err := cid.GetMSPID(ctx.GetStub())

	if err != nil {
		return fmt.Errorf("fail to get MSPID")
	}

	err = json.Unmarshal(contractAsBytes, contract)

	if err != nil {
		return fmt.Errorf("marshal error: %s", err.Error())
	}

	if !s.IsParty(MSPID, contract) {
		return fmt.Errorf("only the process or the application can execute this operation")
	}

	if contract.Parties.Application.MSPID == MSPID {

		if contract.Parties.Application.Aware {
			return fmt.Errorf("the contract is already signed")
		}

		contract.Parties.Application.Aware = true
	}

	if contract.Parties.Process.MSPID == MSPID {

		if contract.Parties.Process.Aware {
			return fmt.Errorf("the contract is already signed")
		}

		contract.Parties.Process.Aware = true
	}

	contract.IsActivated = contract.Parties.Application.Aware && contract.Parties.Process.Aware

	return nil
}

func (s *SmartContract) Query(ctx contractapi.TransactionContextInterface, contractId string) (*Asset, error) {

	contractAsBytes, err := ctx.GetStub().GetState(contractId)

	if err != nil {
		return nil, fmt.Errorf("failed to read from state: %s", err.Error())
	}

	if contractAsBytes == nil {
		return nil, fmt.Errorf("contract %s does not exist", contractId)
	}

	contract := new(Asset)

	err = json.Unmarshal(contractAsBytes, contract)

	if err != nil {
		return nil, fmt.Errorf("marshal error: %s", err.Error())
	}

	return contract, nil
}

func (s *SmartContract) ClauseRightRequestPayment0(ctx contractapi.TransactionContextInterface, contractId string,

	messageContent0 string,

	messageContent1 string,

	accessDateTime int,

) (bool, error) {

	contract, err := s.Query(ctx, contractId)

	if err != nil {
		return false, err
	}

	isValid := true

	maxNumberOfOperationIsInitialized0 := contract.RightRequestPayment0.MaxNumberOfOperation0.Start == 0 && contract.RightRequestPayment0.MaxNumberOfOperation0.End == 0

	endPeriodIsLassThanAccessDateTime0 := contract.RightRequestPayment0.MaxNumberOfOperation0.End < accessDateTime

	if !maxNumberOfOperationIsInitialized0 || endPeriodIsLassThanAccessDateTime0 {
		contract.RightRequestPayment0.MaxNumberOfOperation0.Start = accessDateTime
		contract.RightRequestPayment0.MaxNumberOfOperation0.End = accessDateTime + timeInSeconds[contract.RightRequestPayment0.MaxNumberOfOperation0.TimeUnit]
		contract.RightRequestPayment0.MaxNumberOfOperation0.Used = 0
	}

	isValid = isValid && contract.RightRequestPayment0.MaxNumberOfOperation0.Used <= contract.RightRequestPayment0.MaxNumberOfOperation0.Max

	isValid = isValid && contract.RightRequestPayment0.MessageContent0 <= messageContent0

	isValid = isValid && contract.RightRequestPayment0.MessageContent1 <= messageContent1

	contract.ObligationResponsePayment1.Timeout0.End = accessDateTime + contract.ObligationResponsePayment1.Timeout0.Increase

	if !isValid {

		return isValid, fmt.Errorf("error executing clause: RightRequestPayment0")

	}

	return isValid, nil
}

func (s *SmartContract) ClauseObligationResponsePayment1(ctx contractapi.TransactionContextInterface, contractId string,

	accessDateTime int,

) (bool, error) {

	contract, err := s.Query(ctx, contractId)

	if err != nil {
		return false, err
	}

	isValid := true

	isValid = isValid && accessDateTime <= contract.ObligationResponsePayment1.Timeout0.End

	if !isValid {

		return isValid, fmt.Errorf("error executing clause: ObligationResponsePayment1")

	}

	return isValid, nil
}

func main() {
	chainconde, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Errorf("error create chaincode: %s", err.Error())
		return
	}

	if err := chainconde.Start(); err != nil {
		fmt.Errorf("error create chaincode: %s", err.Error())
	}
}
