package main

import (
    "time"
    "fmt"
)

type ToonName string

const (
    Troy ToonName = "Troy"
    Faraday ToonName = "Faraday"
    Rainboom ToonName = "Rainboom"
    Ikhor ToonName = "Ikhor"
    Crydelia ToonName = "Crydelia"
)


type ToonDetails struct {
        Name    ToonName `json:"name"`
        Account string `json:"account"`
        Level   int `json:"level"`
        XPos    int `json:"x"`
        YPos    int `json:"y"`
        Cooldown int `json:"cooldown"`
        CooldownExpiration string `json:"cooldown_expiration"`
        Task    string `json:"task"`
        TaskTotal int `json:"task_total"`
        TaskProgress int `json:"task_progress"`
        WoodcuttingLevel int `json:"woodcutting_level"`
        MiningLevel int `json:"mining_level"`
        FishingLevel int `json:"fishing_level"`
        AlchemyLevel int `json:"alchemy_level"`
        Inventory []InventoryItem `json:"inventory"`
}

type ToonAction struct {
    Data struct{
        Cooldown struct{
            RemainingSeconds float32 `json:"remaining_seconds"`
        } `json:"cooldown"`
        Character ToonDetails `json:"character"`
    } `json:"data"`
}

type Toon struct {
    Data []ToonDetails `json:"data"`
}


func (td ToonDetails) moveTo(BODY Location) []byte {
    // TO-DO: Don't Try To Move If Already There
    p := "my/"+string(td.Name)+"/action/move";
    return artifactsRest("POST", p, BODY.String());
};

func (td ToonDetails) BankDeposit() {
    td.moveTo(Bank);
    t := GetInfoFor(td.Name);
    if len(t.Inventory) > 0 {
        inv := "";
        for _, i := range t.Inventory {
            if i.Quantity > 0 {
                inv += i.String();
                inv += ","
            }
        }
        if inv == "" {
            inv = "[]";
        } else {
            inv = "[" + inv[:len(inv)-1] + "]";
        }
        fmt.Println(inv)
        artifactsPost("my/"+string(td.Name)+"/action/bank/deposit/item",inv);
    }
    t.moveTo(Location{XPos:td.XPos, YPos:td.YPos});
}

// Check for DateTime Object `cooldown_expiration`
// `cooldown` appears to just be the duration of the last cooldown
// Even if it was in the past and is over
func (td ToonDetails) EnsureOffCooldown() {
    InitialSleepBuffer := 0;
    t, _ := time.Parse(time.RFC3339, td.CooldownExpiration)
    if t.After(time.Now()) {
        InitialSleepBuffer = int(t.Sub(time.Now()).Seconds())+1
    }
    if InitialSleepBuffer > 0 {
        logger.Info(td.Name+" Needs to Wait","seconds",InitialSleepBuffer);
        time.Sleep(time.Duration(InitialSleepBuffer)*time.Second);
    } else {
        logger.Info(td.Name+" is Ready-To-Go!");
    }
}

func (td ToonDetails) FightThe(MonsterLocation Location, HowMany int){
    numberToFight := HowMany;
    numberFought := 0;
    if HowMany < 0 {
        numberToFight = td.TaskTotal
        numberFought = td.TaskProgress
    }
    logger.Info("FIGHT", "ToonName", td.Name, "numberFought", numberFought, "numberToFIght", numberToFight, "Location", MonsterLocation);
    artifactsPost("my/"+string(td.Name)+"/action/rest","");
    td.moveTo(MonsterLocation);
    for c := numberFought; c<=numberToFight; c++ {
        artifactsPost("my/"+string(td.Name)+"/action/fight","");
        artifactsPost("my/"+string(td.Name)+"/action/rest","");
        // TO-DO: This Inventory check should be handled by the Rest Handler
        if c % 25 == 24 {
            td.BankDeposit();
        }
    }
}


func (td ToonDetails) GatherAndCraftThe(Item1 string, Place1 Location, Item2 string, Place2 Location, HowManyLoops int) {
    var l int;
    if HowManyLoops <= 0 {
        l = 10000
    } else {
        l = HowManyLoops
    }
    logger.Info("GATHER/CRAFT", "ToonName", td.Name, "Item", Item1, "Item2", Item2);
    for c := 0; c<l; c++ {
        td.moveTo(Place1);
        for m := AmountOf(Item1, td.Name); m<30; m++ {
            artifactsPost("my/"+string(td.Name)+"/action/gathering","");
        }
        td.moveTo(Place2);
        artifactsPost("my/"+string(td.Name)+"/action/crafting","{\"code\":\""+Item2+"\",\"quantity\":3}");
        if AmountOf(Item2, td.Name) >= 20 {
            td.BankDeposit();
        }
            
    }
}

func (td ToonDetails) GatherThe(Item1 string, Place1 Location, PerformBankDeposit bool) {
    logger.Info("GATHER", "ToonName", td.Name, "Item", Item1, "Location", Place1);
    for c := 1; c>0; c++ {
        td.moveTo(Place1);
        for m := AmountOf(Item1, td.Name); m<50; m++ {
            artifactsPost("my/"+string(td.Name)+"/action/gathering","");
        }
        if PerformBankDeposit {
            td.BankDeposit();
        }
    }
}
