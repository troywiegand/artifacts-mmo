package main

import (
    "fmt"
    "os"
    "strings"
    "strconv"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "time"
    "slices"
)

type Location string
var (
    Spawn Location = "{\"x\":0,\"y\":0}"
    Chicken Location = "{\"x\":0,\"y\":1}"
    CopperMine Location = "{\"x\":2,\"y\":0}"
    IronMine Location = "{\"x\":1,\"y\":7}"
    Forge Location = "{\"x\":1,\"y\":5}"
    WeaponSmith Location = "{\"x\":2,\"y\":1}"
    Bank Location = "{\"x\":4,\"y\":1}"
    AshWood Location = "{\"x\":-1,\"y\":0}"
    SpruceWood Location = "{\"x\":1,\"y\":9}"
    Sawmill Location = "{\"x\":-2,\"y\":-3}"
    YellowSlime Location = "{\"x\":4,\"y\":-1}"
    BlueSlime Location = "{\"x\":2,\"y\":-1}"
    RedSlime Location = "{\"x\":1,\"y\":-1}"
    MonsterTask Location = "{\"x\":1,\"y\":2}"
    Sunflower Location = "{\"x\":2,\"y\":2}"
    Alchemist Location = "{\"x\":2,\"y\":3}"
    GudgeonPond Location = "{\"x\":4,\"y\":2}"
)

type InventoryItem struct {
    Slot int `json:"slot"`
    Code string `json:"code"`
    Quantity int `json:"quantity"`
}

type Monster struct {
    Data struct{
        Level int `json:"level"`
    } `json:"data"`
}

type MapLocation struct {
    Data []struct{
        XPos int `json:"x"`
        YPos int `json:"y"`
    } `json:"data"`
}

type Toon struct {
    Data []ToonDetails `json:"data"`
}

type ToonAction struct {
    Data struct{
        Cooldown struct{
            RemainingSeconds float32 `json:"remaining_seconds"`
        } `json:"cooldown"`
        Character ToonDetails `json:"character"`
    } `json:"data"`
}

type ToonDetails struct {
        Name    string `json:"name"`
        Account string `json:"account"`
        Level   int `json:"level"`
        XPos    int `json:"x"`
        YPos    int `json:"y"`
        Task    string `json:"task"`
        TaskTotal int `json:"task_total"`
        TaskProgress int `json:"task_progress"`
        WoodcuttingLevel int `json:"woodcutting_level"`
        Inventory []InventoryItem `json:"inventory"`
}

func artifactsMove(PATH string, BODY Location) []byte {
    return artifactsRest("POST", PATH, string(BODY));
};

func artifactsPost(PATH string, BODY string) []byte {
    return artifactsRest("POST", PATH, BODY);
};

func artifactsGet(PATH string) []byte {
    return artifactsRest("GET", PATH, "");
};

func artifactsRest(ACTION string, PATH string, PAYLOAD string) []byte {
    ARTIFACTS_API_KEY := os.Getenv("ARTIFACTS_API_KEY");
    ARTIFACTS_BASE_URL := "https://api.artifactsmmo.com/";
    client := &http.Client{};
    req, err := http.NewRequest(ACTION, ARTIFACTS_BASE_URL+PATH, strings.NewReader(PAYLOAD));
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

    fmt.Println(res.Status, ACTION, PATH);

    if strings.Contains(PATH, "action") {
        var ta ToonAction;
        json.Unmarshal(responseData, &ta);
        fmt.Println("Waiting for",ta.Data.Cooldown.RemainingSeconds);
        time.Sleep(time.Duration(ta.Data.Cooldown.RemainingSeconds)*time.Second);
    }

    return responseData

}

func BankDeposit(ToonName string) {
    td := GetInfoFor(ToonName);
    artifactsMove("my/"+ToonName+"/action/move",Bank);
    for i:= 0; i<len(td.Inventory); i++ {
        if td.Inventory[i].Quantity > 0 {
            artifactsPost("my/"+ToonName+"/action/bank/deposit","{\"code\":\""+td.Inventory[i].Code+"\",\"quantity\":"+strconv.Itoa(td.Inventory[i].Quantity)+"}");
        }
    }
    artifactsMove("my/"+ToonName+"/action/move", Location("{\"x\":"+strconv.Itoa(td.XPos)+",\"y\":"+strconv.Itoa(td.YPos)+"}")); 

}

func FightThe(MonsterLocation Location, ToonName string, HowMany int){
    numberToFight := HowMany;
    numberFought := 0;
    if HowMany < 0 {
        ThisToon := GetInfoFor(ToonName);
        numberToFight = ThisToon.TaskTotal
        numberFought = ThisToon.TaskProgress
    }
    fmt.Println(ToonName+" Fights!", numberFought, numberToFight, MonsterLocation);
    artifactsPost("my/"+ToonName+"/action/rest","");
    artifactsMove("my/"+ToonName+"/action/move",MonsterLocation);
    for c := numberFought; c<=numberToFight; c++ {
        artifactsPost("my/"+ToonName+"/action/fight","");
        artifactsPost("my/"+ToonName+"/action/rest","");
        // TO-DO: This Inventory check should be handled by the Rest Handler
        if c % 25 == 24 {
            BankDeposit(ToonName);
        }
    }
}

func GatherThe(Item1 string, Place1 Location, ToonName string) {
    for c := 1; c>0; c++ {
        artifactsMove("my/"+ToonName+"/action/move",Place1);
        for m := AmountOf(Item1, ToonName); m<50; m++ {
            artifactsPost("my/"+ToonName+"/action/gathering","");
        }
        BankDeposit(ToonName);
    }
}

