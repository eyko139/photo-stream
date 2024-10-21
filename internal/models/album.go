package models


type Album struct {
    UID string `json:"UID"`
    Thumb string `json:"Thumb"`
    Title string `json:"Title"`
    Type string `json:"Type"`
    PhotoCount int `json:"PhotoCount"`
    ClassName string
    B64 string  `json:"B64"`
}
