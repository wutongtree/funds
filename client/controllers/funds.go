package controllers

import (
	"fmt"
	"strconv"

	"github.com/wutongtree/funds/client/models"
)

type FundsController struct {
	BaseController
}

func (c *FundsController) ListMyFunds() {
	userid := c.UserUserId
	_, funds, _ := models.ListMyFunds(userid, 1, 100)
	c.Data["funds"] = funds
	c.Data["countFunds"] = len(funds)

	c.TplName = "funds/myfunds.tpl"
}

func (c *FundsController) GetFund() {
	fundid := c.GetString(":id")
	c.Data["fundid"] = fundid
	fmt.Printf("fundid: %v\n", fundid)

	// 余额
	userid := c.UserUserId

	// 我的基金
	myfund, _ := models.GetMyFund(userid, fundid)
	c.Data["myfund"] = myfund
	fmt.Printf("GetFund: %v\n", myfund)

	// 余额
	// _, c.Data["myaccount"] = models.GetMyAccount(userid)
	c.Data["myaccount"] = myfund.MyBalance

	// 净值走势

	// 市场动态
	_, markets := models.GetFundMarkets(fundid)
	c.Data["markets"] = markets

	// 基金公告
	_, notices := models.GetFundNotices(fundid)
	c.Data["notices"] = notices

	c.TplName = "funds/showfund.tpl"
}

func (c *FundsController) GetMyFund() {
	fundid := c.GetString("fundid")
	fmt.Printf("GetMyFund fundid: %v\n", fundid)

	userid := c.UserUserId

	// 获取净值
	netvalue, err := c.GetFloat("netvalue")
	if err != nil {
		logger.Errorf("netvalue GetFloat error: %v", err)

		netvalue = 1
	}

	// 我的基金
	myfund, _ := models.GetMyFund(userid, fundid)

	c.Data["json"] = map[string]interface{}{"code": 0, "myaccount": myfund.MyBalance, "myquotas": myfund.MyQuotas, "mymarketvalue": myfund.MyQuotas * netvalue}
	c.ServeJSON()
}

func (c *FundsController) BuyFund() {
	userid := c.UserUserId
	fundid := c.GetString("fundid")

	// 获取购买金额
	buycount, err := c.GetFloat("buycount")
	fmt.Printf("%v buycount: %v-%v\n", userid, fundid, buycount)
	if err != nil {
		logger.Errorf("buycount GetFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 1, "message": "购买失败：" + err.Error()}
		c.ServeJSON()

		return
	}

	// 获取净值
	netvalue, err := c.GetFloat("netvalue")
	if err != nil {
		logger.Errorf("netvalue GetFloat error: %v", err)

		netvalue = 1
	}

	// 购买基金
	err = models.BuyFund(userid, fundid, buycount/netvalue)
	if err != nil {
		logger.Errorf("BuyFund error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 1, "message": "购买失败：" + err.Error()}
		c.ServeJSON()

		return
	}

	// 获取持有份额
	myquotas, err := c.GetFloat("myquotas")
	if err != nil {
		logger.Errorf("myquotas ParseFloat error: %v", err)

		myquotas = 0
	}

	// 获取参考市值
	mymarketvalue, err := c.GetFloat("mymarketvalue")
	if err != nil {
		logger.Errorf("mymarketvalue ParseFloat error: %v", err)

		myquotas = 0
	}

	// 获取可用余额
	myaccount, err := c.GetFloat("myaccount")
	if err != nil {
		logger.Errorf("myaccount ParseFloat error: %v", err)

		myquotas = 0
	}

	// 获取可用份额

	logger.Infof("buycount=%v netvalue=%v myquotas=%v mymarketvalue=%v myaccount=%v", buycount, netvalue, myquotas, mymarketvalue, myaccount)

	// 更新数据
	c.Data["myaccount"] = myaccount - buycount

	c.Data["json"] = map[string]interface{}{"code": 0, "message": "购买成功：" + fmt.Sprintf("%v", buycount)}
	c.ServeJSON()
}

func (c *FundsController) RedeemFund() {
	userid := c.UserUserId
	fundid := c.GetString("fundid")

	redeemcount := c.GetString("redeemcount")
	fmt.Printf("%v redeemcount: %v-%v\n", userid, fundid, redeemcount)

	quotas, err := strconv.ParseFloat(redeemcount, 64)
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 1, "message": "赎回失败：" + err.Error()}
		c.ServeJSON()

		return
	}

	// 赎回基金
	err = models.RedeemFund(userid, fundid, quotas)
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 1, "message": "赎回失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	c.Data["json"] = map[string]interface{}{"code": 0, "message": "赎回成功：" + redeemcount}
	c.ServeJSON()
}

