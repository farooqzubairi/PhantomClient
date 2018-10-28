package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type PhantomChaincode struct {
}

//Asset - Phantom Account
type PhantomAccount struct {
	ObjectType              string  `json:"docType"`                 // Should be PHANTACNT
	AccountIdentifier       string  `json:"accountIdentifier"`       // Party1 BIC + Party2 BIC + Currency+ XX
	Party1                  string  `json:"party1"`                  // 0. Joint holder bank's BIC
	Party2                  string  `json:"party2"`                  // 1. Joint Holder bank's BIC
	CustodianBIC            string  `json:"custodianAccount"`        // 2. Bank's BIC, operating the bank account
	CorrespondentBIC        string  `json:"correspondentBIC"`        // 3. To do
	AccountCurrency         string  `json:"accountCurrency"`         // 4. Account currency
	BankBalance             float64 `json:"accountBalance"`          // 5. Bank balance
	AccountStatus           string  `json:"accountStatus"`           // 6. Account operation Status
	MaxBalLimit             float64 `json:"maxBalLimit"`             // 7. Upward threshold limit
	MinBalLimit             float64 `json:"minBalLimit"`             // 8. Downward threshold limit
	SettlementSuffix        string  `json:"settlementSuffix"`        // 9. Settlement suffix
	TotalBalInRecvdState    float64 `json:"totalBalInRecvdState"`    // 
	TotalBalInApprovedState float64 `json:"totalBalInApprovedState"` // 
	TotalBalInCompleteState float64 `json:"totalBalInCompleteState"` // 
	//DataExtension
}

"ARABJO","ARABAE","ARABAE","","AED","1234","ACTIVE","123","222","01"

// Asset - Fund Transfer Order
type FundTransferOrder struct {
	ObjectType             string    `json:"docType"`                // Should be FTO
	TransactionId          string    `json:"transactionId"`          // Unique transactionId sender BIC + Receiver BIC + Payment RefOrder
	TransferType           string    `json:"transferType"`           // 0.Transfer Type
	PaymentRefNum          string    `json:"paymentRefNum"`          // 1.Payment ref number
	SenderBankBIC          string    `json:"senderBankBIC"`          // .Sender Bank BIC -- get from ESC certidicate
	ReceiverBankBIC        string    `json:"receiverBankBIC"`        // 2.
	TransactionCurrency    string    `json:"transactionCurrency"`    // 3.
	SettlementSuffix       string    `json:"settlementSuffix"`       // 4.
	PhantomAccountId       string    `json:"phantomAccountId"`       // CC will identify on the basis of sender+receiver+currency+settelemetSuffic
	TransferDate           time.Time `json:"transferDate"`           // 5.
	RequestType            string    `json:"requestType"`            // 6.PHANTOM Transfer/SWIFT Transfer/Phantom Settlement/Charges Settlement
	Narrative              string    `json:"narrative"`              // 7.
	TransferState          string    `json:"transferState"`          // INITIATED, if reqType is SWIFT Transfer
	SecurityToken          string    `json:"securityToken"`          // 8
	TransactionHashKey     string    `json:"transactionHashKey"`     // 9
	ValueDate              time.Time `json:"valueDate"`              //10
	FundTransferAmount     float64   `json:"fundTransferAmount"`     // 11 Fund transfer value
	IsLiabilityToCustodian bool      `json:"isLiabilityToCustodian"` // This has internal use. True if fundtransfer amount is a liability on custodian
	//DataExtension       map[string]
}

// Response message struct
type Response struct {
	Code    string // status code 11 - Successs; 00- Failure
	message string // response message
}

//Enum for transaction statuses
type transactionStatus string

const (
	Initiated transactionStatus = "INITIATED" // Sender bank initiate the transaction
	Receieved transactionStatus = "RECEIVED"  // Receiver Bank Acknowledged the transaction
	Approved  transactionStatus = "APPROVED"  // Reciever Bank Approve the transaction
	Rejected  transactionStatus = "REJECTED"  // Reciver Bak reject the transaction
	Completed transactionStatus = "COMPLETED" //Reciver bank complete the transaction

)

