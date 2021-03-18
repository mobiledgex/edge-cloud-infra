package ormclient

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/cloudcommon/node"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/jaeger/plugin/storage/es/spanstore/dbmodel"
)

type Client struct {
	SkipVerify bool
	Debug      bool
}

func (s *Client) DoLogin(uri, user, pass, otp, apikeyid, apikey string) (string, bool, error) {
	login := ormapi.UserLogin{
		Username: user,
		Password: pass,
		TOTP:     otp,
		ApiKeyId: apikeyid,
		ApiKey:   apikey,
	}
	result := make(map[string]interface{})
	status, err := s.PostJson(uri+"/login", "", &login, &result)
	if err != nil {
		return "", false, fmt.Errorf("login error, %s", err.Error())
	}
	if status != http.StatusOK {
		return "", false, fmt.Errorf("login status %d instead of OK(200)", status)
	}
	tokenI, ok := result["token"]
	if !ok {
		return "", false, fmt.Errorf("login token not found in response")
	}
	token, ok := tokenI.(string)
	if !ok {
		return "", false, fmt.Errorf("login token not string")
	}
	admin := false
	if adminI, ok := result["admin"]; ok {
		if adminB, ok := adminI.(bool); ok {
			admin = adminB
		}
	}
	return token, admin, nil
}

func (s *Client) CreateUser(uri string, user *ormapi.User) (*ormapi.UserResponse, int, error) {
	resp := ormapi.UserResponse{}
	st, err := s.PostJson(uri+"/usercreate", "", user, &resp)
	return &resp, st, err
}

func (s *Client) DeleteUser(uri, token string, user *ormapi.User) (int, error) {
	return s.PostJson(uri+"/auth/user/delete", token, user, nil)
}

func (s *Client) UpdateUser(uri, token string, createUserJSON string) (*ormapi.UserResponse, int, error) {
	resp := ormapi.UserResponse{}
	st, err := s.PostJson(uri+"/auth/user/update", token, createUserJSON, &resp)
	return &resp, st, err
}

func (s *Client) ShowUser(uri, token string, org *ormapi.ShowUser) ([]ormapi.User, int, error) {
	users := []ormapi.User{}
	status, err := s.PostJson(uri+"/auth/user/show", token, org, &users)
	return users, status, err
}

func (s *Client) NewPassword(uri, token, password string) (int, error) {
	newpw := ormapi.NewPassword{
		Password: password,
	}
	return s.PostJson(uri+"/auth/user/newpass", token, newpw, nil)
}

func (s *Client) CreateController(uri, token string, ctrl *ormapi.Controller) (int, error) {
	return s.PostJson(uri+"/auth/controller/create", token, ctrl, nil)
}

func (s *Client) DeleteController(uri, token string, ctrl *ormapi.Controller) (int, error) {
	return s.PostJson(uri+"/auth/controller/delete", token, ctrl, nil)
}

func (s *Client) ShowController(uri, token string) ([]ormapi.Controller, int, error) {
	ctrls := []ormapi.Controller{}
	status, err := s.PostJson(uri+"/auth/controller/show", token, nil, &ctrls)
	return ctrls, status, err
}

func (s *Client) CreateOrg(uri, token string, org *ormapi.Organization) (int, error) {
	return s.PostJson(uri+"/auth/org/create", token, org, nil)
}

func (s *Client) DeleteOrg(uri, token string, org *ormapi.Organization) (int, error) {
	return s.PostJson(uri+"/auth/org/delete", token, org, nil)
}

func (s *Client) UpdateOrg(uri, token string, jsonData string) (int, error) {
	return s.PostJson(uri+"/auth/org/update", token, jsonData, nil)
}

func (s *Client) ShowOrg(uri, token string) ([]ormapi.Organization, int, error) {
	orgs := []ormapi.Organization{}
	status, err := s.PostJson(uri+"/auth/org/show", token, nil, &orgs)
	return orgs, status, err
}

func (s *Client) CreateBillingOrg(uri, token string, bOrg *ormapi.BillingOrganization) (int, error) {
	return s.PostJson(uri+"/auth/billingorg/create", token, bOrg, nil)
}

