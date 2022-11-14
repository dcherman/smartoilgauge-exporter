package models

type TankDetailsList struct {
	Result  string                `json:"result"`
	Tanks   []TankDetailsResponse `json:"tanks"`
	Message string                `json:"message"`
}

type TankDetailsResponse struct {
	Result         string      `json:"result"`
	TankID         string      `json:"tank_id"`
	TankName       string      `json:"tank_name"`
	ZipCode        string      `json:"zip_code"`
	LowLevel       string      `json:"low_level"`
	SensorGallons  string      `json:"sensor_gallons"`
	SensorRt       string      `json:"sensor_rt"`
	Message        string      `json:"message"`
	ModelGallons   interface{} `json:"model_gallons"`
	Ddd            interface{} `json:"ddd"`
	Nominal        string      `json:"nominal"`
	Fillable       string      `json:"fillable"`
	SensorUsg      string      `json:"sensor_usg"`
	CanFs          string      `json:"can_fs"`
	PtID           interface{} `json:"pt_id"`
	BrandLocked    interface{} `json:"brand_locked"`
	ModelMonitored string      `json:"model_monitored"`
	Battery        string      `json:"battery"`
	SensorStatus   string      `json:"sensor_status"`
	LastRead       string      `json:"last_read"`
	Deadzone       string      `json:"deadzone"`
	DataServers    string      `json:"data_servers"`
	Sensors        []struct {
		RegistrationID   string      `json:"registration_id"`
		SensorID         string      `json:"sensor_id"`
		Description      string      `json:"description"`
		RegistrationDate string      `json:"registration_date"`
		EndDate          interface{} `json:"end_date"`
		LRead            string      `json:"l_read"`
		LGallons         string      `json:"l_gallons"`
		LBattery         string      `json:"l_battery"`
		LPulse           string      `json:"l_pulse"`
		Rlm              string      `json:"rlm"`
		Plm              string      `json:"plm"`
		Ulm              string      `json:"ulm"`
		Ublf             interface{} `json:"ublf"`
		Rblf             interface{} `json:"rblf"`
		Usg              string      `json:"usg"`
		Nrbf             string      `json:"nrbf"`
	} `json:"sensors"`
	Buy struct {
		BrandID    int    `json:"brand_id"`
		BrandName  string `json:"brand_name"`
		CanBuy     string `json:"can_buy"`
		Pro        string `json:"pro"`
		BrandBtn   string `json:"brand_btn"`
		LinkBtn    string `json:"link_btn"`
		BuyBtn     string `json:"buy_btn"`
		BrandEp    string `json:"brand_ep"`
		LinkEp     string `json:"link_ep"`
		BuyEp      string `json:"buy_ep"`
		BrandEpExt int    `json:"brand_ep_ext"`
		CanSell    int    `json:"can_sell"`
	} `json:"buy"`
	Status int `json:"Status"`
}
