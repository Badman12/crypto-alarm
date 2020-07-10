/*
 * Author: Anton Volokha
 * Copyright (c) 2020
 */

package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "time"

  "github.com/faiface/beep"
  "github.com/faiface/beep/mp3"
  "github.com/faiface/beep/speaker"
  "github.com/akamensky/argparse"
)

/**
 * Constants variables
 */
const (
  TOKEN   = "29c1d91366a8dec4292e5f57d0173b5ef8953b3e33e9f82ed3b5ff84c1c5bff8"
  API_URL = "https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=%s&api_key=%s"
  SELL_WHEN = 270
)

/**
 * Coin data struct
 */
type CoinData struct {
  Usd float64 `json:"USD"`
  Uah float64 `json:"UAH"`
  Eur float64 `json:"EUR"`
  danger   *chan bool
}

/**
 * Create empty CoinData object
 */
func createCoinData(danger *chan bool) *CoinData {
  return &CoinData{
    Usd: 0,
    Uah: 0,
    Eur: 0,
    danger:   danger,
  }
}

/**
 * Check on signal or not
 */
func (c *CoinData) check(remoteData *CoinData) bool {
  return remoteData.Usd >= SELL_WHEN
}

/**
 * Update coinData parameters
 */
func (c *CoinData) saveNewParams(remoteData *CoinData) {
  c.Usd = remoteData.Usd
  c.Uah = remoteData.Uah
  c.Eur = remoteData.Eur
}

/**
 * Parse http response and update data
 */
func (c *CoinData) request(client http.Client, url string) {
  resp, err := client.Get(url)
  if err != nil {
    errorLog(err)
    return
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    errorLog(err)
    return
  }

  var buffer CoinData

  json.Unmarshal(body, &buffer)

  if c.check(&buffer) {
    *c.danger <- true
    c.saveNewParams(&buffer)
    return
  }

  c.saveNewParams(&buffer)
}

/**
 * Play alarm sound
 */
func playSound() error {
  f, err := os.Open("./sounds/alarm.mp3")
  if err != nil {
    errorLog(err)
    return err
  }

  streamer, format, err := mp3.Decode(f)
  if err != nil {
    errorLog(err)
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

func main() {
  setupLogger()

  parser := argparse.NewParser("default", "Program to test default values")
  coin := parser.String("c", "coin", &argparse.Options{Required: false, Help: "Coin to check", Default: "ETH"})
  currency := parser.String("o", "currency", &argparse.Options{Required: false, Help: "Currency to output", Default: "USD,UAH,EUR"})

  // Parse input
	err := parser.Parse(os.Args)
	if err != nil {
    errorLog(err)
		os.Exit(1)
	}

  danger := make(chan bool, 0)
  coinData := createCoinData(&danger)

  url := fmt.Sprintf(API_URL, *coin, *currency, TOKEN)

  go func() {
    client := http.Client {
      Timeout: 5 * time.Second,
    }

    for {
      coinData.request(client, url)
      infoLog(coinData)

      time.Sleep(5 * time.Minute)
    }
  }()

  go func() {
    for {
      <-*coinData.danger
      dangerLog(coinData)
      playSound()
    }
  }()

  for {
  }
}