func (s *Client) DeleteBillingOrg(uri, token string, bOrg *ormapi.BillingOrganization) (int, error) {
	return s.PostJson(uri+"/auth/billingorg/delete", token, bOrg, nil)
}

func (s *Client) UpdateBillingOrg(uri, token string, jsonData string) (int, error) {
	return s.PostJson(uri+"/auth/billingorg/update", token, jsonData, nil)
}

func (s *Client) ShowBillingOrg(uri, token string) ([]ormapi.BillingOrganization, int, error) {
	bOrgs := []ormapi.BillingOrganization{}
	status, err := s.PostJson(uri+"/auth/billingorg/show", token, nil, &bOrgs)
	return bOrgs, status, err
}

func (s *Client) AddChildOrg(uri, token string, bOrg *ormapi.BillingOrganization) (int, error) {
	return s.PostJson(uri+"/auth/billingorg/addchild", token, bOrg, nil)
}

func (s *Client) RemoveChildOrg(uri, token string, bOrg *ormapi.BillingOrganization) (int, error) {
	return s.PostJson(uri+"/auth/billingorg/removechild", token, bOrg, nil)
}

func (s *Client) CreateCloudletPoolAccessInvitation(uri, token string, op *ormapi.OrgCloudletPool) (int, error) {
	return s.PostJson(uri+"/auth/cloudletpoolaccessinvitation/create", token, op, nil)
}

func (s *Client) DeleteCloudletPoolAccessInvitation(uri, token string, op *ormapi.OrgCloudletPool) (int, error) {
	return s.PostJson(uri+"/auth/cloudletpoolaccessinvitation/delete", token, op, nil)
}

func (s *Client) ShowCloudletPoolAccessInvitation(uri, token string, filter *ormapi.OrgCloudletPool) ([]ormapi.OrgCloudletPool, int, error) {
	ops := []ormapi.OrgCloudletPool{}
	status, err := s.PostJson(uri+"/auth/cloudletpoolaccessinvitation/show", token, filter, &ops)
	return ops, status, err
}

func (s *Client) CreateCloudletPoolAccessConfirmation(uri, token string, op *ormapi.OrgCloudletPool) (int, error) {
	return s.PostJson(uri+"/auth/cloudletpoolaccessconfirmation/create", token, op, nil)
}

func (s *Client) DeleteCloudletPoolAccessConfirmation(uri, token string, op *ormapi.OrgCloudletPool) (int, error) {
	return s.PostJson(uri+"/auth/cloudletpoolaccessconfirmation/delete", token, op, nil)
}

func (s *Client) ShowCloudletPoolAccessConfirmation(uri, token string, filter *ormapi.OrgCloudletPool) ([]ormapi.OrgCloudletPool, int, error) {
	ops := []ormapi.OrgCloudletPool{}
	status, err := s.PostJson(uri+"/auth/cloudletpoolaccessconfirmation/show", token, filter, &ops)
	return ops, status, err
}

func (s *Client) ShowCloudletPoolAccessGranted(uri, token string, filter *ormapi.OrgCloudletPool) ([]ormapi.OrgCloudletPool, int, error) {
	ops := []ormapi.OrgCloudletPool{}
	status, err := s.PostJson(uri+"/auth/cloudletpoolaccessgranted/show", token, filter, &ops)
	return ops, status, err
}

func (s *Client) ShowOrgCloudlet(uri, token string, in *ormapi.OrgCloudlet) ([]edgeproto.Cloudlet, int, error) {
	out := []edgeproto.Cloudlet{}
	status, err := s.PostJson(uri+"/auth/orgcloudlet/show", token, in, &out)
	return out, status, err
}

func (s *Client) ShowOrgCloudletInfo(uri, token string, in *ormapi.OrgCloudlet) ([]edgeproto.CloudletInfo, int, error) {
	out := []edgeproto.CloudletInfo{}
	status, err := s.PostJson(uri+"/auth/orgcloudletinfo/show", token, in, &out)
	return out, status, err
}

func (s *Client) AddUserRole(uri, token string, role *ormapi.Role) (int, error) {
	return s.PostJson(uri+"/auth/role/adduser", token, role, nil)
}

