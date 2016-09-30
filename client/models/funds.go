package models

import (
	"encoding/json"
	"fmt"
)

type Fund struct {
	Id             string  `json:"Id,omitempty"`
	Name           string  `json:"Name,omitempty"`
	CreatTime      string  `json:"CreatTime,omitempty"`
	Quotas         float64 `json:"Quotas,omitempty"`
	MarketValue    float64 `json:"MarketValue,omitempty"`
	NetValue       float64 `json:"NetValue,omitempty"`
	NetDelta       string  `json:"NetDelta,omitempty"`
	ThresholdValue float64 `json:"ThresholdValue,omitempty"`
}

type MyFund struct {
	Fund
	MyQuotas      float64 `json:"MyQuotas,omitempty"`
	MyMarketValue float64 `json:"MyQuotas,omitempty"`
	MyBalance     float64 `json:"MyBalance,omitempty"`
}

type FundMarket struct {
	Index int
	Size  float64
	Type  string
}

type FundNotice struct {
	Title       string
	PublishTime string
}

// ---------- struct with app ------------
// getMyFundResponse getMyFundResponse
type getMyFundResponse struct {
	Name   string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Owner  string `protobuf:"bytes,2,opt,name=owner" json:"owner,omitempty"`
	Assets string `protobuf:"bytes,3,opt,name=assets" json:"assets,omitempty"`
	Fund   string `protobuf:"bytes,4,opt,name=fund" json:"fund,omitempty"`
}

// AppFund AppFund
type AppFund struct {
	Name          string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Funds         int    `protobuf:"bytes,1,opt,name=funds" json:"funds,omitempty"`
	Assets        int    `protobuf:"bytes,1,opt,name=assets" json:"assets,omitempty"`
	PartnerAssets int    `protobuf:"bytes,1,opt,name=partnerAssets" json:"partnerAssets,omitempty"`
	PartnerTime   int    `protobuf:"bytes,1,opt,name=partnerTime" json:"partnerTime,omitempty"`
	BuyStart      int    `protobuf:"bytes,1,opt,name=buyStart" json:"buyStart,omitempty"`
	BuyPer        int    `protobuf:"bytes,1,opt,name=buyPer" json:"buyPer,omitempty"`
	BuyAll        int    `protobuf:"bytes,1,opt,name=buyAll" json:"buyAll,omitempty"`
	Net           int    `protobuf:"bytes,1,opt,name=net" json:"net,omitempty"`
}

// AppFundsResponse AppFundsResponse
type AppFundsResponse struct {
	Status string    `json:"status,omitempty"`
	Msg    []AppFund `json:"msg,omitempty"`
}

// AppFundResponse AppFundResponse
type AppFundResponse struct {
	Status string  `json:"status,omitempty"`
	Msg    AppFund `json:"msg,omitempty"`
}

// AppMyFund AppMyFund
type AppMyFund struct {
	Name   string `json:"Name,omitempty"`
	Owner  string `json:"owner,omitempty"`
	Assets int    `json:"assets,omitempty"`
	Fund   int    `json:"fund,omitempty"`
}

// AppMyFundResponse AppMyFundResponse
type AppMyFundResponse struct {
	Status string    `json:"status,omitempty"`
	Msg    AppMyFund `json:"msg,omitempty"`
}

// AppTransfterFundRequest AppTransfterFundRequest
type AppTransfterFundRequest struct {
	EnrollID string `json:"enrollID,omitempty"`
	Name     string `json:"name,omitempty"`
	Funds    int    `json:"funds,omitempty"`
}

