package kotak_neo_api

import (
	"github.com/shikharvaish28/kotak-neo-api/api"
	"github.com/shikharvaish28/kotak-neo-api/websocket"
	"log"
	"context"
	"errors"
	"fmt"
)

type KotakClient struct {
	config          *api.Configuration
	websocket       *websocket.HSWrapper // TODO: think about abstracting this away from the wrapper and building a simpler class instead.
	loginAPI        *api.LoginAPI
	orderAPI        *api.OrderService
	orderReportAPI  *api.OrderReportAPI
	orderHistoryAPI *api.OrderHistoryService
	tradeReportAPI  *api.TradeReportAPI
	positionsAPI    *api.PositionsAPI
	holdingsAPI     *api.PortfolioHoldingsAPI
	modificationAPI *api.ModifyOrderAPI
	marginAPI       *api.MarginAPI
	limitsAPI       *api.LimitsAPI
	logoutAPI       *api.LogoutAPI
}

// note: a broker interface should give you a channel for consumption and a client for placing orders.
func NewKotakClient(configuration *api.Configuration) (*KotakClient, chan websocket.BrokerEvent) {
	ws, wsChannel := websocket.NewHSWrapper()
	apiClient := api.NewAPIClient(configuration)
	loginAPI := api.NewLoginAPI(apiClient)
	orderAPI := api.NewOrderService(apiClient)
	orderReportAPI := api.NewOrderReportAPI(apiClient)
	orderHistoryAPI := api.NewOrderHistoryService(apiClient)
	tradeReportAPI := api.NewTradeReportAPI(apiClient)
	positionsAPI := api.NewPositionsAPI(apiClient)
	holdingsAPI := api.NewPortfolioHoldingsAPI(apiClient)
	modificationAPI := api.NewModifyOrderAPI(apiClient)
	marginAPI := api.NewMarginAPI(apiClient)
	limitsAPI := api.NewLimitsAPI(apiClient)
	logoutAPI := api.NewLogoutAPI(apiClient)

	return &KotakClient{
		config:          configuration,
		websocket:       ws,
		loginAPI:        loginAPI,
		orderAPI:        orderAPI,
		orderReportAPI:  orderReportAPI,
		orderHistoryAPI: orderHistoryAPI,
		tradeReportAPI:  tradeReportAPI,
		positionsAPI:    positionsAPI,
		holdingsAPI:     holdingsAPI,
		modificationAPI: modificationAPI,
		marginAPI:       marginAPI,
		limitsAPI:       limitsAPI,
		logoutAPI:       logoutAPI,
	}, wsChannel
}

// Subscribe handles the subscription to live feeds
func (c *KotakClient) Subscribe(ctx context.Context, instrumentTokens []map[string]string, isIndex, isDepth bool) {
	if c.config.EditToken == "" || c.config.EditSid == "" {
		log.Println("Please complete the Login Flow to Subscribe the Scrips")
		return
	}

	err := c.websocket.GetLiveFeed(instrumentTokens, isIndex, isDepth)
	if err != nil {
		log.Println("Failed to subscribe to live feeds:", err)
	}
}

func (c *KotakClient) Login(password, mobileNumber, userid, pan, MPin string) (map[string]interface{}, error) {
	if mobileNumber == "" && userid == "" && pan == "" {
		errorResponse := map[string]interface{}{
			"error": []map[string]string{
				{"code": "10300", "message": "Validation Errors! Any of Mobile Number, User Id and Pan has to pass as part of login"},
			},
		}
		return errorResponse, errors.New("validation error")
	}

	viewToken, err := c.loginAPI.GenerateViewToken(password, mobileNumber, userid, pan, MPin)
	if err != nil {
		return nil, err
	}

	_, err = c.loginAPI.GenerateOTP()
	if err != nil {
		return map[string]interface{}{
			"error": []map[string]string{
				{"code": "10522", "message": "Issues while generating OTP! Try to login again."},
			},
		}, errors.New("issues while generating OTP")
	}

	return viewToken, nil
}

func (c *KotakClient) Session2FA(otp string) (map[string]interface{}, error) {
	editToken, err := c.loginAPI.Login2FA(otp)
	if err != nil {
		return nil, err
	}
	return editToken, nil
}

