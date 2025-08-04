package domain

/*
	 type Items struct {
		ChrtID     int    `json:"chrt_id"`
		Price      int    `json:"price"`
		Rid        string `json:"rid"`
		Name       string `json:"name"`
		Sale       int    `json:"sale"`
		Size       string `json:"size"`
		TotalPrice int    `json:"total_price"`
		NmID       int    `json:"nm_id"`
		Brand      string `json:"brand"`
	}
*/
type Items struct {
	ChrtID     int    `json:"chrt_id" validate:"required"`
	Price      int    `json:"price" validate:"required,min=0"`
	Rid        string `json:"rid" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Sale       int    `json:"sale" validate:"min=0,max=100"` // Assuming sale is a percentage
	Size       string `json:"size"`
	TotalPrice int    `json:"total_price" validate:"required,min=0"`
	NmID       int    `json:"nm_id" validate:"required"`
	Brand      string `json:"brand" validate:"required"`
}