// Defining Map for TO Request Type
var requestTypeMap map[string]string
var transferTypeMap map[string]string
var phantomAccountStatusMap map[string]string

//Init chaincode
func (t *PhantomChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Phantom Chaincode is bootstrapping ...")
	transferTypeMap = make(map[string]string)
	transferTypeMap["ST"] = "SETTLEMENT"
	transferTypeMap["MT"] = "MONEY TRANSFER"
	transferTypeMap["FT"] = "FUND TRANSFER"
	transferTypeMap["CT"] = "FEE - ChHARGES"

	phantomAccountStatusMap = make(map[string]string)
	phantomAccountStatusMap["ACTIVE"] = "ACTIVE"
	phantomAccountStatusMap["DORMANT"] = "DORMANT"
	return shim.Success(getResponseMessage("11", "Phantom_cc started sucessfully"))
}

// Invoke
func (t *PhantomChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("starting invoke, for - " + function)
	fmt.Printf("%d\n", len(args))
	if function == "Init" {
		return t.Init(stub)
	} else if function == "ADDACCOUNT" {
		return t.AddBankAccount(stub, args)
	} else if function == "Test" {
		return t.Test(stub, args)
	} else if function == "GETASSET" {
		return t.GetAssetDetails(stub, args)
	}
	return shim.Error(string(getResponseMessage("00", "Unknown function invocation: "+function)))
}

// To Do add three function to change accountStatus, max/min limit -- add event for change

// Add Bank account to World State  -- By treasury (Operator) --technical task
func (t *PhantomChaincode) AddBankAccount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	if len(args[0]) == 0 {
		return shim.Error("Incorrect length for party 1 BIC. Expecting XXXXXX")
	}

	if len(args[1]) == 0 {
		return shim.Error("Incorrect length for party 2 BIC. Expecting XXXXXX")
	}

	if len(args[2]) == 0 {
		return shim.Error("Incorrect length for custodian Bank BIC. Expecting XXXXXX")
	}

	if len(args[3]) == 0 {
		//return shim.Error("Incorrect length for correspondent bank BIC. Expecting XXXXXX")
	}
	if len(args[4]) == 0 {
		return shim.Error("Incorrect account currency. Expecting XXX")
	}
	if len(args[5]) == 0 {
		return shim.Error("Empty Bank Balance.")
	}
	if len(args[6]) == 0 {
		return shim.Error("Empty account status. Expecting ACTIVE or DORMANT")
	}
	_, ok := phantomAccountStatusMap[args[6]]
	if !ok {
		return shim.Error("Unsupported account status")
	}

	if len(args[7]) == 0 {
		return shim.Error("Incorrect max balance limit")
	}
	if len(args[8]) == 0 {
		return shim.Error("Incorrect min balance limit")
	}
	// setllement suffix
	if len(args[9]) == 0 {
		return shim.Error("Incorrect length for settlement suffix. Expecting XX")
	}

	objType := "PHANTACNT"
	party1 := args[0]
	party2 := args[1]
	if party1 == party2 {
		return shim.Error("Party1 BIC can not be equal to Party2 BIC")
	}
	custodianBIC := args[2]
	correspondentBIC := args[3]
	accountCurrency := args[4]
	bankBalance, err := strconv.ParseFloat(args[5], 64)
	accountStatus := args[6]
	maxBalLimit, _ := strconv.ParseFloat(args[7], 64)
	minBalLimit, _ := strconv.ParseFloat(args[8], 64)
	settlementSuffix := args[9]

	//party1 + "-" + party2 + "-" + accountCurrency + "-" + settlementSuffix
	var accountIdentifierSb strings.Builder
	accountIdentifierSb.WriteString(party1)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(party2)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(accountCurrency)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(settlementSuffix)
	accountIdentifier := accountIdentifierSb.String()

	// Check if Bank Account already added
	bankAccountAsBytes, err := stub.GetState(accountIdentifier)
	if err != nil {
		return shim.Error(" Existing bank account check failed.")
	} else if bankAccountAsBytes != nil {
		return shim.Error(accountIdentifier + " already exists.")
	}
	// other alternate combination
	accountIdentifierSb.Reset()
	accountIdentifierSb.WriteString(party2)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(party1)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(accountCurrency)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(settlementSuffix)
	accountIdentifier = accountIdentifierSb.String()
	bankAccountAsBytes, err = stub.GetState(accountIdentifierSb.String())
	if err != nil {
		return shim.Error(" Existing bank account check failed.")
	} else if bankAccountAsBytes != nil {
		return shim.Error(accountIdentifier + " already exists.")
	}

	bankAccount := PhantomAccount{ObjectType: objType, AccountIdentifier: accountIdentifier, Party1: party1, Party2: party2, CustodianBIC: custodianBIC, CorrespondentBIC: correspondentBIC, AccountCurrency: accountCurrency, BankBalance: bankBalance, AccountStatus: accountStatus, MaxBalLimit: maxBalLimit, MinBalLimit: minBalLimit, SettlementSuffix: settlementSuffix}

	bankAccountJSONBytes, err := json.Marshal(bankAccount)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(accountIdentifier, bankAccountJSONBytes)

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("Phantom account added :- " + accountIdentifier + " with account status " + accountStatus))
}

