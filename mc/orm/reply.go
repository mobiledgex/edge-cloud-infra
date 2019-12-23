package orm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

type M map[string]interface{}

func Msg(msg string) *ormapi.Result {
	return &ormapi.Result{Message: msg}
}

func MsgErr(err error) *ormapi.Result {
	return &ormapi.Result{Message: err.Error()}
}

func ctrlErr(c echo.Context, err error) error {
	msg := "controller connect error, " + err.Error()
	return c.JSON(http.StatusBadRequest, Msg(msg))
}

func dbErr(err error) error {
	return fmt.Errorf("database error, %s", err.Error())
}

func bindErr(c echo.Context, err error) error {
	msg := "Invalid POST data, " + err.Error()
	return c.JSON(http.StatusBadRequest, Msg(msg))
}

func setReply(c echo.Context, err error, data interface{}) error {
	code := http.StatusOK
	if err != nil {
		switch err {
		case echo.ErrForbidden:
			code = http.StatusForbidden
		case echo.ErrNotFound:
			code = http.StatusNotFound
		default:
			code = http.StatusBadRequest
		}
	}
	if ws := GetWs(c); ws != nil {
		res := ormapi.Result{}
		res.Code = code
		payload := ormapi.StreamPayload{Result: &res}
		if err != nil {
			res.Message = err.Error()
		}
		if data != nil {
			payload.Data = data
		}
		return ws.WriteJSON(payload)
	}
	if err != nil {
		return c.JSON(code, MsgErr(err))
	}
	return c.JSON(code, data)
}

// streamReply funcs used by alldata always send back just a status
// message, never an error - even if an error was generated. So they
// never use payload.Result, which is used to convey an error.
func streamReply(c echo.Context, desc string, err error, hadErr *bool) {
	res := "ok"
	if err == echo.ErrForbidden {
		res = "forbidden"
		*hadErr = true
	} else if err != nil {
		res = err.Error()
		*hadErr = true
	}
	streamReplyMsg(c, desc, res)
}

func streamReplyMsg(c echo.Context, desc, res string) {
	payload := ormapi.StreamPayload{
		Data: &ormapi.Result{
			Message: fmt.Sprintf("%s: %s", desc, res),
			Code:    0,
		},
	}
	json.NewEncoder(c.Response()).Encode(payload)
	c.Response().Flush()
}

func streamErr(c echo.Context, msg string) {
	payload := ormapi.StreamPayload{
		Result: &ormapi.Result{
			Message: msg,
		},
	}
	json.NewEncoder(c.Response()).Encode(payload)
	c.Response().Flush()
}
