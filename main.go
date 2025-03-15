package main

import (
    "os"
    "strings"
    "strconv"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "time"
    "slices"
)
var logger = GetLogger();
type Location string
const (
    Spawn Location = "{\"x\":0,\"y\":0}"
    Chicken Location = "{\"x\":0,\"y\":1}"
    CopperMine Location = "{\"x\":2,\"y\":0}"
    IronMine Location = "{\"x\":1,\"y\":7}"
    CoalMine Location = "{\"x\":1,\"y\":6}"
    Forge Location = "{\"x\":1,\"y\":5}"
    WeaponSmith Location = "{\"x\":2,\"y\":1}"
    Bank Location = "{\"x\":4,\"y\":1}"
    AshWood Location = "{\"x\":-1,\"y\":0}"
    SpruceWood Location = "{\"x\":1,\"y\":9}"
    BirchWood Location = "{\"x\":-1,\"y\":6}"
    Sawmill Location = "{\"x\":-2,\"y\":-3}"
    YellowSlime Location = "{\"x\":4,\"y\":-1}"
    BlueSlime Location = "{\"x\":2,\"y\":-1}"
    RedSlime Location = "{\"x\":1,\"y\":-1}"
    Mushmush Location = "{\"x\":5,\"y\":3}"
    MonsterTask Location = "{\"x\":1,\"y\":2}"
    Sunflower Location = "{\"x\":2,\"y\":2}"
    Alchemist Location = "{\"x\":2,\"y\":3}"
    GudgeonPond Location = "{\"x\":4,\"y\":2}"
)

type ToonName string
const (
    Troy ToonName = "Troy"
    Faraday ToonName = "Faraday"
    Rainboom ToonName = "Rainboom"
    Ikhor ToonName = "Ikhor"
    Crydelia ToonName = "Crydelia"
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
        Name    ToonName `json:"name"`
        Account string `json:"account"`
        Level   int `json:"level"`
        XPos    int `json:"x"`
        YPos    int `json:"y"`
        Cooldown int `json:"cooldown"`
        Task    string `json:"task"`
        TaskTotal int `json:"task_total"`
        TaskProgress int `json:"task_progress"`
        WoodcuttingLevel int `json:"woodcutting_level"`
        MiningLevel int `json:"mining_level"`
        FishingLevel int `json:"fishing_level"`
        AlchemyLevel int `json:"alchemy_level"`
        Inventory []InventoryItem `json:"inventory"`
}