// Submit transaction -- only one function for all state change
func (t *PhantomChaincode) SubmitFundTransferOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 12")
	}
	if len(args[0]) == 0 {
		return shim.Error("Empty for transfer type.")
	}
	_, ok := requestTypeMap[args[0]]
	if !ok {
		return shim.Error("Unsupported Transfer Type")
	}
	if len(args[1]) == 0 {
		return shim.Error("Empty payment reference no")
	}
	if len(args[2]) == 0 {
		return shim.Error("Empty account receiver BIC. Expecting XXXXXX")
	}
	if len(args[3]) == 0 {
		return shim.Error("Empty transaction currency")
	}
	if len(args[4]) == 0 || len(args[4]) != 2 {
		return shim.Error(" Incorrect length for settlemet suffix. Expecting XX")
	}
	if len(args[5]) != 15 {
		return shim.Error("Transfer date should be in DDMMMYYYYHHMMSS format")
	}
	if len(args[6]) == 0 {
		return shim.Error("Empty Request type")
	}
	if len(args[7]) == 0 {
		return shim.Error("Empty narrative")
	}
	if len(args[8]) == 0 {
		return shim.Error("Empty security token")
	}
	if len(args[9]) == 0 {
		return shim.Error("Empty transaction hash")
	}
	if len(args[10]) != 15 {
		return shim.Error("Value date should be in DDMMMYYYYHHMMSS format")
	}
	if len(args[11]) <= 0 {
		return shim.Error("Fund transfer amount can not be zero or negative")
	}
	objType := "FTO"
	paymentRefNum := args[1]
	attr, ok, err := cid.GetAttributeValue(stub, "bic")
	if err != nil {
		return shim.Error("There was an error trying to retrieve the attribute bic")
	}
	if !ok {
		return shim.Error("The client identity does not possess the attribute bic")
	}
	senderBankBIC := attr
	receiverBankBIC := args[2]
	if senderBankBIC == receiverBankBIC {
		return shim.Error("Sender Bank BIC can not be equal to receiver bank BIC.")
	}

	fundTransferAmount, _ := strconv.ParseFloat(args[5], 64)
	transactionId := senderBankBIC + receiverBankBIC + paymentRefNum

	// Check if transaction id is unique
	transactionIdAsBytes, err := stub.GetState(transactionId)
	if err != nil {
		return shim.Error(" Unique transaction id check failed.")
	} else if transactionIdAsBytes != nil {
		return shim.Error(transactionId + " already exists.")
	}
	transferType := args[0]
	transactionCurrency := args[3]
	settlementSuffix := args[4]
	var accountIdentifierSb strings.Builder
	accountIdentifierSb.WriteString(senderBankBIC)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(receiverBankBIC)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(transactionCurrency)
	accountIdentifierSb.WriteString(" - ")
	accountIdentifierSb.WriteString(settlementSuffix)
	phantomAccountId := accountIdentifierSb.String()
	bankAccountAsBytes, err := stub.GetState(phantomAccountId)
	if err != nil {
		return shim.Error(" Existing bank account check failed.")
	} else if bankAccountAsBytes == nil {
		accountIdentifierSb.Reset()
		accountIdentifierSb.WriteString(receiverBankBIC)
		accountIdentifierSb.WriteString(" - ")
		accountIdentifierSb.WriteString(senderBankBIC)
		accountIdentifierSb.WriteString(" - ")
		accountIdentifierSb.WriteString(transactionCurrency)
		accountIdentifierSb.WriteString(" - ")
		accountIdentifierSb.WriteString(settlementSuffix)
		phantomAccountId = accountIdentifierSb.String()
		bankAccountAsBytes, _ := stub.GetState(phantomAccountId)
		if bankAccountAsBytes == nil {
			return shim.Error(phantomAccountId + " does not exists.")
		}
	}

	// Fetch Max/Min limit from Phantom account asset
	tmpPhantomAccount := PhantomAccount{}
	err = json.Unmarshal(bankAccountAsBytes, &tmpPhantomAccount)
	if err != nil {
		return shim.Error(err.Error())
	}
	maxBalLimit := tmpPhantomAccount.MaxBalLimit
	minBalLimit := tmpPhantomAccount.MinBalLimit
	custodianBankBIC := tmpPhantomAccount.CustodianBIC
	accountBalance := tmpPhantomAccount.BankBalance
	var isLiability bool

	amountToBeSettled := fundTransferAmount
	if senderBankBIC != custodianBankBIC {
		amountToBeSettled = -(fundTransferAmount)
		isLiability = true
	}

	if (accountBalance - amountToBeSettled) <= minBalLimit {
		return shim.Error("Minimum threshold limit breached")
	} else if (accountBalance + amountToBeSettled) >= maxBalLimit {
		return shim.Error("Maximum balance threshold breached")
	}

	transferDate, _ := time.Parse("2006-01-02T15:04:05", args[5]) // input args[5] should be in YYYY-MM-DDTHH:MM:SS format
	if transferDate.After(time.Now()) {
		return shim.Error("Transfer date " + transferDate.Format("2006-01-02T15:04:05") + "can not be in future. Current date: " + time.Now().Format("2006-01-02T15:04:05"))
	}

	valueDate, _ := time.Parse("2006-01-02T15:04:05", args[10]) // input args[10] should be in YYYY-MM-DDTHH:MM:SS format

	requestType := args[7]
	narrative := senderBankBIC + " - " + args[8]
	transferState := Initiated //from enum
	securityToken := args[9]
	transactionHashKey := args[10]

	fundTransferOrder := FundTransferOrder{ObjectType: objType, TransactionId: transactionId, TransferType: transferType, PaymentRefNum: paymentRefNum, SenderBankBIC: senderBankBIC, ReceiverBankBIC: receiverBankBIC, TransactionCurrency: transactionCurrency, SettlementSuffix: settlementSuffix, PhantomAccountId: phantomAccountId, TransferDate: transferDate, RequestType: requestType, Narrative: narrative, TransferState: string(transferState), SecurityToken: securityToken, TransactionHashKey: transactionHashKey, ValueDate: valueDate, IsLiabilityToCustodian: isLiability}
	fundTransferOrderAsJSONBytes, err := json.Marshal(fundTransferOrder)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(transactionId, fundTransferOrderAsJSONBytes)

	if err != nil {
		return shim.Error(err.Error())
	}

	// Update position in Phantom account
	tmpPhantomAccount.BankBalance = accountBalance + amountToBeSettled
	tmpPhantomAccountAsJSONBytes, err := json.Marshal(tmpPhantomAccount)

	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(phantomAccountId, tmpPhantomAccountAsJSONBytes)
	// if failed delete the FT
	if err != nil {
		stub.DelState(transactionId)
		return shim.Error(err.Error())
	}

	stub.SetEvent("TX_INITIATED", fundTransferOrderAsJSONBytes)
	return shim.Success([]byte("Fund transfer order with transaction Id " + transactionId + " " + string(Initiated) + " successfully"))
}