func (s *Client) RemoveUserRole(uri, token string, role *ormapi.Role) (int, error) {
	return s.PostJson(uri+"/auth/role/removeuser", token, role, nil)
}

func (s *Client) ShowUserRole(uri, token string) ([]ormapi.Role, int, error) {
	roles := []ormapi.Role{}
	status, err := s.PostJson(uri+"/auth/role/showuser", token, nil, &roles)
	return roles, status, err
}

func (s *Client) ShowRoleAssignment(uri, token string) ([]ormapi.Role, int, error) {
	roles := []ormapi.Role{}
	status, err := s.PostJson(uri+"/auth/role/assignment/show", token, nil, &roles)
	return roles, status, err
}

func (s *Client) CreateData(uri, token string, data *ormapi.AllData, cb func(res *ormapi.Result)) (int, error) {
	res := ormapi.Result{}
	var reserr error
	var resstatus int
	status, err := s.PostJsonStreamOut(uri+"/auth/data/create", token, data, &res, func() {
		if res.Code != 0 {
			reserr = fmt.Errorf(res.Message)
			resstatus = res.Code
		}
		cb(&res)
	})
	if reserr != nil {
		return resstatus, reserr
	}
	return status, err
}

func (s *Client) DeleteData(uri, token string, data *ormapi.AllData, cb func(res *ormapi.Result)) (int, error) {
	res := ormapi.Result{}
	var reserr error
	var resstatus int
	status, err := s.PostJsonStreamOut(uri+"/auth/data/delete", token, data, &res, func() {
		if res.Code != 0 {
			reserr = fmt.Errorf(res.Message)
			resstatus = res.Code
		}
		cb(&res)
	})
	if reserr != nil {
		return resstatus, reserr
	}
	return status, err
}

func (s *Client) ShowData(uri, token string) (*ormapi.AllData, int, error) {
	data := ormapi.AllData{}
	status, err := s.PostJson(uri+"/auth/data/show", token, nil, &data)
	return &data, status, err
}

func (s *Client) UpdateConfig(uri, token string, config map[string]interface{}) (int, error) {
	return s.PostJson(uri+"/auth/config/update", token, config, nil)
}

func (s *Client) ResetConfig(uri, token string) (int, error) {
	return s.PostJson(uri+"/auth/config/reset", token, nil, nil)
}

func (s *Client) ShowConfig(uri, token string) (*ormapi.Config, int, error) {
	config := ormapi.Config{}
	status, err := s.PostJson(uri+"/auth/config/show", token, nil, &config)
	return &config, status, err
}

func (s *Client) PublicConfig(uri string) (*ormapi.Config, int, error) {
	config := ormapi.Config{}
	status, err := s.PostJson(uri+"/publicconfig", "", nil, &config)
	return &config, status, err
}

func (s *Client) RestrictedUserUpdate(uri, token string, user map[string]interface{}) (int, error) {
	return s.PostJson(uri+"/auth/restricted/user/update", token, user, nil)
}

func (s *Client) ShowAuditSelf(uri, token string, query *ormapi.AuditQuery) ([]ormapi.AuditResponse, int, error) {
	resp := []ormapi.AuditResponse{}
	status, err := s.PostJson(uri+"/auth/audit/showself", token, query, &resp)
	return resp, status, err
}

func (s *Client) ShowAuditOrg(uri, token string, query *ormapi.AuditQuery) ([]ormapi.AuditResponse, int, error) {
	resp := []ormapi.AuditResponse{}
	status, err := s.PostJson(uri+"/auth/audit/showorg", token, query, &resp)
	return resp, status, err
}

func (s *Client) ShowEvents(uri, token string, query *node.EventSearch) ([]node.EventData, int, error) {
	resp := []node.EventData{}
	status, err := s.PostJson(uri+"/auth/events/show", token, query, &resp)
	return resp, status, err
}

func (s *Client) FindEvents(uri, token string, query *node.EventSearch) ([]node.EventData, int, error) {
	resp := []node.EventData{}
	status, err := s.PostJson(uri+"/auth/events/find", token, query, &resp)
	return resp, status, err
}

func (s *Client) EventTerms(uri, token string, query *node.EventSearch) (*node.EventTerms, int, error) {
	resp := node.EventTerms{}
	status, err := s.PostJson(uri+"/auth/events/terms", token, query, &resp)
	return &resp, status, err
}

