package main

import (
    "fmt"
    "os"
    "strings"
    "encoding/json"
    "io"
    "io/ioutil"
    "log"
    "net/http"
)

type Toon struct {

    Data []struct{
        Name    string `json:"name"`
        Account string `json:"account"`
    } `json:"data"`
}

func artifactsPost(PATH string, BODY io.Reader) []byte {
    return artifactsRest("POST", PATH, BODY);
};

func artifactsGet(PATH string) []byte {
    return artifactsRest("GET", PATH, nil);
};

func artifactsRest(ACTION string, PATH string, PAYLOAD io.Reader) []byte {
    ARTIFACTS_API_KEY := os.Getenv("ARTIFACTS_API_KEY");
    ARTIFACTS_BASE_URL := "https://api.artifactsmmo.com/";
    client := &http.Client{};
    req, err := http.NewRequest(ACTION, ARTIFACTS_BASE_URL+PATH, PAYLOAD);
    req.Header.Add("Authorization","Bearer "+ARTIFACTS_API_KEY); 
    res, err := client.Do(req);

    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Fatal(err)
    }

    return responseData

}

func main() {
    fmt.Println("Troy's Artifacts Runner")
    ARTIFACTS_API_KEY := os.Getenv("ARTIFACTS_API_KEY")
    fmt.Println("ARTIFACTS_API_KEY:", ARTIFACTS_API_KEY)
    myToons := artifactsGet("my/characters");
    fmt.Println(string(myToons))

    var toons Toon
    err := json.Unmarshal(myToons, &toons)
    if err != nil {
        panic(err)
    }

    fmt.Println(toons);
    
    var doesToon1Exist bool;
    toon1Name := "Troy";
    
    for _, toon := range toons.Data {
        if toon.Name == toon1Name {
            doesToon1Exist = true;
        }
    }

    if !doesToon1Exist {
        toon1 := artifactsPost("characters/create",strings.NewReader("{\n  \"name\": \""+toon1Name+"\",\n  \"skin\": \"men2\"\n}"));
        fmt.Println(string(toon1));
    } else {
        fmt.Println(toon1Name+" Already Exists!");
    }

}
