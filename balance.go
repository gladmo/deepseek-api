package deepseek_api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (th DeepSeekError) Error() string {
	return fmt.Sprintf("DeepSeekApiError, code: %d, message: %s, cause: %s, how to fix: %s",
		th.Code, th.Message, th.Cause, th.Fix,
	)
}

func (th *Client) UserBalance() (reply UserBalanceReply, err error) {
	req, err := th.NewRequest(http.MethodGet, BalanceRoute, nil)
	if err != nil {
		return
	}

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = findReplyError(res.StatusCode)
		return
	}

	err = json.NewDecoder(res.Body).Decode(&reply)
	return
}

type UserBalanceReply struct {
	IsAvailable  bool          `json:"is_available"`
	BalanceInfos []BalanceInfo `json:"balance_infos"`
}

type BalanceInfo struct {
	Currency        string `json:"currency"`
	TotalBalance    string `json:"total_balance"`
	GrantedBalance  string `json:"granted_balance"`
	ToppedUpBalance string `json:"topped_up_balance"`
}