// Process Transfer Request
func (t *PhantomChaincode) ProcessTransferRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting transaction id and narrative")
	}
	if len(args[1]) == 0 {
		return shim.Error("Narrative can not empty")
	}
	transactionId := args[0]
	narrative := args[1] + "\n"
	// Check if transaction id exists
	trfOrderAsBytes, err := stub.GetState(transactionId)
	if err != nil {
		return shim.Error("Search for transaction id failed.")
	} else if trfOrderAsBytes == nil {
		return shim.Error(transactionId + " does not exists.")
	}

	updatedTrfOrder := FundTransferOrder{}
	err = json.Unmarshal(trfOrderAsBytes, &updatedTrfOrder) //unmarshal json
	if err != nil {
		return shim.Error(err.Error())
	}

	receiverBankBIC := updatedTrfOrder.ReceiverBankBIC
	senderBankBIC := updatedTrfOrder.SenderBankBIC
	transferState := updatedTrfOrder.TransferState
	isLiability := updatedTrfOrder.IsLiabilityToCustodian
	var eventName string

	// Fetch attribute value
	attr, ok, err := cid.GetAttributeValue(stub, "bic")
	if err != nil {
		return shim.Error("There was an error trying to retrieve the attribute bic")
	}
	if !ok {
		return shim.Error("The client identity does not possess the attribute bic")
	}
	// Fetch Phantom account details
	phantomAccountAsBytes, err := stub.GetState(updatedTrfOrder.PhantomAccountId)
	if err != nil {
		return shim.Error("Search for phantom account failed.")
	} else if phantomAccountAsBytes == nil {
		return shim.Error(updatedTrfOrder.PhantomAccountId + " phantom account does not exists.")
	}
	phantomAccount := PhantomAccount{}
	err = json.Unmarshal(phantomAccountAsBytes, &phantomAccount)
	if err != nil {
		return shim.Error(err.Error())
	}

	if attr == receiverBankBIC {
		updatedTrfOrder.Narrative = receiverBankBIC + " - " + narrative
		if transferState == string(Initiated) {
			if isLiability {
				phantomAccount.TotalBalInRecvdState = phantomAccount.TotalBalInRecvdState - updatedTrfOrder.FundTransferAmount
			} else {
				phantomAccount.TotalBalInRecvdState += updatedTrfOrder.FundTransferAmount
			}
			updatedTrfOrder.TransferState = string(Receieved)
			eventName = "TX_RECEIVED"

		} else if transferState == string(Receieved) {
			if isLiability {
				phantomAccount.TotalBalInRecvdState += updatedTrfOrder.FundTransferAmount
				phantomAccount.TotalBalInApprovedState -= updatedTrfOrder.FundTransferAmount
			} else {
				phantomAccount.TotalBalInRecvdState -= updatedTrfOrder.FundTransferAmount
				phantomAccount.TotalBalInApprovedState += updatedTrfOrder.FundTransferAmount
			}
			updatedTrfOrder.TransferState = string(Approved)
			eventName = "TX_APPROVED"
		} else {
			return shim.Error(receiverBankBIC + " is unauthoized to take action on " + transactionId + " in state " + transferState)
		}
	} else if attr == senderBankBIC {
		updatedTrfOrder.Narrative = senderBankBIC + " - " + narrative
		if transferState == string(Approved) || transferState == string(Rejected) {
			if isLiability {
				phantomAccount.TotalBalInApprovedState += updatedTrfOrder.FundTransferAmount
				phantomAccount.TotalBalInCompleteState -= updatedTrfOrder.FundTransferAmount
			} else {
				phantomAccount.TotalBalInApprovedState -= updatedTrfOrder.FundTransferAmount
				phantomAccount.TotalBalInCompleteState += updatedTrfOrder.FundTransferAmount
			}
			updatedTrfOrder.TransferState = string(Completed)
			eventName = "TX_COMPLETED"
			// ToDo complete logic
		} else {
			return shim.Error(senderBankBIC + " is unauthoized to take action on " + transactionId + " in state " + transferState)
		}
	}

	updatedTrfOrderAsBytes, err := json.Marshal(updatedTrfOrder)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(transactionId, updatedTrfOrderAsBytes)

	if err != nil {
		return shim.Error(err.Error())
	}
	updatedPhantomAccountasBytes, err := json.Marshal(phantomAccount)
	if err != nil {
		err = stub.PutState(transactionId, trfOrderAsBytes)
		return shim.Error(err.Error())
	}
	err = stub.PutState(phantomAccount.AccountIdentifier, updatedPhantomAccountasBytes)

	if err != nil {
		err = stub.PutState(transactionId, trfOrderAsBytes)
		return shim.Error(err.Error())
	}

	stub.SetEvent(eventName, updatedTrfOrderAsBytes)
	return shim.Success([]byte(transactionId + " succesfully " + updatedTrfOrder.TransferState))
}