func (s *Client) ShowSpans(uri, token string, query *node.SpanSearch) ([]node.SpanOutCondensed, int, error) {
	resp := []node.SpanOutCondensed{}
	status, err := s.PostJson(uri+"/auth/spans/show", token, query, &resp)
	return resp, status, err
}

func (s *Client) ShowSpansVerbose(uri, token string, query *node.SpanSearch) ([]dbmodel.Span, int, error) {
	resp := []dbmodel.Span{}
	status, err := s.PostJson(uri+"/auth/spans/showverbose", token, query, &resp)
	return resp, status, err
}

func (s *Client) SpanTerms(uri, token string, query *node.SpanSearch) (*node.SpanTerms, int, error) {
	resp := node.SpanTerms{}
	status, err := s.PostJson(uri+"/auth/spans/terms", token, query, &resp)
	return &resp, status, err
}

func (s *Client) ShowAppMetrics(uri, token string, query *ormapi.RegionAppInstMetrics) (*ormapi.AllMetrics, int, error) {
	metrics := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"/auth/metrics/app", token, query, &metrics)
	return &metrics, status, err
}

func (s *Client) ShowClusterMetrics(uri, token string, query *ormapi.RegionClusterInstMetrics) (*ormapi.AllMetrics, int, error) {
	metrics := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"/auth/metrics/cluster", token, query, &metrics)
	return &metrics, status, err
}

func (s *Client) ShowCloudletMetrics(uri, token string, query *ormapi.RegionCloudletMetrics) (*ormapi.AllMetrics, int, error) {
	metrics := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"/auth/metrics/cloutlet", token, query, &metrics)
	return &metrics, status, err
}

func (s *Client) ShowClientMetrics(uri, token string, query *ormapi.RegionClientMetrics) (*ormapi.AllMetrics, int, error) {
	metrics := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"/auth/metrics/client", token, query, &metrics)
	return &metrics, status, err
}

func (s *Client) ShowAppEvents(uri, token string, query *ormapi.RegionAppInstEvents) (*ormapi.AllMetrics, int, error) {
	metrics := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"/auth/events/app", token, query, &metrics)
	return &metrics, status, err
}
func (s *Client) ShowClusterEvents(uri, token string, query *ormapi.RegionClusterInstEvents) (*ormapi.AllMetrics, int, error) {
	metrics := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"/auth/events/cluster", token, query, &metrics)
	return &metrics, status, err
}

func (s *Client) ShowCloudletEvents(uri, token string, query *ormapi.RegionCloudletEvents) (*ormapi.AllMetrics, int, error) {
	metrics := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"/auth/events/cloudlet", token, query, &metrics)
	return &metrics, status, err
}

func (s *Client) ShowAppUsage(uri, token string, query *ormapi.RegionAppInstUsage) (*ormapi.AllMetrics, int, error) {
	usage := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"auth/usage/app", token, query, &usage)
	return &usage, status, err
}

func (s *Client) ShowClusterUsage(uri, token string, query *ormapi.RegionClusterInstUsage) (*ormapi.AllMetrics, int, error) {
	usage := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"auth/usage/cluster", token, query, &usage)
	return &usage, status, err
}

func (s *Client) ShowCloudletPoolUsage(uri, token string, query *ormapi.RegionCloudletPoolUsage) (*ormapi.AllMetrics, int, error) {
	usage := ormapi.AllMetrics{}
	status, err := s.PostJson(uri+"auth/usage/cloudletpool", token, query, &usage)
	return &usage, status, err
}

