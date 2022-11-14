package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/dcherman/smartoilgauge-exporter/models"
)

var (
	sensorGallonsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sensor_gallons",
		Help: "The number of gallons detected by the sensor",
	}, []string{"tank_id", "tank_name", "zipcode"})

	nominalTankGallonsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nominal_tank_gallons",
		Help: "The size of the oil tank",
	}, []string{"tank_id", "tank_name", "zipcode"})

	fillableTankGallonsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "fillable_tank_gallons",
		Help: "The fillable size of the oil tank",
	}, []string{"tank_id", "tank_name", "zipcode"})

	httpClient = &http.Client{}
)

func login(username string, password string) error {
	// 1.  Load the login page
	loginPageResponse, err := httpClient.Get("https://app.smartoilgauge.com/login.php")

	if err != nil {
		return fmt.Errorf("failed to create http request to login page: %v", err)
	}

	defer loginPageResponse.Body.Close()

	if loginPageResponse.StatusCode != 200 {
		return fmt.Errorf("expected status 200 when loading login page, got %d", loginPageResponse.StatusCode)
	}

	// 2: Retrive the nonce that's needed

	doc, err := goquery.NewDocumentFromReader(loginPageResponse.Body)

	if err != nil {
		return fmt.Errorf("failed to parse login page response: %v", err)
	}

	nonce, exists := doc.Find(`input[name="ccf_nonce"]`).Attr("value")

	if !exists {
		return fmt.Errorf("failed to find nonce in login page response")
	}

	// 3: Post the credentials to login

	loginResponse, err := httpClient.PostForm("https://app.smartoilgauge.com/login.php", url.Values{
		"username":  []string{username},
		"user_pass": []string{password},
		"ccf_nonce": []string{nonce},
	})

	if err != nil {
		return fmt.Errorf("failed to create http request to send login request: %v", err)
	}

	defer loginResponse.Body.Close()

	if loginResponse.StatusCode >= 400 {
		return fmt.Errorf("failed to login, got status %v", loginResponse.StatusCode)
	}

	return nil
}

func recordTankDetailsMetrics(username string, password string) {
	logrus.Debug("scraping url")

	query := url.Values{
		"action":  []string{"get_tanks_list"},
		"tank_id": []string{"0"},
	}

	response, err := httpClient.PostForm("https://app.smartoilgauge.com/ajax/main_ajax.php", query)

	if err != nil {
		logrus.Errorf("failed to create http response: %v", err)
		return
	}

	defer response.Body.Close()

	var tankListResponse models.TankDetailsList

	if err := json.NewDecoder(response.Body).Decode(&tankListResponse); err != nil {
		logrus.Errorf("failed to decode response: %v", err)
		return
	}

	if response.StatusCode == 401 || response.StatusCode == 403 || tankListResponse.Message == "Access Denied" {
		logrus.Infof("authentication required.  logging in")

		jar, _ := cookiejar.New(nil)

		httpClient.Jar = jar

		if err := login(username, password); err != nil {
			logrus.Errorf("failed to login: %v", err)
			return
		} else {
			logrus.Info("successfully logged in")
		}

		request, _ := http.NewRequest("POST", "https://app.smartoilgauge.com/ajax/main_ajax.php", bytes.NewBufferString(query.Encode()))

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		request.Header.Set("X-Requested-With", "XMLHttpRequest")

		response, err = httpClient.Do(request)

		if err != nil {
			logrus.Errorf("failed to create http response: %v", err)
			return
		}

		defer response.Body.Close()
	}

	if response.StatusCode != 200 {
		logrus.Errorf("expected status code 200, got %d", response.StatusCode)
		return
	}

	bodyBytes, _ := io.ReadAll(response.Body)

	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&tankListResponse); err != nil {
		logrus.Errorf("failed to decode response: %v", err)
		return
	}

	if tankListResponse.Result == "error" {
		logrus.Errorf("failed to retrieve tank details: %v", string(bodyBytes))
		return
	}

	for _, tank := range tankListResponse.Tanks {
		sensorGallons, err := strconv.ParseFloat(tank.SensorGallons, 64)

		if err != nil {
			logrus.Errorf("failed to convert sensor gallons to float: %v", err)
		} else {
			sensorGallonsMetric.WithLabelValues(tank.TankID, tank.TankName, tank.ZipCode).Set(sensorGallons)
		}

		nominalGallons, err := strconv.ParseFloat(tank.Nominal, 64)

		if err != nil {
			logrus.Errorf("failed to convert nominal gallons to float: %v", err)
		} else {
			nominalTankGallonsMetric.WithLabelValues(tank.TankID, tank.TankName, tank.ZipCode).Set(nominalGallons)
		}

		fillableGallons, err := strconv.ParseFloat(tank.Fillable, 64)

		if err != nil {
			logrus.Errorf("failed to convert fillable gallons to float: %v", err)
		} else {
			fillableTankGallonsMetric.WithLabelValues(tank.TankID, tank.TankName, tank.ZipCode).Set(fillableGallons)
		}
	}

}

func main() {
	port := pflag.Int("port", 8000, "The port to listen on")
	username := pflag.String("username", "", "The username for app.smartoilgauge.com")
	password := pflag.String("password", "", "The password for app.smartoilgauge.com")

	scrapeInterval := pflag.Duration("scrape-interval", time.Minute*15, "The interval at which to scrape the URL")
	metricsPath := pflag.String("metrics-path", "/metrics", "The path to serve metrics on")

	pflag.Parse()

	if *username == "" {
		*username = os.Getenv("SMARTOILGAUGE_USERNAME")

		if *username == "" {
			panic("--username is a required flag")
		}
	}

	if *password == "" {
		*password = os.Getenv("SMARTOILGAUGE_PASSWORD")

		if *password == "" {
			panic("--password is a required flag")
		}
	}

	go func() {
		for {
			recordTankDetailsMetrics(*username, *password)
			time.Sleep(*scrapeInterval)
		}
	}()

	http.Handle(*metricsPath, promhttp.Handler())

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)

	logrus.Error(err)
}
