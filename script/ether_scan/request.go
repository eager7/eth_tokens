package ether_scan

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/eager7/eth_tokens/script/built"
	"github.com/parnurzeal/gorequest"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const urlErc20EtherScan = `https://etherscan.io/tokens?ps=100&p=`
const urlNftEtherScan = `https://etherscan.io/tokens-nft?ps=50&p=`
const urlTokenInfo = `https://etherscan.io/token/%s`
const pageMax = 10

func RequestTokenListByPage(url string) ([]built.TokenInfo, error) {
	res := gorequest.New().Proxy(UserProxyLists[rand.Intn(len(UserProxyLists))]).Set("user-agent", UserAgentLists[rand.Intn(len(UserAgentLists))])
	ret, body, errs := res.Timeout(time.Second*5).Retry(5, time.Second, http.StatusRequestTimeout, http.StatusBadRequest).Get(url).End()
	if errs != nil || ret.StatusCode != http.StatusOK {
		req, err := res.MakeRequest()
		if err == nil && req != nil && ret != nil {
			fmt.Printf("request status:%d, body:%+v\n", ret.StatusCode, req)
		}
		var errStr string
		for _, e := range errs {
			errStr += e.Error()
		}
		return nil, errors.New("request err:" + errStr)
	}
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, errors.New("NewDocumentFromReader err:" + err.Error())
	}

	var tokens []built.TokenInfo
	dom.Find("div.media").Each(func(i int, selection *goquery.Selection) {
		ret, err := selection.Html()
		if err != nil {
			fmt.Println("selection to html err:" + err.Error())
			return
		}
		info := selection.Find("a").Text()
		index := strings.Index(info, "(")
		var name, symbol string
		if index == -1 {
			name = info
		} else {
			symbol = strings.Replace(info[index+1:], ")", "", -1)
		}
		description := selection.Find("p").Text()

		reIcon := regexp.MustCompile(`.*?src="(?P<icon>[^"]*)(?s:".*)`)
		icon := reIcon.ReplaceAllString(ret, "$icon")

		reContract := regexp.MustCompile(`.*?href="/token/(?P<contract>[^"]*)(?s:".*)`)
		contract := reContract.ReplaceAllString(ret, "$contract")

		if !IsValidAddress(contract) {
			fmt.Println("IsValidAddress false:", contract, ret)
		}
		if strings.Contains(name, contract) { //有的代币没有名称，ether scan会返回代币地址，我们过滤掉这种代币
			name = ""
		}
		var token = built.TokenInfo{
			Symbol:     symbol,
			Name:       name,
			Type:       "",
			Address:    contract,
			EnsAddress: "",
			Decimals:   0,
			Website:    "",
			Logo: built.Logo{
				Src: "https://etherscan.io" + icon,
			},
			Support: built.Support{
				Email: "",
				Url:   description,
			},
			Social: built.Social{},
		}
		tokens = append(tokens, token)
	})

	return tokens, nil
}

func IsValidAddress(address string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(address)
}

func RequestTokenLogo(token *built.TokenInfo) error {
	u := fmt.Sprintf(urlTokenInfo, token.Address)
	log.Debug("request token logo:", u)
	res := gorequest.New().Proxy(UserProxyLists[rand.Intn(len(UserProxyLists))]).Set("user-agent", UserAgentLists[rand.Intn(len(UserAgentLists))])
	ret, body, errs := res.Timeout(time.Second*5).Retry(5, time.Second, http.StatusRequestTimeout, http.StatusBadRequest).Get(u).End()
	if errs != nil || ret.StatusCode != http.StatusOK {
		req, err := res.MakeRequest()
		if err == nil && req != nil && ret != nil {
			fmt.Printf("request status:%d, body:%+v\n", ret.StatusCode, req)
		}
		var errStr string
		for _, e := range errs {
			errStr += e.Error()
		}
		return errors.New("request err:" + errStr)
	}
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return errors.New("NewDocumentFromReader err:" + err.Error())
	}
	dom.Find("h1").Each(func(i int, selection *goquery.Selection) {
		ret, err := selection.Html()
		if err != nil {
			fmt.Println("selection to html err:" + err.Error())
			return
		}
		reIcon := regexp.MustCompile(`.*?src="(?P<icon>[^"]*)(?s:".*)`)
		icon := reIcon.ReplaceAllString(ret, "$icon")
		token.Logo.Src = "https://etherscan.io" + icon
		token.Logo.Src = strings.Replace(token.Logo.Src, "\n", "", -1)
	})
	return nil
}