// Reject Transfer Order
func (t *PhantomChaincode) RejectTO(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting transaction id and narrative")
	}
	if len(args[1]) == 0 {
		return shim.Error("Narrative can not empty")
	}
	transactionId := args[0]
	narrative := args[1] + "\n"

	// Check if transaction id is unique
	orderAsBytes, err := stub.GetState(transactionId)
	if err != nil {
		return shim.Error(" Search for transaction id failed.")
	} else if orderAsBytes == nil {
		return shim.Error(transactionId + " does not exists.")
	}
	// Fetch attribute value
	attr, ok, err := cid.GetAttributeValue(stub, "bic")
	if err != nil {
		return shim.Error("There was an error trying to retrieve the attribute bic")
	}
	if !ok {
		return shim.Error("The client identity does not possess the attribute bic")
	}
	newOrder := FundTransferOrder{}
	err = json.Unmarshal(orderAsBytes, &newOrder) //unmarshal json
	if err != nil {
		return shim.Error(err.Error())
	}
	if attr != newOrder.ReceiverBankBIC {
		return shim.Error(attr + " is not authorized to reject the transaction")
	}
	newOrder.TransferState = string(Rejected)
	newOrder.Narrative = narrative
	fundTransferOrderAsJSONBytes, err := json.Marshal(newOrder)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(transactionId, fundTransferOrderAsJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	stub.SetEvent("TX_REJECTED", fundTransferOrderAsJSONBytes)
	return shim.Success([]byte("Fund transfer order with transaction Id " + transactionId + " " + string(Rejected) + " successfully"))
}

// To Test something
func (t *PhantomChaincode) Test(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	id, err := cid.GetID(stub)
	fmt.Printf("The client ide- %s", id)
	mspid, err := cid.GetMSPID(stub)
	fmt.Printf("The MSP ide- %s", mspid)
	val, ok, err := cid.GetAttributeValue(stub, "bic")
	if err != nil {
		fmt.Printf("There was an error trying to retrieve the attribute - %s", err)
	}
	if !ok {
		fmt.Printf("The client identity does not possess the attribute- %s", err)
	}
	fmt.Printf("The client ide- %s", val)
	// creator, _ := stub.GetCreator()
	//fmt.Printf("creator is::::::::::::::: - %s", creator)
	return shim.Success([]byte(val))
}
func main() {
	err := shim.Start(new(PhantomChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}
