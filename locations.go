package main

import "fmt"

//TO-DO: Include Logic to generate locations dynamically from API

type Location struct{
    XPos int `json:"x"`
    YPos int `json:"y"`
}

var (
Spawn Location = Location{XPos:0, YPos:0}
// Battles
Chicken Location = Location{XPos:0, YPos:1}
Cow Location = Location{XPos:0, YPos:2}
GreenSlime Location = Location{XPos:0, YPos:-1}
YellowSlime Location = Location{XPos:4, YPos:-1}
BlueSlime Location = Location{XPos:2, YPos:-1}
RedSlime Location = Location{XPos:1, YPos:-1}
Mushmush Location = Location{XPos:5, YPos:3}
Wolf Location = Location{XPos:-2, YPos:1}
FlyingSerpent Location = Location{XPos:5, YPos:4}
// Mining
CopperMine Location = Location{XPos:2, YPos:0}
IronMine Location = Location{XPos:1, YPos:7}
CoalMine Location = Location{XPos:1, YPos:6}
GoldMine Location = Location{XPos:6, YPos:-3}
// Lumbering
AshWood Location = Location{XPos:-1, YPos:0}
SpruceWood Location = Location{XPos:1, YPos:9}
BirchWood Location = Location{XPos:-1, YPos:6}
// Crafting
Kitchen Location = Location{XPos:1, YPos:1}
WeaponSmith Location = Location{XPos:2, YPos:1}
Forge Location = Location{XPos:1, YPos:5}
Sawmill Location = Location{XPos:-2, YPos:-3}
Alchemist Location = Location{XPos:2, YPos:3}
// Tasks
MonsterTask Location = Location{XPos:1, YPos:2}
ItemTask Location = Location{XPos:4, YPos:13}
// Fishing
GudgeonPond Location = Location{XPos:4, YPos:2}
// Alchemy
Sunflower Location = Location{XPos:2, YPos:2};
Bank Location = Location{XPos:4, YPos:1}
)

func (l Location) String() string {
    return fmt.Sprintf("{\"x\":%d,\"y\":%d}", l.XPos, l.YPos)
}
