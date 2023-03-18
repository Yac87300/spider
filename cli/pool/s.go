package pool



type Iplist struct {
	Msg string `json:"msg"`
	Error string `json:"error"`
	Iplists []Data `json:"iplists"`
}

type Data struct {
	Ip string `json:"ip"`
	Status string `json:"status"`
	DelayName string `json:"delay"`
	DelayData string `json:"delaydata"`
	Group string `json:"group"`
}

type Mastejson struct {
	Msg string `json:"msg"`
	Error string `json:"error"`
	Master string `json:"master"`
	Else string `json:"else"`

}

type Leg struct {
	Msg string `json:"msg"`
	Error string `json:"error"`
}
