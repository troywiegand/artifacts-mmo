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
    "time"
)

type Toon struct {
    Data []struct{
        Name    string `json:"name"`
        Account string `json:"account"`
    } `json:"data"`
}

type ToonAction struct {
    Data struct{
        Cooldown struct{
            RemainingSeconds float32 `json:"remaining_seconds"`
        } `json:"cooldown"`
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

    var toon1Action ToonAction
    if !doesToon1Exist {
        toon1 := artifactsPost("characters/create",strings.NewReader("{\n  \"name\": \""+toon1Name+"\",\n  \"skin\": \"men2\"\n}"));
        fmt.Println(string(toon1));
    } else {
        fmt.Println(toon1Name+" Already Exists!");
        fmt.Println(toon1Name+" Fights the Chicken!");
        for c := 1; c>0; c++ {
            moveToChicken := artifactsPost("my/"+toon1Name+"/action/move",strings.NewReader("{\"x\":0,\"y\":1}"));
            fmt.Println(string(moveToChicken));
            json.Unmarshal(moveToChicken, &toon1Action);
            time.Sleep(time.Duration(toon1Action.Data.Cooldown.RemainingSeconds)*time.Second);
            fightChicken := artifactsPost("my/"+toon1Name+"/action/fight",nil);
            fmt.Println(string(fightChicken));
            json.Unmarshal(fightChicken, &toon1Action);
            time.Sleep(time.Duration(toon1Action.Data.Cooldown.RemainingSeconds)*time.Second);
            restChicken := artifactsPost("my/"+toon1Name+"/action/rest",nil);
            fmt.Println(string(restChicken));
            json.Unmarshal(restChicken, &toon1Action);
            time.Sleep(time.Duration(toon1Action.Data.Cooldown.RemainingSeconds)*time.Second);
        }
        //// Gathering
        //fmt.Println(toon1Name+" Gathers Wood!");
        //moveToTree := artifactsPost("my/"+toon1Name+"/action/move",strings.NewReader("{\"x\":-1,\"y\":0}"));
        //fmt.Println(string(moveToTree));
        time.Sleep(1*time.Second);
        //for i := 0; i < 4; i++ {
            /// TO-DO: Derive time vs hard coding
        //    GatherWood := artifactsPost("my/"+toon1Name+"/action/gathering",nil);
        //    fmt.Println(string(GatherWood));
        //    json.Unmarshal(GatherWood, &toon1Action);
        //    time.Sleep(time.Duration(toon1Action.Data.Cooldown.RemainingSeconds)*time.Second);
        //}
        //fmt.Println(toon1Name+" Upgrades to Staff!");
        //unequip := artifactsPost("my/"+toon1Name+"/action/unequip",strings.NewReader("{\"slot\":\"weapon\"}"));
        //fmt.Println(string(unequip));
        //json.Unmarshal(unequip, &toon1Action);
        //time.Sleep(time.Duration(toon1Action.Data.Cooldown.RemainingSeconds)*time.Second);
        //moveToCrafter := artifactsPost("my/"+toon1Name+"/action/move",strings.NewReader("{\"x\":2,\"y\":1}"));
        //fmt.Println(string(moveToCrafter));
        //json.Unmarshal(moveToCrafter, &toon1Action);
        //time.Sleep(time.Duration(toon1Action.Data.Cooldown.RemainingSeconds)*time.Second);
        //craftStaff := artifactsPost("my/"+toon1Name+"/action/crafting",strings.NewReader("{\"code\":\"wooden_staff\"}"));
        //fmt.Println(string(craftStaff));
        //json.Unmarshal(craftStaff, &toon1Action);
        //time.Sleep(time.Duration(toon1Action.Data.Cooldown.RemainingSeconds)*time.Second);
        //equipStaff := artifactsPost("my/"+toon1Name+"/action/equip",strings.NewReader("{\"code\":\"wooden_staff\",\"slot\":\"weapon\"}"));
        //fmt.Println(string(equipStaff));
        //json.Unmarshal(equipStaff, &toon1Action);
    }
}
