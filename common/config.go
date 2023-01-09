package common

import (
    "io/ioutil"
    "encoding/json"
)

// just reads a file into a string
func ReadFile(filename string) string {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        panic(err)
    }
    return string(bytes)
}

// loads the default configuration file
func LoadConfig() map[string]string {
    rawConfig := ReadFile("./resources/config.json")
    outlet := make(map[string]string)
    json.Unmarshal([]byte(rawConfig), &outlet)
    return outlet
}