func GatherAndCraftThe(Item1 string, Place1 Location, Item2 string, Place2 Location, ToonName string) {
    for c := 1; c>0; c++ {
        artifactsMove("my/"+ToonName+"/action/move",Place1);
        for m := AmountOf(Item1, ToonName); m<30; m++ {
            artifactsPost("my/"+ToonName+"/action/gathering","");
        }
        artifactsMove("my/"+ToonName+"/action/move",Place2);
        artifactsPost("my/"+ToonName+"/action/crafting","{\"code\":\""+Item2+"\",\"quantity\":3}");
        if AmountOf(Item2, ToonName) >= 20 {
            BankDeposit(ToonName);
        }
            
    }
}

func AmountOf(itemName string, ToonName string) int {
    ThisToon := GetInfoFor(ToonName);
    idx := slices.IndexFunc(ThisToon.Inventory, func(i InventoryItem) bool { return i.Code == itemName })
    var AmountOfItem int;
    if idx == -1 {
        AmountOfItem = 0;
    } else {
        AmountOfItem = ThisToon.Inventory[idx].Quantity 
    }
    return AmountOfItem
}

func GetInfoFor(ToonName string) ToonDetails {
    myToons := artifactsGet("my/characters");
    var toons Toon
    json.Unmarshal(myToons, &toons)
    idx := slices.IndexFunc(toons.Data, func(td ToonDetails) bool { return td.Name == ToonName })
    return toons.Data[idx];
};

func RunMonsterTasks(ToonName string) {
    for c := 1; c>0; c++ {
        //Check For Task
        t := GetInfoFor(ToonName);
        if t.Task == "" {
            artifactsMove("my/"+ToonName+"/action/move",MonsterTask);
            // Get New Task
            artifactsPost("my/"+ToonName+"/action/task/new","");
            t = GetInfoFor(ToonName);
        }
        
        if t.Task != "" || t.TaskProgress == t.TaskTotal {
            artifactsMove("my/"+ToonName+"/action/move",MonsterTask);
            artifactsPost("my/"+ToonName+"/action/task/complete","");
            artifactsPost("my/"+ToonName+"/action/task/new","");
            t = GetInfoFor(ToonName);
        }
            
        //Monster Level
        monsterInfo := artifactsGet("monsters/"+t.Task);
        var mon Monster;
        json.Unmarshal(monsterInfo, &mon);
        if (mon.Data.Level > t.Level-5) {
            c = -1;
            break;
        } else {
            var mapLoc MapLocation;
            monsterLocation := artifactsGet("maps?content_code="+t.Task+"&size=1");
            json.Unmarshal(monsterLocation, &mapLoc);
            fmt.Println("monsterLocation:", string(monsterLocation));
            fmt.Println(mapLoc);
            FightThe(Location("{\"x\":"+strconv.Itoa(mapLoc.Data[0].XPos)+",\"y\":"+strconv.Itoa(mapLoc.Data[0].YPos)+"}"), ToonName, -1); 
        }
    }

}


func main() {
    fmt.Println("Troy's Artifacts Runner")
    ARTIFACTS_API_KEY := os.Getenv("ARTIFACTS_API_KEY")
    fmt.Println("ARTIFACTS_API_KEY:", ARTIFACTS_API_KEY)
    myToons := artifactsGet("my/characters");
    var toons Toon
    err := json.Unmarshal(myToons, &toons)
    if err != nil {
        panic(err)
    }
    fmt.Println(toons);
    
    var doesToon1Exist bool;
    toon1Name := "Troy";
    var doesToon2Exist bool;
    toon2Name := "Faraday";
    var doesToon3Exist bool;
    toon3Name := "Rainboom";
    var doesToon4Exist bool;
    toon4Name := "Ikhor";
    var doesToon5Exist bool;
    toon5Name := "Crydelia";

    for _, toon := range toons.Data {
        if toon.Name == toon1Name {
            doesToon1Exist = true;
        }
        if toon.Name == toon2Name {
            doesToon2Exist = true;
        }
        if toon.Name == toon3Name {
            doesToon3Exist = true;
        }
        if toon.Name == toon4Name {
            doesToon4Exist = true;
        }
        if toon.Name == toon5Name {
            doesToon5Exist = true;
        }
    }
    
    if !doesToon1Exist {
        toon1 := artifactsPost("characters/create","{\n  \"name\": \""+toon1Name+"\",\n  \"skin\": \"men2\"\n}");
        fmt.Println(string(toon1));
        doesToon1Exist=true;
    }
    if !doesToon2Exist {
        toon2 := artifactsPost("characters/create","{\n  \"name\": \""+toon1Name+"\",\n  \"skin\": \"men1\"\n}");
        fmt.Println(string(toon2));
        doesToon2Exist=true;
    }

    if doesToon1Exist && doesToon2Exist && doesToon3Exist && doesToon4Exist && doesToon5Exist {
        go GatherAndCraftThe("spruce_wood", SpruceWood, "spruce_plank", Sawmill, toon2Name);
        go GatherAndCraftThe("copper_ore", CopperMine, "copper", Forge, toon3Name);
        go GatherAndCraftThe("copper_ore", CopperMine, "copper", Forge, toon5Name);
        go GatherAndCraftThe("ash_wood", AshWood, "ash_plank", Sawmill, toon4Name);
        GatherAndCraftThe("iron_ore", IronMine, "iron", Forge, toon1Name);
    }
}