func (c *FundsController) GetNewFund() {

	c.TplName = "funds/createfund.tpl"
}

func (c *FundsController) CreateNewFund() {
	userid := c.UserUserId

	fundname := c.GetString("fundname")
	quotas, err := c.GetFloat("quotas")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	balance, err := c.GetFloat("balance")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	tbalance, err := c.GetFloat("tbalance")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	ttime, err := c.GetInt("ttime")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	tcount, err := c.GetFloat("tcount")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	tbuyper, err := c.GetFloat("tbuyper")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	tbuyall, err := c.GetFloat("tbuyall")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	netvalue, err := c.GetFloat("netvalue")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}

	// 新建基金
	err = models.CreateNewFund(userid, fundname, quotas, balance, tbalance, ttime, tcount, tbuyper, tbuyall, netvalue)
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "新建基金失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	c.Data["json"] = map[string]interface{}{"code": 1, "message": "新建基金成功：" + fundname}
	c.ServeJSON()
}

func (c *FundsController) ManageFund() {
	userid := c.UserUserId
	_, funds, _ := models.ListMyFunds(userid, 1, 100)
	c.Data["funds"] = funds
	c.Data["countFunds"] = len(funds)

	c.TplName = "funds/managefund.tpl"
}

func (c *FundsController) FundNetvalue() {
	userid := c.UserUserId

	fundname := c.GetString(":id")
	fund, _ := models.GetFund(userid, fundname)
	c.Data["fund"] = fund

	c.TplName = "funds/setfundnetvalue.tpl"
}

func (c *FundsController) FundThreshhold() {
	userid := c.UserUserId

	fundname := c.GetString(":id")
	fund, _ := models.GetFund(userid, fundname)
	c.Data["fund"] = fund

	c.TplName = "funds/setfundthreshhold.tpl"
}

func (c *FundsController) SetFundNetvalue() {
	userid := c.UserUserId

	fundname := c.GetString("fundname")
	netvalue, err := c.GetFloat("netvalue")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金净值失败：" + err.Error()}
		c.ServeJSON()

		return
	}

	// 设置基金净值
	err = models.SetFundNetvalue(userid, fundname, netvalue)
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金净值失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	c.Data["json"] = map[string]interface{}{"code": 1, "message": "设置基金净值成功：" + fundname}
	c.ServeJSON()
}

func (c *FundsController) SetFundThreshhold() {
	userid := c.UserUserId

	fundname := c.GetString("fundname")
	tbalance, err := c.GetFloat("tbalance")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金限制失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	ttime, err := c.GetInt("ttime")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金限制失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	tcount, err := c.GetFloat("tcount")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金限制失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	tbuyper, err := c.GetFloat("tbuyper")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金限制失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	tbuyall, err := c.GetFloat("tbuyall")
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金限制失败：" + err.Error()}
		c.ServeJSON()

		return
	}

	// 设置基金限制
	err = models.SetFundThreshhold(userid, fundname, tbalance, ttime, tcount, tbuyper, tbuyall)
	if err != nil {
		logger.Errorf("ParseFloat error: %v", err)

		c.Data["json"] = map[string]interface{}{"code": 0, "message": "设置基金限制失败：" + err.Error()}
		c.ServeJSON()

		return
	}
	c.Data["json"] = map[string]interface{}{"code": 1, "message": "设置基金限制成功：" + fundname}
	c.ServeJSON()
}