func (s *Client) PostJsonSend(uri, token string, reqData interface{}) (*http.Response, error) {
	var body io.Reader
	var datastr string
	if reqData != nil {
		str, ok := reqData.(string)
		if ok {
			// assume string is json data
			body = bytes.NewBuffer([]byte(str))
			datastr = str
		} else {
			out, err := json.Marshal(reqData)
			if err != nil {
				return nil, fmt.Errorf("post %s marshal req failed, %s", uri, err.Error())
			}
			body = bytes.NewBuffer(out)
			datastr = string(out)
		}
	} else {
		body = nil
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, fmt.Errorf("post %s http req failed, %s", uri, err.Error())
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	tlsConfig := &tls.Config{}
	if s.SkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}
	if s.Debug {
		curlcmd := fmt.Sprintf(`curl -X POST "%s" -H "Content-Type: application/json"`, uri)
		if token != "" {
			curlcmd += ` -H "Authorization: Bearer ${TOKEN}"`
		}
		if s.SkipVerify {
			curlcmd += " -k"
		}
		if datastr != "" {
			curlcmd += ` --data-raw '` + datastr + `'`
		}
		fmt.Printf("%s\n", curlcmd)
	}

	client := &http.Client{Transport: tr}
	return client.Do(req)
}

func (s *Client) PostJson(uri, token string, reqData interface{}, replyData interface{}) (int, error) {
	resp, err := s.PostJsonSend(uri, token, reqData)
	if err != nil {
		return 0, fmt.Errorf("post %s client do failed, %s", uri, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK && replyData != nil {
		err = json.NewDecoder(resp.Body).Decode(replyData)
		if err != nil && err != io.EOF {
			return resp.StatusCode, fmt.Errorf("post %s decode resp failed, %v", uri, err)
		}
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}
		res := ormapi.Result{}
		err = json.Unmarshal(body, &res)
		if err != nil {
			// string error
			return resp.StatusCode, fmt.Errorf("%s", body)
		}
		return resp.StatusCode, errors.New(res.Message)
	}
	return resp.StatusCode, nil
}

func (s *Client) PostJsonStreamOut(uri, token string, reqData, replyData interface{}, replyReady func()) (int, error) {
	if strings.Contains(uri, "ws/api/v1") {
		return s.HandleWebsocketStreamOut(uri, token, nil, reqData, replyData, replyReady)
	} else {
		return s.handleHttpStreamOut(uri, token, reqData, replyData, replyReady)
	}
}

func (s *Client) handleHttpStreamOut(uri, token string, reqData, replyData interface{}, replyReady func()) (int, error) {
	resp, err := s.PostJsonSend(uri, token, reqData)
	if err != nil {
		return 0, fmt.Errorf("post %s client do failed, %s", uri, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}
		res := ormapi.Result{}
		err = json.Unmarshal(body, &res)
		if err != nil {
			// string error
			return resp.StatusCode, fmt.Errorf("%s", body)
		}
		return resp.StatusCode, errors.New(res.Message)
	}
	payload := ormapi.StreamPayload{}
	if replyData != nil {
		payload.Data = replyData
	}

	dec := json.NewDecoder(resp.Body)
	for {
		if replyData != nil {
			// clear passed in buffer for next iteration.
			// replyData must be pointer to object.
			p := reflect.ValueOf(replyData).Elem()
			p.Set(reflect.Zero(p.Type()))
		}

		payload.Result = nil
		err := dec.Decode(&payload)
		if err != nil {
			if err == io.EOF {
				break
			}
			return resp.StatusCode, fmt.Errorf("post %s decode resp failed, %s", uri, err.Error())
		}
		if payload.Result != nil {
			return resp.StatusCode, errors.New(payload.Result.Message)
		}
		if replyReady != nil {
			replyReady()
		}
	}
	return resp.StatusCode, nil
}

func (s *Client) WebsocketConn(uri, token string, reqData interface{}) (*websocket.Conn, error) {
	var body []byte
	if reqData != nil {
		str, ok := reqData.(string)
		if ok {
			// assume string is json data
			body = []byte(str)
		} else {
			out, err := json.Marshal(reqData)
			if err != nil {
				return nil, fmt.Errorf("post %s marshal req failed, %s", uri, err.Error())
			}
			if s.Debug {
				fmt.Printf("posting %s\n", string(out))
			}
			body = out
		}
	} else {
		body = nil
	}

	var ws *websocket.Conn
	var err error
	if strings.HasPrefix(uri, "wss") {
		d := websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		}
		ws, _, err = d.Dial(uri, nil)
	} else {
		ws, _, err = websocket.DefaultDialer.Dial(uri, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("websocket connect to %s failed, %s", uri, err.Error())
	}

	// Authorize JWT with server
	authData := fmt.Sprintf(`{"token": "%s"}`, token)
	if err := ws.WriteMessage(websocket.TextMessage, []byte(authData)); err != nil {
		return nil, fmt.Errorf("websocket auth to %s failed with data %v, %s", uri, authData, err.Error())
	}

	// Send request data
	if err := ws.WriteMessage(websocket.TextMessage, []byte(body)); err != nil {
		return nil, fmt.Errorf("websocket send to %s failed, %s", uri, err.Error())
	}
	return ws, nil
}

func (s *Client) HandleWebsocketStreamOut(uri, token string, reader *bufio.Reader, reqData, replyData interface{}, replyReady func()) (int, error) {
	wsPayload, ok := replyData.(*ormapi.WSStreamPayload)
	if !ok {
		return 0, fmt.Errorf("response can only be of type WSStreamPayload")
	}
	ws, err := s.WebsocketConn(uri, token, reqData)
	if err != nil {
		return 0, fmt.Errorf("post %s client do failed, %s", uri, err.Error())
	}
	if reader != nil {
		go func() {
			for {
				text, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				if err := ws.WriteMessage(websocket.TextMessage, []byte(text)); err != nil {
					break
				}
			}
		}()
	}
	payload := wsPayload
	for {
		if payload != nil {
			// clear passed in buffer for next iteration.
			// payload must be pointer to object.
			p := reflect.ValueOf(payload).Elem()
			p.Set(reflect.Zero(p.Type()))
		}

		err := ws.ReadJSON(&payload)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				break
			}
			return http.StatusBadRequest, fmt.Errorf("post %s decode resp failed, %s", uri, err.Error())
		}
		if payload.Code != http.StatusOK {
			if payload.Data == nil {
				return payload.Code, nil
			}
			errRes := edgeproto.Result{}
			err = mapstructure.Decode(payload.Data, &errRes)
			if err == nil {
				return payload.Code, errors.New(errRes.Message)
			}
			return payload.Code, nil
		}
		if replyReady != nil {
			replyReady()
		}
	}
	return http.StatusOK, nil
}

