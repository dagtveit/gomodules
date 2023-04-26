package opareceiver

import (
	"time"
)

type opaDecPayload struct {
	DecisionId string `json:"decision_id" msg:"decision_id"`
	Input      struct {
		Attributes struct {
			Request struct {
				HTTP struct {
					Headers map[string]string `json:"headers" msg:"headers"`
				} `json:"http" msg:"http"`
			} `json:"request" msg:"request"`
		} `json:"attributes" msg:"attributes"`
	} `json:"input" msg:"input"`
	Timestamp time.Time `json:"timestamp" msg:"timestamp"`
	Result    struct {
		Allowed              string            `json:"result" msg:"result"`
		HttpStatus           string            `json:"http_status" msg:"http_status"`
		ResponseHeadersToAdd map[string]string `json:"response_headers_to_add" msg:"response_headers_to_add"`
		Headers              map[string]string `json:"headers" msg:"headers"`
		Meta                 map[string]string `json:"meta" msg:"meta"`
	} `json:"result" msg:"result"`
}
