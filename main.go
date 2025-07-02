package main

import (
    "fmt"
    "os"
    "strings"
    "strconv"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "time"
    "slices"
    "github.com/joho/godotenv"
)
var logger = GetLogger();


type InventoryItem struct {
    Slot int `json:"slot"`
    Code string `json:"code"`
    Quantity int `json:"quantity"`
}
func (i InventoryItem) String() string {
    return fmt.Sprintf("{\"code\": \"%s\", \"quantity\": %d}", i.Code, i.Quantity)
}

type ItemDetails struct {
    Data struct{
        Craft struct{
            Items []InventoryItem `json:items`
        } `json:craft`
    } `json:data`
}

type BankInventory struct {
    Data []InventoryItem `json:data`
}

type Monster struct {
    Data struct{
        Level int `json:"level"`
    } `json:"data"`
}

type MapLocation struct {
    Data []Location `json:"data"`
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

    if res.StatusCode != 490 {
        if res.StatusCode != 200 {
            logger.Error(res.Status, ACTION, PATH, "data", string(responseData))
        } else {
            logger.Debug(res.Status, ACTION, PATH, "data", string(responseData));
        }
    	if strings.Contains(PATH, "action") {
        	var ta ToonAction;
        	json.Unmarshal(responseData, &ta);
        	logger.Debug("SLEEP","ToonName",ta.Data.Character.Name,"seconds",ta.Data.Cooldown.RemainingSeconds);
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
            artifactsPost("my/"+string(ToonName)+"/action/bank/deposit/item","[{\"code\":\""+td.Inventory[i].Code+"\",\"quantity\":"+strconv.Itoa(td.Inventory[i].Quantity)+"}]");
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
    logger.Info("FIGHT", "ToonName", ToonName, "numberFought", numberFought, "numberToFIght", numberToFight, "Location", MonsterLocation);
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

func GatherThe(Item1 string, Place1 Location, ToonName ToonName, PerformBankDeposit bool) {
    logger.Info("GATHER", "ToonName", ToonName, "Item", Item1, "Location", Place1);
    for c := 1; c>0; c++ {
        artifactsMove(GetInfoFor(ToonName),Place1);
        for m := AmountOf(Item1, ToonName); m<50; m++ {
            artifactsPost("my/"+string(ToonName)+"/action/gathering","");
        }
        if PerformBankDeposit {
            BankDeposit(ToonName);
        }
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
    logger.Info("GATHER/TRADE", "ToonName", ToonName, "Item", Item1, "Location", Place1, "ForTask", HowManyLoops);
    for c := numberGathered; c<numberToGather; c+=30 {
        artifactsMove(td,Place1);
        for m := AmountOf(Item1, ToonName); m<30; m++ {
            artifactsPost("my/"+string(ToonName)+"/action/gathering","");
        }
        artifactsMove(td,ItemTask);
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
        artifactsMove(td,ItemTask);
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
    logger.Info("GATHER/CRAFT", "ToonName", ToonName, "Item", Item1, "Item2", Item2);
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

func FromBankCraftThe(Item1 string, Item2 string, Place2 Location, ToonName ToonName) {
    logger.Info("BANK/CRAFT", "ToonName", ToonName, "Item", Item1, "Item2", Item2);
    td := GetInfoFor(ToonName);
    // Check Items in Bank
    var bank BankInventory;
    rawbank := artifactsGet("my/bank/items?item_code="+Item2);
    json.Unmarshal(rawbank, &bank);
    // Get recipe
    var item ItemDetails;
    rawitem := artifactsGet("items/"+Item2);
    json.Unmarshal(rawitem, &item);
    // loop math
    var numberToCraft = item.Data.Craft.Items[0].Quantity;
    var numberInBank = bank.Data[0].Quantity;
    var Loops = numberInBank / numberToCraft;


    for l:=0; l<Loops; l++ {
    // bank deposit
        BankDeposit(ToonName);    
        artifactsMove(td,Bank);
    // grab items
        artifactsPost("my/"+string(ToonName)+"/action/bank/withdraw/item","[{\"code\":\""+Item2+"\",\"quantity\":"+strconv.Itoa(100)+"}]");
    // craft bank deposit
        artifactsMove(td,Place2);
        artifactsPost("my/"+string(ToonName)+"/action/crafting","{\"code\":\""+Item2+"\",\"quantity\":"+strconv.Itoa(100/numberToCraft)+"}");
        BankDeposit(ToonName);
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
            artifactsMove(t,ItemTask);
            // Get New Task
            artifactsPost("my/"+string(ToonName)+"/action/task/new","");
            t = GetInfoFor(ToonName);
        }
        
        if t.Task != "" || t.TaskProgress == t.TaskTotal {
            artifactsMove(t,ItemTask);
            artifactsPost("my/"+string(ToonName)+"/action/task/complete","");
            artifactsPost("my/"+string(ToonName)+"/action/task/new","");
            t = GetInfoFor(ToonName);
        }
        
        var mapLoc MapLocation;
        monsterLocation := artifactsGet("maps?content_code="+t.Task+"&size=1");
        json.Unmarshal(monsterLocation, &mapLoc);
        if len(mapLoc.Data) == 0 {
            logger.Warn("No Map Data Found For Task: "+t.Task);
            lookup := strings.Split(t.Task,"_")[0]
            monsterLocation := artifactsGet("maps?content_code="+lookup+"&size=1");
            json.Unmarshal(monsterLocation, &mapLoc);
            if len(mapLoc.Data) == 0 {
                logger.Error("No Map Data Found For Task After Split: "+lookup);
                break;
            }
        }
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
            
        //TO-DO: Replace with a Re-Equip For Monster Weakness
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

// Check for DateTime Object `cooldown_expiration`
// `cooldown` appears to just be the duration of the last cooldown
// Even if it was in the past and is over
func EnsureOffCooldown(Toons Toon, ToonName ToonName) {
    idx := slices.IndexFunc(Toons.Data, func(td ToonDetails) bool { return td.Name == ToonName })
    InitialSleepBuffer := 0;
    t, _ := time.Parse(time.RFC3339, Toons.Data[idx].CooldownExpiration)
    if t.After(time.Now()) {
        InitialSleepBuffer = int(t.Sub(time.Now()).Seconds())+1
    }
    if InitialSleepBuffer > 0 {
        logger.Info(ToonName+" Needs to Wait","seconds",InitialSleepBuffer);
        time.Sleep(time.Duration(InitialSleepBuffer)*time.Second);
    } else {
        logger.Info(ToonName+" is Ready-To-Go!");
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
        idx := slices.IndexFunc(toons.Data, func(td ToonDetails) bool { return td.Name == t });
        td := toons.Data[idx];
        td.EnsureOffCooldown();
        td.BankDeposit();
        td.GatherAndCraftThe("spruce_wood", SpruceWood, "spruce_plank", Sawmill, -1);
        td.FightThe(Chicken, 100);
    }(Faraday);
    go func(t ToonName){
        idx := slices.IndexFunc(toons.Data, func(td ToonDetails) bool { return td.Name == t });
        td := toons.Data[idx];
        td.EnsureOffCooldown();
        td.BankDeposit();
        td.GatherThe("gudgeon", GudgeonPond,true);
        td.GatherAndCraftThe("sunflower", Sunflower, "small_health_potion", Alchemist,-1);
    }(Rainboom);
    go func(t ToonName){
        idx := slices.IndexFunc(toons.Data, func(td ToonDetails) bool { return td.Name == t });
        td := toons.Data[idx];
        td.EnsureOffCooldown();
        td.BankDeposit();
        td.GatherAndCraftThe("iron_ore", IronMine, "iron_bar", Forge,-1);
    }(Crydelia);
    go func(t ToonName){
        idx := slices.IndexFunc(toons.Data, func(td ToonDetails) bool { return td.Name == t });
        td := toons.Data[idx];
        td.EnsureOffCooldown();
        td.BankDeposit();
        td.GatherAndCraftThe("iron_ore", IronMine, "iron_bar", Forge,-1);
        td.GatherAndCraftThe("copper_ore", CopperMine, "copper_bar", Forge,-1);
        td.FightThe(Chicken, 800);
    }(Ikhor);
    func(t ToonName){
        idx := slices.IndexFunc(toons.Data, func(td ToonDetails) bool { return td.Name == t });
        td := toons.Data[idx];
        td.EnsureOffCooldown();
        td.BankDeposit();
        td.FightThe(Chicken, -1);
        td.GatherAndCraftThe("spruce_wood", SpruceWood, "spruce_plank", Sawmill, -1);
    }(Troy)
}
