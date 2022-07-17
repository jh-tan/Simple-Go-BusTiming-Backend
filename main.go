package main

import (
	"encoding/json"
	_ "errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var busStopIds = []string{
	"378204", "383050", "378202", "383049", "382998", "378237", "378233", "378230", "378229", "378228", "378227", "382995", "378224", "378226", "383010", "383009",
	"383006", "383004", "378234", "383003", "378222", "383048", "378203", "382999",
	"378225", "383014", "383013", "383011", "377906", "383018", "383015", "378207",
}

var busLineIds = []string{"44478", "44479", "44480", "44481"}

type busStop struct {
	Name string `json:"busStopName"`
	Lat  string `json:"busStopLatitude"`
	Lon  string `json:"busStopLongitude"`
}

type busStopArrivalInfo struct {
	EstimatedRemainingTime float32 `json:"estimatedRemainingTime"`
	RouteName              string  `json:"routeName"`
	RoutevariantID         int     `json:"rv_id"`
	BusStopInfo            busStop `json:"busStop"`
}

type busStopJSON struct {
	ExternalID string `json:"-"`
	Forecast   []struct {
		EstimatedRemainingTime float32 `json:"forecast_seconds"`
		Route                  struct {
			ID        int    `json:"-"`
			Name      string `json:"-"`
			ShortName string `json:"short_name"`
		} `json:"route"`
		RvID        int     `json:"rv_id"`
		TotalPass   float32 `json:"-"`
		VehicleName string  `json:"vehicle"`
		VehicleId   int     `json:"-"`
	} `json:"forecast"`
	Geometry []struct {
		ExternalID interface{} `json:"-"`
		Lat        string      `json:"lat"`
		Lon        string      `json:"lon"`
		Seq        int         `json:"-"`
	} `json:"geometry"`
	ID          int         `json:"-"`
	Name        string      `json:"name"`
	EngName     interface{} `json:"-"`
	RuName      interface{} `json:"-"`
	Nameslug    string      `json:"-"`
	ResourceURI string      `json:"-"`
}

type busLocationInfo struct {
	RouteName       string `json:"routeName"`
	RoutevariantID  int    `json:"rv_id"`
	BusCurrentLat   string `json:"busCurrentLat"`
	BusCurrentLong  string `json:"busCurrentLong"`
	BusCurrentSpeed string `json:"busCurrentSpeed"`
}

type busLineJSON struct {
	ExternalID  interface{} `json:"-"`
	ID          int         `json:"id"`
	Name        string      `json:"-"`
	EngName     interface{} `json:"-"`
	RuName      interface{} `json:"-"`
	Nameslug    interface{} `json:"-"`
	ResourceURI string      `json:"-"`
	RouteName   string      `json:"routename"`
	Vehicle     []struct {
		Bearing  int    `json:"-"`
		DeviceTS string `json:"-"`
		Route    struct {
			EnterpriseID   int    `json:"-"`
			EnterpriseNAME string `json:"-"`
		} `json:"-"`
		Lat  string `json:"lat"`
		Lon  string `json:"lon"`
		Park struct {
			ParkId   int    `json:"-"`
			ParkName string `json:"-"`
		} `json:"-"`
		Position struct {
			Bearing  int    `json:"-"`
			DeviceTS int    `json:"-"`
			Lat      string `json:"-"`
			Lon      string `json:"-"`
			Speed    int    `json:"-"`
			Ts       int    `json:"-"`
		} `json:"-"`
		Projection struct {
			EdgeDistance    string `json:"-"`
			EdgeId          int    `json:"-"`
			EdgeProjection  string `json:"-"`
			EdgeStartNodeID int    `json:"-"`
			EdgeStopNodeID  int    `json:"-"`
			Lat             string `json:"-"`
			Lon             string `json:"-"`
			OrigLat         string `json:"-"`
			OrigLon         string `json:"-"`
			RoutevariantID  int    `json:"-"`
			Ts              int    `json:"-"`
		} `json:"projection"`
		RegistrationCode string `json:"-"`
		RoutevariantID   int    `json:"-"`
		Speed            string `json:"speed"`
		Stats            struct {
			AvgSpeed    string `json:"-"`
			Bearing     int    `json:"-"`
			CummSpeed10 string `json:"-"`
			CummSpeed2  string `json:"-"`
			DeviceTS    int    `json:"-"`
			Lat         string `json:"-"`
			Lon         string `json:"-"`
			Speed       int    `json:"-"`
			Ts          int    `json:"-"`
		} `json:"-"`
		Ts        string `json:"-"`
		VehicleId int    `json:"-"`
	} `json:"vehicles"`
	Via interface{} `json:"-"`
}