func (s *Client) ArtifactoryResync(uri, token string) (int, error) {
	return s.PostJson(uri+"/auth/artifactory/resync", token, nil, nil)
}

func (s *Client) GitlabResync(uri, token string) (int, error) {
	return s.PostJson(uri+"/auth/gitlab/resync", token, nil, nil)
}

func (s *Client) CreateAlertReceiver(uri, token string, receiver *ormapi.AlertReceiver) (int, error) {
	return s.PostJson(uri+"/auth/alertreceiver/create", token, receiver, nil)
}

func (s *Client) DeleteAlertReceiver(uri, token string, receiver *ormapi.AlertReceiver) (int, error) {
	return s.PostJson(uri+"/auth/alertreceiver/delete", token, receiver, nil)
}

func (s *Client) ShowAlertReceiver(uri, token string, in *ormapi.AlertReceiver) ([]ormapi.AlertReceiver, int, error) {
	receivers := []ormapi.AlertReceiver{}
	status, err := s.PostJson(uri+"/auth/alertreceiver/show", token, in, &receivers)
	return receivers, status, err
}

func (s *Client) CreateUserApiKey(uri, token string, userApiKey *ormapi.CreateUserApiKey) (*ormapi.CreateUserApiKey, int, error) {
	resp := ormapi.CreateUserApiKey{}
	st, err := s.PostJson(uri+"/auth/user/create/apikey", token, userApiKey, &resp)
	return &resp, st, err
}

func (s *Client) DeleteUserApiKey(uri, token string, userApiKey *ormapi.CreateUserApiKey) (int, error) {
	return s.PostJson(uri+"/auth/user/delete/apikey", token, userApiKey, nil)
}

func (s *Client) ShowUserApiKey(uri, token string, userApiKey *ormapi.CreateUserApiKey) ([]ormapi.CreateUserApiKey, int, error) {
	userApiKeys := []ormapi.CreateUserApiKey{}
	status, err := s.PostJson(uri+"/auth/user/show/apikey", token, userApiKey, &userApiKeys)
	return userApiKeys, status, err
}
