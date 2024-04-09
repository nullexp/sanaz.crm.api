package response

type Asset struct {
	Id string `json:"id" validate:"required,uuid"`
}
