package trillianlambda

import "github.com/google/trillian"

type LogRootVerificationData struct {
	SignedLogRoot trillian.SignedLogRoot `json:"signed_log_root"`
	PublicKey     string                 `json:"public_key"`
}
