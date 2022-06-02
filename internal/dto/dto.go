package dto

type Comment struct {
	Comment string `json:"comment"`
	ItemId  int32  `json:"itemId"`
	UserId  int32  `json:"userId"`
}
