package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// Логування

type Log struct {
	fileName string
}

func (loger *Log) writeLog(w *Weather) {
	str := fmt.Sprintf("%s\t temp: %.2f\t humidity: %d\t pressure: %d\n",
		time.Now(), w.Temp, w.Humidity, w.Pressure,
	)

	loger.write(str)
}

func (loger *Log) danger(w *Weather) {
	str := fmt.Sprintf("\nТРИВОГА ТРИВОГА ТРИВОГА %s\t temp: %.2f\t humidity: %d\t pressure: %d\n",
		time.Now(), w.Temp, w.Humidity, w.Pressure,
	)

	loger.write(str)
}

func (loger *Log) write(str string) {
	file, err := os.OpenFile(loger.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	fmt.Printf("%s\n", str)
	file.WriteString(str)
}

func createLoger(fileName string) *Log {
	return &Log{
		fileName: fileName,
	}
}

// Основана логіка

const (
	TOKEN   = "56d1ac917f05a465682b89a2ddd05b0a"
	API_URL = "http://api.openweathermap.org/data/2.5/weather?q=city&appid="
)

type BodyResponse struct {
	Main Weather `json:"main"`
}

type MainResponse struct {
	Temp string `json:"temp"`
}

type Weather struct {
	Temp     float64 `json:"temp"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
	init     bool
	danger   *chan bool
}

func createWether(danger *chan bool) *Weather {
	return &Weather{
		Temp:     0,
		Pressure: 0,
		Humidity: 0,
		init:     false,
		danger:   danger,
	}
}

func (w *Weather) check(remoteData *Weather) bool {
	if w.Temp != remoteData.Temp-273.15 {
		return false
	}

	if w.Pressure != remoteData.Pressure {
		return false
	}

	if w.Humidity != remoteData.Humidity {
		return false
	}

	return true
}

func (w *Weather) getWeather() *Weather {
	return w
}

func (w *Weather) saveNewParams(remoteData *Weather) {
	w.Temp = remoteData.Temp - 273.15
	w.Humidity = remoteData.Humidity
	w.Pressure = remoteData.Pressure
	w.init = true
}

func (w *Weather) changeWeather(resp *http.Response) {
	var buffer BodyResponse

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	json.Unmarshal(body, &buffer)

	if w.init == true {
		if w.check(&buffer.Main) {
			return
		}
		*w.danger <- true
		w.saveNewParams(&buffer.Main)
		return
	}
	w.saveNewParams(&buffer.Main)
}

func playSound() error {
	f, err := os.Open("./sounds/alarm.mp3")
	if err != nil {
		log.Fatal(err)
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer streamer.Close()

	sr := format.SampleRate * 2
	speaker.Init(sr, sr.N(time.Second/10))

	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan bool, 0)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done

	return nil
}

func getCity(stream *os.File) string {
	fmt.Printf("Введи назву міста, яке хочеш охороняти xD:\t")

	reader := bufio.NewReader(os.Stdin)
	str, _ := reader.ReadString('\n')
	str = strings.Replace(str, "\n", "", -1)

	return str
}

func main() {
	whiteLoger := createLoger("log.txt")
	danger := make(chan bool, 0)
	weather := createWether(&danger)
	city := getCity(os.Stdin)

	url := strings.Replace(API_URL, "city", city, -1) + TOKEN

	go func() {
		client := http.Client{
			Timeout: 5 * time.Second,
		}

		for {
			resp, err := client.Get(url)
			if err != nil {
				log.Println(err, "\nСталася помилка з мережою")
				os.Exit(0)
			}
			defer resp.Body.Close()

			weather.changeWeather(resp)
			whiteLoger.writeLog(weather)

			time.Sleep(2 * time.Minute)
		}
	}()

	go func() {
		for {
			<-*weather.danger
			whiteLoger.danger(weather)
			playSound()
		}
	}()

	for {
	}
}