func (c *KotakClient) PlaceOrder(req api.OrderRequest) (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		err := api.PlaceOrderValidation(req.ExchangeSegment,
			req.Product,
			fmt.Sprintf("%f", req.Price),
			req.OrderType,
			fmt.Sprintf("%d", req.Quantity),
			req.Validity,
			req.TradingSymbol,
			req.TransactionType,
			fmt.Sprintf("%f", req.Amo),
			fmt.Sprintf("%d", req.DisclosedQuantity),
			req.MarketProtection,
			req.Pf,
			fmt.Sprintf("%f", req.TriggerPrice),
			req.Tag)
		if err != nil {
			return nil, err
		}
		return c.orderAPI.PlaceOrder(req)
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) CancelOrder(orderId, amo string) (map[string]interface{}, error) {
	if amo == "" {
		amo = "NO"
	}

	if c.config.EditToken != "" && c.config.EditSid != "" {
		err := api.CancelOrderValidation(orderId, amo)
		if err != nil {
			return nil, err
		}
		// TODO: recheck amount?
		return c.orderAPI.CancelOrder(orderId, false, 0.0)
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) OrderReport() (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		return c.orderReportAPI.GetOrderReport()
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) OrderHistory(orderId string) (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		err := api.OrderHistoryValidation(orderId)
		if err != nil {
			return nil, err
		}
		return c.orderHistoryAPI.FetchOrderHistory(orderId)
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) TradeReport(orderId string) (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		return c.tradeReportAPI.GetTradeReport(orderId)
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) Positions() (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		return c.positionsAPI.GetPositions()
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) Holdings() (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		return c.holdingsAPI.Fetch()
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) ModifyOrder(modificationRequest api.ModifyOrderRequest) (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		if modificationRequest.OrderID != "" &&
			modificationRequest.InstrumentToken != "" &&
			modificationRequest.ExchangeSegment != "" &&
			modificationRequest.Product != "" &&
			modificationRequest.TradingSymbol != "" {

			modificationRequest.ExchangeSegment = api.ExchangeSegment[modificationRequest.ExchangeSegment]
			modificationRequest.Product = api.Product[modificationRequest.Product]
			modificationRequest.OrderType = api.OrderType[modificationRequest.OrderType]

			return c.modificationAPI.ModifyQuick(modificationRequest)
		} else if modificationRequest.OrderID != "" &&
			modificationRequest.InstrumentToken == "" &&
			modificationRequest.ExchangeSegment == "" &&
			modificationRequest.TradingSymbol == "" {
			return c.modificationAPI.ModifyWithOrderID(modificationRequest)
		} else {
			return map[string]interface{}{
				"error": []map[string]string{
					{"code": "900", "message": "Order ID is Mandate if we need to proceed further!"},
				},
			}, errors.New("order ID is Mandate if we need to proceed further")
		}
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) MarginRequired(req api.MarginRequest) (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		err := api.MarginValidation(req.ExchangeSegment,
			fmt.Sprintf("%f", req.Price),
			req.OrderType,
			req.Product,
			fmt.Sprintf("%d", req.Quantity),
			req.InstrumentToken,
			req.TransactionType,
			fmt.Sprintf("%f", req.TriggerPrice))
		if err != nil {
			return nil, err
		}

		req.ExchangeSegment = api.ExchangeSegment[req.ExchangeSegment]
		req.Product = api.Product[req.Product]
		req.OrderType = api.OrderType[req.OrderType]

		return c.marginAPI.GetMargin(req)
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) Limits(segment, exchange, product string) (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		err := api.LimitsValidation(segment, exchange, product)
		if err != nil {
			return nil, err
		}
		return c.limitsAPI.GetLimits(segment, exchange, product)
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

func (c *KotakClient) Logout() (map[string]interface{}, error) {
	if c.config.EditToken != "" && c.config.EditSid != "" {
		_, err := c.logoutAPI.LogoutUser()
		if err != nil {
			return nil, err
		}
		c.config.BearerToken = ""
		c.config.EditToken = ""
		c.config.EditSid = ""
		return map[string]interface{}{
			"State": "OK", "message": "You have been successfully logged out",
		}, nil
	}
	return map[string]interface{}{
		"error": []map[string]string{
			{"code": "900", "message": "Complete the 2fa process before accessing this application"},
		},
	}, errors.New("please complete the 2FA process before placing orders")
}

// TODO: complete unsubscribe WS.