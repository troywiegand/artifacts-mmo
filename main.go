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
    "fmt"
    "github.com/joho/godotenv"
)
var logger = GetLogger();

type Location struct{
    XPos int `json:"x"`
    YPos int `json:"y"`
}

var (
Spawn Location = Location{XPos:0, YPos:0}
Chicken Location = Location{XPos:0, YPos:1}
CopperMine Location = Location{XPos:2, YPos:0}
IronMine Location = Location{XPos:1, YPos:7}
CoalMine Location = Location{XPos:1, YPos:6}
Forge Location = Location{XPos:1, YPos:5}
Kitchen Location = Location{XPos:1, YPos:1}
WeaponSmith Location = Location{XPos:2, YPos:1}
Bank Location = Location{XPos:4, YPos:1}
AshWood Location = Location{XPos:-1, YPos:0}
SpruceWood Location = Location{XPos:1, YPos:9}
BirchWood Location = Location{XPos:-1, YPos:6}
Sawmill Location = Location{XPos:-2, YPos:-3}
GreenSlime Location = Location{XPos:0, YPos:-1}
YellowSlime Location = Location{XPos:4, YPos:-1}
BlueSlime Location = Location{XPos:2, YPos:-1}
RedSlime Location = Location{XPos:1, YPos:-1}
Mushmush Location = Location{XPos:5, YPos:3}
MonsterTask Location = Location{XPos:1, YPos:2}
Sunflower Location = Location{XPos:2, YPos:2}
Alchemist Location = Location{XPos:2, YPos:3}
GudgeonPond Location = Location{XPos:4, YPos:2}
Wolf Location = Location{XPos:-2, YPos:1}
FlyingSerpent Location = Location{XPos:5, YPos:4}
Trader Location = Location{XPos:4, YPos:13}
)

func (l Location) String() string {
    return fmt.Sprintf("{\"x\":%d,\"y\":%d}", l.XPos, l.YPos)
}

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
    Data []Location `json:"data"`
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
    return artifactsRest("POST", p, BODY.String());
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
    artifactsMove(td, Location{XPos:td.XPos, YPos:td.YPos}); 

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

func GatherAndTradeThe(Item1 string, Place1 Location, ToonName ToonName, HowManyLoops int) {
    var numberToGather int;
    var numberGathered int;
    ThisToon := GetInfoFor(ToonName);
    if HowManyLoops < 0 {
        numberToGather = ThisToon.TaskTotal
        numberGathered = ThisToon.TaskProgress
    } else {
	numberToGather = 10000
	numberGathered = 0
    }
    td := GetInfoFor(ToonName);
    for c := numberGathered; c<numberToGather; c+=30 {
        artifactsMove(td,Place1);
        for m := AmountOf(Item1, ToonName); m<30; m++ {
            artifactsPost("my/"+string(ToonName)+"/action/gathering","");
        }
        artifactsMove(td,Trader);
	ThisToon = GetInfoFor(ToonName);
	RemainingTrade := ThisToon.TaskTotal - ThisToon.TaskProgress
	AmountToTrade := min(RemainingTrade,30)
        artifactsPost("my/"+string(ToonName)+"/action/task/trade","{\"code\":\""+Item1+"\",\"quantity\":"+strconv.Itoa(AmountToTrade)+"}");
    }
}

func GatherAndCraftAndTradeThe(Item1 string, Place1 Location, Item2 string, Place2 Location, ToonName ToonName, HowManyLoops int) {
    var l int;
    if HowManyLoops <= 0 {
        l = 10000
    } else {
        l = HowManyLoops
    }
    td := GetInfoFor(ToonName);
    for c := 0; c<l; c++ {
        artifactsMove(td,Place1);
        for m := AmountOf(Item1, ToonName); m<80; m++ {
            artifactsPost("my/"+string(ToonName)+"/action/gathering","");
        }
        artifactsMove(td,Place2);
        artifactsPost("my/"+string(ToonName)+"/action/crafting","{\"code\":\""+Item2+"\",\"quantity\":80}");
        artifactsMove(td,Trader);
	ThisToon := GetInfoFor(ToonName);
	RemainingTrade := ThisToon.TaskTotal - ThisToon.TaskProgress
	AmountToTrade := min(RemainingTrade,80)
        artifactsPost("my/"+string(ToonName)+"/action/task/trade","{\"code\":\""+Item2+"\",\"quantity\":"+strconv.Itoa(AmountToTrade)+"}");
            
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

func RunItemTasks(ToonName ToonName) {
    for c := 1; c>0; c++ {
        //Check For Task
        t := GetInfoFor(ToonName);
        if t.Task == "" {
            artifactsMove(t,Trader);
            // Get New Task
            artifactsPost("my/"+string(ToonName)+"/action/task/new","");
            t = GetInfoFor(ToonName);
        }
        
        if t.Task != "" || t.TaskProgress == t.TaskTotal {
            artifactsMove(t,Trader);
            artifactsPost("my/"+string(ToonName)+"/action/task/complete","");
            artifactsPost("my/"+string(ToonName)+"/action/task/new","");
            t = GetInfoFor(ToonName);
        }
        
	//TO-DO: Split on task _ as iron_ore !== iron_rocks
        var mapLoc MapLocation;
	lookup := strings.Split(t.Task,"_")[0]
        monsterLocation := artifactsGet("maps?content_code="+lookup+"&size=1");
        json.Unmarshal(monsterLocation, &mapLoc);
        logger.Debug("monsterLocation:"+string(monsterLocation));
        GatherAndTradeThe(t.Task, Location{XPos: mapLoc.Data[0].XPos,YPos:mapLoc.Data[0].YPos}, ToonName, -1); 
    }

}

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
            FightThe(Location{XPos: mapLoc.Data[0].XPos,YPos:mapLoc.Data[0].YPos}, ToonName, -1); 
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
    EnvErr := godotenv.Load()
    if EnvErr != nil {
       logger.Fatal("Error loading .env file")
    }
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
        GatherAndCraftAndTradeThe("ash_wood", AshWood, "ash_plank", Sawmill, t, -1);
        GatherAndCraftThe("birch_wood", BirchWood, "birch_plank", Sawmill, t, 100);
    }(Faraday);
    go func(t ToonName){
        EnsureOffCooldown(toons, t);
        //artifactsPost("my/"+string(t)+"/action/task/trade","{\"code\":\"gudgeon\",\"quantity\":67}");
        BankDeposit(t);
	RunItemTasks(t);
        GatherAndTradeThe("gudgeon", GudgeonPond, t, -1);
        GatherThe("sunflower", Sunflower, t);
        GatherAndCraftThe("copper_ore", CopperMine, "copper", Forge, t,50);
    }(Rainboom);
    go func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
	RunItemTasks(t);
	GatherAndTradeThe("iron_ore", IronMine,t,-1);
	GatherAndCraftThe("iron_ore", IronMine, "iron", Forge, t,-1);
    }(Crydelia);
    go func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
        FightThe(FlyingSerpent, t, -1);
        FightThe(Chicken, t, 800);
    }(Ikhor);
    func(t ToonName){
        EnsureOffCooldown(toons, t);
        BankDeposit(t);
        FightThe( Wolf , t , -1);
        GatherThe("coal", CoalMine, t);
    }(Troy)
}