func artifactsMove(td ToonDetails, BODY Location) []byte {
    // TO-DO: Don't Try To Move If Already There
    p := "my/"+string(td.Name)+"/action/move";
    return artifactsRest("POST", p, string(BODY));
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
        logger.Error(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(res.Body)
    if err != nil {
        logger.Fatal(err)
    }

    logger.Info(res.Status, ACTION, PATH);

    if res.StatusCode != 490 {
    	if strings.Contains(PATH, "action") {
        	var ta ToonAction;
        	json.Unmarshal(responseData, &ta);
        	logger.Info("SLEEP","ToonName",ta.Data.Character.Name,"seconds",ta.Data.Cooldown.RemainingSeconds);
        	time.Sleep(time.Duration(ta.Data.Cooldown.RemainingSeconds)*time.Second);
    	}
    }

    return responseData

}

func BankDeposit(ToonName ToonName) {
    td := GetInfoFor(ToonName);
    artifactsMove(td,Bank);
    for i:= 0; i<len(td.Inventory); i++ {
        if td.Inventory[i].Quantity > 0 {
            artifactsPost("my/"+string(ToonName)+"/action/bank/deposit","{\"code\":\""+td.Inventory[i].Code+"\",\"quantity\":"+strconv.Itoa(td.Inventory[i].Quantity)+"}");
        }
    }
    artifactsMove(td, Location("{\"x\":"+strconv.Itoa(td.XPos)+",\"y\":"+strconv.Itoa(td.YPos)+"}")); 

}

func FightThe(MonsterLocation Location, ToonName ToonName, HowMany int){
    numberToFight := HowMany;
    numberFought := 0;
    ThisToon := GetInfoFor(ToonName);
    if HowMany < 0 {
        numberToFight = ThisToon.TaskTotal
        numberFought = ThisToon.TaskProgress
    }
    logger.Info("FIGHT", "ToonName", ToonName, numberFought, numberToFight, MonsterLocation);
    artifactsPost("my/"+string(ToonName)+"/action/rest","");
    artifactsMove(ThisToon,MonsterLocation);
    for c := numberFought; c<=numberToFight; c++ {
        artifactsPost("my/"+string(ToonName)+"/action/fight","");
        artifactsPost("my/"+string(ToonName)+"/action/rest","");
        // TO-DO: This Inventory check should be handled by the Rest Handler
        if c % 25 == 24 {
            BankDeposit(ToonName);
        }
    }
}

func GatherThe(Item1 string, Place1 Location, ToonName ToonName) {
    for c := 1; c>0; c++ {
        artifactsMove(GetInfoFor(ToonName),Place1);
        for m := AmountOf(Item1, ToonName); m<50; m++ {
            artifactsPost("my/"+string(ToonName)+"/action/gathering","");
        }
        BankDeposit(ToonName);
    }
}

func GatherAndCraftThe(Item1 string, Place1 Location, Item2 string, Place2 Location, ToonName ToonName, HowManyLoops int) {
    var l int;
    if HowManyLoops <= 0 {
        l = 10000
    } else {
        l = HowManyLoops
    }
    td := GetInfoFor(ToonName);
    for c := 0; c<l; c++ {
        artifactsMove(td,Place1);
        for m := AmountOf(Item1, ToonName); m<30; m++ {
            artifactsPost("my/"+string(ToonName)+"/action/gathering","");
        }
        artifactsMove(td,Place2);
        artifactsPost("my/"+string(ToonName)+"/action/crafting","{\"code\":\""+Item2+"\",\"quantity\":3}");
        if AmountOf(Item2, ToonName) >= 20 {
            BankDeposit(ToonName);
        }
            
    }
}

func AmountOf(itemName string, ToonName ToonName) int {
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

func GetInfoFor(ToonName ToonName) ToonDetails {
    myToons := artifactsGet("my/characters");
    var toons Toon
    json.Unmarshal(myToons, &toons)
    idx := slices.IndexFunc(toons.Data, func(td ToonDetails) bool { return td.Name == ToonName })
    return toons.Data[idx];
};

func RunMonsterTasks(ToonName ToonName) {
    for c := 1; c>0; c++ {
        //Check For Task
        t := GetInfoFor(ToonName);
        if t.Task == "" {
            artifactsMove(t,MonsterTask);
            // Get New Task
            artifactsPost("my/"+string(ToonName)+"/action/task/new","");
            t = GetInfoFor(ToonName);
        }
        
        if t.Task != "" || t.TaskProgress == t.TaskTotal {
            artifactsMove(t,MonsterTask);
            artifactsPost("my/"+string(ToonName)+"/action/task/complete","");
            artifactsPost("my/"+string(ToonName)+"/action/task/new","");
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
            logger.Debug("monsterLocation:", string(monsterLocation));
            logger.Debug(mapLoc);
            FightThe(Location("{\"x\":"+strconv.Itoa(mapLoc.Data[0].XPos)+",\"y\":"+strconv.Itoa(mapLoc.Data[0].YPos)+"}"), ToonName, -1); 
        }
    }

}

//TO-DO: Check for DateTime Object `cooldown_expiration`
// If that isn't present use Cooldown as a fallback
// Or potentially if that isn't present just send it
// Cooldown appears to just be the duration of the last cooldown
// Even if it was in the past and is over
func EnsureOffCooldown(Toons Toon, ToonName ToonName) {
    idx := slices.IndexFunc(Toons.Data, func(td ToonDetails) bool { return td.Name == ToonName })
    InitialSleepBuffer := Toons.Data[idx].Cooldown
    if InitialSleepBuffer > 0 {
        logger.Debug(ToonName," needs to wait ",InitialSleepBuffer," before they are ready to rumble!");
        time.Sleep(time.Duration(InitialSleepBuffer)*time.Second);
    } else {
        logger.Debug(ToonName+" is Ready-To-Go!");
    }
}

func main() {
    logger.Info("Troy's Artifacts Runner")
    os.Getenv("ARTIFACTS_API_KEY")
    myToons := artifactsGet("my/characters");
    var toons Toon
    err := json.Unmarshal(myToons, &toons)
    if err != nil {
        panic(err)
    }
    logger.Debug(toons);
    
    go func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
        GatherAndCraftThe("birch_wood", BirchWood, "birch_plank", Sawmill, t, 100);
    }(Faraday);
    go func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
        GatherThe("sunflower", Sunflower, t);
        GatherAndCraftThe("copper_ore", CopperMine, "copper", Forge, t,50);
    }(Rainboom);
    go func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
        GatherAndCraftThe("copper_ore", CopperMine, "copper", Forge, t,100);
        GatherAndCraftThe("iron_ore", IronMine, "iron", Forge, t,-1);
    }(Crydelia);
    go func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
        FightThe(Chicken, t, 800);
        FightThe(YellowSlime, t, 800);
    }(Ikhor);
    func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
        FightThe( Mushmush , t , 500 );
        GatherAndCraftThe("iron_ore", IronMine, "iron", Forge, t, 100);
        GatherThe("coal", CoalMine, t);
    }(Troy)
}