type busTrackingInfo struct {
	BusStopArrivalInfo busStopArrivalInfo `json:"busStopArrivalInfo"`
	BusCurrentLat      string             `json:"busCurrentLat"`
	BusCurrentLong     string             `json:"busCurrentLong"`
	BusCurrentSpeed    string             `json:"busCurrentSpeed"`
}

type errorMessage struct {
	code    int
	message string
}

type Server struct {
	// logger Logger
	logic Logic
}

type busInfoService struct{}

type Logic interface {
	fetchBusTimingAPI(id string) (busStopArrivalInfo, error)
	fetchBusLocationAPI(id string) (busLocationInfo, error)
}

func (e errorMessage) Error() string {
	errMessage := fmt.Sprint("Error Code:", e.code, " Error Message:", e.message)
	return errMessage
}

func extractJSONForecastValues(result *busStopJSON) (float32, string, int) {
	if len((*result).Forecast) == 0 {
		return -1, "Unknown Route Name", -1
	} else {
		return (*result).Forecast[0].EstimatedRemainingTime, (*result).Forecast[0].Route.ShortName, (*result).Forecast[0].RvID
	}
}

func (service *busInfoService) fetchBusTimingAPI(id string) (busStopArrivalInfo, error) {
	resp, err := http.Get("https://baseride.com/routes/api/platformbusarrival/" + id + "/?format=json")

	if err != nil || resp.StatusCode == 500 {
		return busStopArrivalInfo{}, errorMessage{
			code:    400,
			message: "No Such ID",
		}
	}

	d := json.NewDecoder(resp.Body)
	var result busStopJSON
	err = d.Decode(&result)

	if err != nil {
		fmt.Println(err)
		return busStopArrivalInfo{}, errorMessage{
			code:    400,
			message: "Failed to parse the result",
		}
	}
	busStopInfo := busStop{
		Name: result.Name,
		Lat:  result.Geometry[0].Lat,
		Lon:  result.Geometry[0].Lon,
	}

	estimatedTime, routeName, rv_id := extractJSONForecastValues(&result)

	arrivalInfo := busStopArrivalInfo{
		EstimatedRemainingTime: estimatedTime,
		RouteName:              routeName,
		RoutevariantID:         rv_id,
		BusStopInfo:            busStopInfo,
	}

	return arrivalInfo, nil
}

func (s Server) getAllBusArrivalTiming(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	allResult := make([]busStopArrivalInfo, 0, 32)
	for _, id := range busStopIds {
		result, _ := s.logic.fetchBusTimingAPI(id)
		allResult = append(allResult, result)
	}

	json.NewEncoder(w).Encode(allResult)
}

func (s Server) getBusArrivalTiming(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result, err := s.logic.fetchBusTimingAPI(params["id"])
	fmt.Println(err)
	if err != nil {
		fmt.Println("No such id")
    fmt.Fprint(w, "No such id")
    return 
	}

	json.NewEncoder(w).Encode(result)
}

