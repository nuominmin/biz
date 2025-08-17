package types

type Upload struct {
	Url      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}