// AppTransfterFundResponse AppTransfterFundResponse
type AppTransfterFundResponse struct {
	Status string `json:"status,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

// AppCreateFundRequest AppCreateFundRequest
type AppCreateFundRequest struct {
	EnrollID      string `json:"enrollID,omitempty"`
	Name          string `json:"name,omitempty"`
	Funds         int    `json:"funds,omitempty"`
	Assets        int    `json:"assets,omitempty"`
	PartnerAssets int    `json:"partnerAssets,omitempty"`
	PartnerTime   int    `json:"partnerTime,omitempty"`
	BuyStart      int    `json:"buyStart,omitempty"`
	BuyPer        int    `json:"buyPer,omitempty"`
	BuyAll        int    `json:"buyAll,omitempty"`
	Netvalue      int    `json:"net,omitempty"`
}

// AppCreateFundResponse AppCreateFundResponse
type AppCreateFundResponse struct {
	Status string `json:"status,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

// AppSetFundNetvalueRequest AppSetFundNetvalueRequest
type AppSetFundNetvalueRequest struct {
	EnrollID string `json:"enrollID,omitempty"`
	Name     string `json:"name,omitempty"`
	Netvalue int    `json:"net,omitempty"`
}

// AppSetFundNetvalueResponse AppSetFundNetvalueResponse
type AppSetFundNetvalueResponse struct {
	Status string `json:"status,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

// AppSetFundThreshholdRequest AppSetFundThreshholdRequest
type AppSetFundThreshholdRequest struct {
	EnrollID      string `json:"enrollID,omitempty"`
	Name          string `json:"name,omitempty"`
	PartnerAssets int    `json:"partnerAssets,omitempty"`
	PartnerTime   int    `json:"partnerTime,omitempty"`
	BuyStart      int    `json:"buyStart,omitempty"`
	BuyPer        int    `json:"buyPer,omitempty"`
	BuyAll        int    `json:"buyAll,omitempty"`
}

// AppSetFundThreshholdResponse AppSetFundThreshholdResponse
type AppSetFundThreshholdResponse struct {
	Status string `json:"status,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

// ListMyFunds ListMyFunds
func ListMyFunds(userId string, page int, offset int) (nums int, funds []Fund, err error) {
	err = nil

	// Get fund
	urlstr := getHTTPURL("funds")
	response, err := performHTTPGet(urlstr)
	if err != nil {
		logger.Errorf("ListMyFunds failed: %v", err)
		return
	}

	logger.Debugf("ListMyFunds: url=%v response=%v", urlstr, string(response))

	var result AppFundsResponse
	err = json.Unmarshal(response, &result)
	if err != nil {
		logger.Errorf("ListMyFunds failed: %v", err)
		return
	}

	if result.Status != "OK" {
		logger.Errorf("ListMyFunds failed: %v", result.Status)
		return
	}

	// result
	for _, v := range result.Msg {
		fund := Fund{
			Id:          v.Name,
			Name:        v.Name,
			CreatTime:   "2016-09-19",
			Quotas:      float64(v.Funds),
			MarketValue: float64(v.Funds * v.Net),
			NetValue:    float64(v.Net),
		}

		funds = append(funds, fund)
	}

	nums = len(funds)
	return nums, funds, err
}

// GetMyFund GetMyFund
func GetMyFund(userId string, fundid string) (myfund MyFund, err error) {
	err = nil

	// Get fund
	urlstr := getHTTPURL("fund/" + fundid)
	response, err := performHTTPGet(urlstr)
	if err != nil {
		logger.Errorf("GetMyFund failed: %v", err)
		return
	}

	logger.Debugf("GetMyFund: url=%v response=%v", urlstr, string(response))

	var resultAppFund AppFundResponse
	err = json.Unmarshal(response, &resultAppFund)
	if err != nil {
		logger.Errorf("GetMyFund failed: %v", err)
		return
	}

	if resultAppFund.Status != "OK" {
		logger.Errorf("GetMyFund failed: %v", resultAppFund.Status)
		return
	}

	// Get My fund
	urlstr = getHTTPURL("user/" + fundid + "/" + userId)
	response, err = performHTTPGet(urlstr)
	if err != nil {
		logger.Errorf("GetMyFund failed: %v", err)
		return
	}

	logger.Debugf("GetMyFund: url=%v response=%v", urlstr, string(response))

	myfund = MyFund{
		MyQuotas:      0,
		MyMarketValue: 0,
		MyBalance:     0,
	}

	var resultAppMyFund AppMyFundResponse
	err = json.Unmarshal(response, &resultAppMyFund)
	if err != nil {
		logger.Errorf("GetMyFund failed: %v", err)
	} else {
		myfund.MyQuotas = float64(resultAppMyFund.Msg.Fund)
		myfund.MyMarketValue = float64(resultAppMyFund.Msg.Fund * resultAppFund.Msg.Net)
		myfund.MyBalance = float64(resultAppMyFund.Msg.Assets)
	}

	myfund.Id = resultAppFund.Msg.Name
	myfund.Name = resultAppFund.Msg.Name
	myfund.CreatTime = "2016-09-19"
	myfund.Quotas = float64(resultAppFund.Msg.Funds)
	myfund.MarketValue = float64(resultAppFund.Msg.Net * resultAppFund.Msg.Funds)
	myfund.NetValue = float64(resultAppFund.Msg.Net)
	myfund.NetDelta = "+0.001|0.94%"
	myfund.ThresholdValue = float64(resultAppFund.Msg.Net * resultAppFund.Msg.BuyPer)

	return
}

// GetFund GetFund
func GetFund(userId string, fundid string) (fund AppFund, err error) {
	err = nil

	// Get fund
	urlstr := getHTTPURL("fund/" + fundid)
	response, err := performHTTPGet(urlstr)
	if err != nil {
		logger.Errorf("GetFund failed: %v", err)
		return
	}

	logger.Debugf("GetFund: url=%v response=%v", urlstr, string(response))

	var resultAppFund AppFundResponse
	err = json.Unmarshal(response, &resultAppFund)
	if err != nil {
		logger.Errorf("GetFund failed: %v", err)
		return
	}

	if resultAppFund.Status != "OK" {
		logger.Errorf("GetFund failed: %v", resultAppFund.Status)
		return
	}

	// Get My fund
	urlstr = getHTTPURL("user/" + fundid + "/" + userId)
	response, err = performHTTPGet(urlstr)
	if err != nil {
		logger.Errorf("GetFund failed: %v", err)
		return
	}

	logger.Debugf("GetFund: url=%v response=%v", urlstr, string(response))

	var resultAppMyFund AppMyFundResponse
	err = json.Unmarshal(response, &resultAppMyFund)
	if err != nil {
		logger.Errorf("GetFund failed: %v", err)
	} else {
		fund = resultAppFund.Msg
	}

	return
}

func GetFundMarkets(fundid string) (err error, fundmarkets []FundMarket) {
	err = nil

	fundmarket := FundMarket{
		Index: 1,
		Size:  1023.0,
		Type:  "购买",
	}
	fundmarkets = append(fundmarkets, fundmarket)

	fundmarket = FundMarket{
		Index: 2,
		Size:  1024.0,
		Type:  "购买",
	}
	fundmarkets = append(fundmarkets, fundmarket)

	fundmarket = FundMarket{
		Index: 3,
		Size:  1025.0,
		Type:  "购买",
	}
	fundmarkets = append(fundmarkets, fundmarket)

	fundmarket = FundMarket{
		Index: 4,
		Size:  1026.0,
		Type:  "购买",
	}
	fundmarkets = append(fundmarkets, fundmarket)

	fundmarket = FundMarket{
		Index: 5,
		Size:  1023.0,
		Type:  "赎回",
	}
	fundmarkets = append(fundmarkets, fundmarket)

	return
}

func GetFundNotices(fundid string) (err error, fundnotices []FundNotice) {
	err = nil

	fundnotice := FundNotice{
		Title:       "科瑞基金关于旗下部分产品增加农商银行为代销机构的公告",
		PublishTime: "2016-08-12",
	}
	fundnotices = append(fundnotices, fundnotice)

	fundnotice = FundNotice{
		Title:       "科瑞基金关于旗下部分产品增加农商银行为代销机构的公告",
		PublishTime: "2016-08-12",
	}
	fundnotices = append(fundnotices, fundnotice)

	fundnotice = FundNotice{
		Title:       "科瑞基金关于旗下部分产品增加农商银行为代销机构的公告",
		PublishTime: "2016-08-12",
	}
	fundnotices = append(fundnotices, fundnotice)

	fundnotice = FundNotice{
		Title:       "科瑞基金关于旗下部分产品增加农商银行为代销机构的公告",
		PublishTime: "2016-08-12",
	}
	fundnotices = append(fundnotices, fundnotice)

	return
}

func BuyFund(userId string, fundid string, amount float64) error {
	// Buy fund
	urlstr := getHTTPURL("transfer")
	request := AppTransfterFundRequest{
		EnrollID: userId,
		Name:     fundid,
		Funds:    int(amount),
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	response, err := performHTTPPost(urlstr, reqBody)
	if err != nil {
		logger.Errorf("BuyFund failed: %v", err)
		return err
	}

	logger.Debugf("BuyFund: url=%v request=%v response=%v", urlstr, request, string(response))

	var result AppTransfterFundResponse
	err = json.Unmarshal(response, &result)
	if err != nil {
		logger.Errorf("BuyFund failed: %v", err)
		return err
	}

	if result.Status != "OK" {
		logger.Errorf("BuyFund failed: %v", result.Status)
		return fmt.Errorf(result.Msg)
	}

	return nil
}

func RedeemFund(userId string, fundid string, quotas float64) error {
	// Redeem fund
	urlstr := getHTTPURL("transfer")
	request := AppTransfterFundRequest{
		EnrollID: userId,
		Name:     fundid,
		Funds:    int(quotas) * -1,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	response, err := performHTTPPost(urlstr, reqBody)
	if err != nil {
		logger.Errorf("RedeemFund failed: %v", err)
		return err
	}

	logger.Debugf("RedeemFund: url=%v request=%v response=%v", urlstr, request, string(response))

	var result AppTransfterFundResponse
	err = json.Unmarshal(response, &result)
	if err != nil {
		logger.Errorf("RedeemFund failed: %v", err)
		return err
	}

	if result.Status != "OK" {
		logger.Errorf("RedeemFund failed: %v", result.Status)
		return fmt.Errorf(result.Msg)
	}

	return nil
}

// CreateNewFund CreateNewFund
func CreateNewFund(
	userId string,
	fundid string,
	quotas float64,
	balance float64,
	tbalance float64,
	ttime int,
	tcount float64,
	tbuyper float64,
	tbuyall float64,
	netvalue float64) error {
	// Create New Fund
	urlstr := getHTTPURL("create")
	request := AppCreateFundRequest{
		Name:          fundid,
		Funds:         int(quotas),
		Assets:        int(balance),
		PartnerAssets: int(tbalance),
		PartnerTime:   ttime,
		BuyStart:      int(tcount),
		BuyPer:        int(tbuyper),
		BuyAll:        int(tbuyall),
		Netvalue:      int(netvalue),
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	response, err := performHTTPPost(urlstr, reqBody)
	if err != nil {
		logger.Errorf("CreateNewFund failed: %v", err)
		return err
	}

	logger.Debugf("CreateNewFund: url=%v request=%v response=%v", urlstr, request, string(response))

	var result AppCreateFundResponse
	err = json.Unmarshal(response, &result)
	if err != nil {
		logger.Errorf("CreateNewFund failed: %v", err)
		return err
	}

	if result.Status != "OK" {
		logger.Errorf("CreateNewFund failed: %v", result.Status)
		return fmt.Errorf(result.Msg)
	}

	return nil
}

// SetFundNetvalue SetFundNetvalue
func SetFundNetvalue(
	userId string,
	fundid string,
	netvalue float64) error {
	// Set Fund netvalue
	urlstr := getHTTPURL("setnet")
	request := AppSetFundNetvalueRequest{
		Name:     fundid,
		Netvalue: int(netvalue),
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	response, err := performHTTPPost(urlstr, reqBody)
	if err != nil {
		logger.Errorf("SetFundNetvalue failed: %v", err)
		return err
	}

	logger.Debugf("SetFundNetvalue: url=%v request=%v response=%v", urlstr, request, string(response))

	var result AppSetFundNetvalueResponse
	err = json.Unmarshal(response, &result)
	if err != nil {
		logger.Errorf("SetFundNetvalue failed: %v", err)
		return err
	}

	if result.Status != "OK" {
		logger.Errorf("SetFundNetvalue failed: %v", result.Status)
		return fmt.Errorf(result.Msg)
	}

	return nil
}

// SetFundThreshhold SetFundThreshhold
func SetFundThreshhold(
	userId string,
	fundid string,
	tbalance float64,
	ttime int,
	tcount float64,
	tbuyper float64,
	tbuyall float64) error {
	// Create New Fund
	urlstr := getHTTPURL("setlimit")
	request := AppSetFundThreshholdRequest{
		Name:          fundid,
		PartnerAssets: int(tbalance),
		PartnerTime:   ttime,
		BuyStart:      int(tcount),
		BuyPer:        int(tbuyper),
		BuyAll:        int(tbuyall),
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	response, err := performHTTPPost(urlstr, reqBody)
	if err != nil {
		logger.Errorf("SetFundThreshhold failed: %v", err)
		return err
	}

	logger.Debugf("SetFundThreshhold: url=%v request=%v response=%v", urlstr, request, string(response))

	var result AppSetFundThreshholdResponse
	err = json.Unmarshal(response, &result)
	if err != nil {
		logger.Errorf("SetFundThreshhold failed: %v", err)
		return err
	}

	if result.Status != "OK" {
		logger.Errorf("SetFundThreshhold failed: %v", result.Status)
		return fmt.Errorf(result.Msg)
	}

	return nil
}