func extractJSONVehicleValues(result *busLineJSON) (string, string, string) {
	if len((*result).Vehicle) == 0 {
		return "Bus Latitude Not Detected", "Bus Longitude Not Detected", "Bus Speed Not Detected"
	} else {
		return (*result).Vehicle[0].Lat, (*result).Vehicle[0].Lon, (*result).Vehicle[0].Speed
	}
}

func (service *busInfoService) fetchBusLocationAPI(id string) (busLocationInfo, error) {
	resp, err := http.Get("https://baseride.com/routes/apigeo/routevariantvehicle/" + id + "/?format=json")

	if err != nil || resp.StatusCode == 500 {
		return busLocationInfo{}, errorMessage{
			code:    400,
			message: "No Such ID",
		}
	}

	d := json.NewDecoder(resp.Body)
	var result busLineJSON
	err = d.Decode(&result)

	if err != nil {
		return busLocationInfo{}, errorMessage{
			code:    400,
			message: "Failed to parse the result",
		}
	}

	lat, lon, speed := extractJSONVehicleValues(&result)

	busLocation := busLocationInfo{
		RouteName:       result.RouteName,
		RoutevariantID:  result.ID,
		BusCurrentLat:   lat,
		BusCurrentLong:  lon,
		BusCurrentSpeed: speed,
	}

	return busLocation, nil
}

func (s Server) getBusLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result, err := s.logic.fetchBusLocationAPI(params["id"])
	if err != nil {
		fmt.Println("No such id")
    fmt.Fprint(w, "No such id")
    return 
	}

	json.NewEncoder(w).Encode(result)
}

func (s Server) getAllBusLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	allResult := make([]busLocationInfo, 0, 4)
	for _, id := range busLineIds {
		result, _ := s.logic.fetchBusLocationAPI(id)
		allResult = append(allResult, result)
	}

	json.NewEncoder(w).Encode(allResult)
}

func (s Server) getBusEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	allResult := make([]busTrackingInfo, 0, 32)
	allBusLocs := make(map[int]busLocationInfo, 4)
	for _, id := range busStopIds {
		busStop, _ := s.logic.fetchBusTimingAPI(id)
		allResult = append(allResult, busTrackingInfo{BusStopArrivalInfo: busStop})
	}

	for _, id := range busLineIds {
		busLoc, _ := s.logic.fetchBusLocationAPI(id)
		allBusLocs[busLoc.RoutevariantID] = busLoc
	}


	for i := 0; i < len(allResult); i++ {
		rv_id := allResult[i].BusStopArrivalInfo.RoutevariantID
    if rv_id == -1 {
      allResult[i].BusCurrentLat = "Bus Latitude Not Detected"
      allResult[i].BusCurrentLong = "Bus Longitude Not Detected"
      allResult[i].BusCurrentSpeed = "Bus Speed Not Detected"
    } else {
      allResult[i].BusCurrentLat = allBusLocs[rv_id].BusCurrentLat
      allResult[i].BusCurrentLong = allBusLocs[rv_id].BusCurrentLong
      allResult[i].BusCurrentSpeed = allBusLocs[rv_id].BusCurrentSpeed
    }
	}

	json.NewEncoder(w).Encode(allResult)
}

func (s *Server) handleRequest() {
	router := mux.NewRouter()
	router.HandleFunc("/busstop", s.getAllBusArrivalTiming).Methods("GET")
	router.HandleFunc("/busstop/{id}", s.getBusArrivalTiming).Methods("GET")
	router.HandleFunc("/busline", s.getAllBusLocation).Methods("GET")
	router.HandleFunc("/busline/{id}", s.getBusLocation).Methods("GET")
	router.HandleFunc("/busevents", s.getBusEvents).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func NewServer(logic Logic) Server {
	return Server{
		logic: logic,
	}
}

func NewBusInfoService() *busInfoService {
	return &busInfoService{}
}

func main() {
	busInfoService := NewBusInfoService()
	server := NewServer(busInfoService)
	server.handleRequest()
}
