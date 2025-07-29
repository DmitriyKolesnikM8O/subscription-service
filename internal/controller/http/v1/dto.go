package v1

type CreateServiceRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Price int `json:"price" validate:"required,gt=0"`
}

type CreateRequest struct {
	Service   CreateServiceRequest `json:"service" validate:"required"`
	UserID    string               `json:"user_id" validate:"required,uuid4"`
	StartDate string               `json:"start_date" validate:"required,datetime=01-2006"`
}

type CalculateTotalCostRequest struct {
	UserID      string `query:"user_id" validate:"omitempty,uuid4"`
	ServiceName string `query:"service_name" validate:"omitempty,min=2,max=100"`
	StartDate   string `query:"start_date" validate:"required,datetime=01-2006"`
	EndDate     string `query:"end_date" validate:"required,datetime=01-2006"`
}

type UpdateServiceRequest struct {
	Name string `json:"name" validate:"omitempty,min=2,max=100"`
	Price int `json:"price" validate:"omitempty,gt=0"`
}

type UpdateRequest struct {
	Service   UpdateServiceRequest `json:"service" validate:"required"`
	EndDate   string               `json:"end_date" validate:"omitempty,datetime=01-2006"`
}

// type ErrorResponse struct {
// 	Message string `json:"message"`
// }

type TotalCostResponse struct {
	Total int `json:"total"`
}
