package entity

// 回覆狀態
const (
	StateListening = "listening"
	StateResult    = "result"
)

// 指令
const (
	ActionStart = "start"
	ActionStop  = "stop"
)

// Error Code
const (
	ErrOK = 0 - iota // 正確
	ErrFork
	ErrParamInvalid   // 參數錯誤
	ErrLongTimeNoData // 太久沒有傳送資料
	ErrServerFails    // 系統錯誤
	ErrCanNotUse      // 目前系統無法使用
	ErrNoResult       // 沒有辨識結果
)

type RecognizeWord struct {
	StartTime  int    `json:"start_time,omitempty"`
	EndTime    int    `json:"end_time,omitempty"`
	Confidence int    `json:"confidence,omitempty"`
	Word       string `json:"word,omitempty"`
}

// 辨識結果 (json)
type Response struct {
	ErrCode       int             `json:"err_code"`
	ErrMsg        string          `json:"err_msg,omitempty"`
	State         string          `json:"state,omitempty"`
	OriBestResult []RecognizeWord `json:"ori_best_result,omitempty"`
	RecogNbest    []string        `json:"recog_nbest,omitempty"`
	IsFinish      bool            `json:"isFinish,omitempty"`
	ResultIndex   int             `json:"result_index,omitempty"`
	RecogWord     []RecognizeWord `json:"recog_word,omitempty"`
	RecogResult   string          `json:"recog_result,omitempty"`
}

// 傳給辨識的指令 (json)
type Action struct {
	Action         string `json:"action,omitempty"`
	Domain         string `json:"domain,omitempty"`
	Platform       string `json:"platform,omitempty"`
	UID            string `json:"uid,omitempty"`
	IsGetPartial   bool   `json:"isGetPartial,omitempty"`
	IsGetWordInfo  bool   `json:"isGetWordInfo,omitempty"`
	NBestNum       int    `json:"nBestNum,omitempty"`
	Type           string `json:"type,omitempty"`
	Token          string `json:"token,omitempty"`
	IsDoEPD        bool   `json:"bIsDoEPD"`
	PCMD           string `json:"pcmd,omitempty"`
	Text           string `json:"text,omitempty"`
	RejectionLevel int    `json:"rejectionLevel,omitempty"`
}
