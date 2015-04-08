package main

import (
	"encoding/xml"
)

type CharacterXML struct {
	XMLName xml.Name `xml:"Character"`
	Name    string   `xml:"Name"`
	RoomIN  int      `xml:"RoomIN"`
	HP      int      `xml:"HitPoints"`
	Defense int      `xml:"Defense"`

	Strength     int `xml:"Strength"`
	Constitution int `xml:"Constitution"`
	Dexterity    int `xml:"Dexterity"`
	Wisdom       int `xml:"Wisdom"`
	Charisma     int `xml:"Charisma"`
	Inteligence  int `xml:"Inteligence"`

	Level      int `xml:"Level"`
	experience int `xml:"experience"`

	CurrentWorld string `xml:"CurrentWorld"`

	EquipedWeapon WeaponXML    `xml:"Weapon"`
	ArmSet        ArmourSetXML `xml:"ArmourSet"`
	PersInv       InventoryXML `xml:"Inventory"`
}

type InventoryXML struct {
	XMLName xml.Name    `xml:"Inventory"`
	Items   []ItemXML   `xml:"Item"`
	Weapons []WeaponXML `xml:"Weapon"`
	Armours []ArmourXML `xml:"Armour"`
}

type ArmourSetXML struct {
	XMLName xml.Name    `xml:"ArmourSet"`
	ArmSet  []ArmourXML `xml:"Armour"`
}

type ArmourXML struct {
	XMLName      xml.Name `xml:"Armour"`
	ItemInfo     ItemXML  `xml:"Item"`
	Defense      int      `xml:"Defense"`
	WearLocation string   `xml:"Location"`
}

type WeaponXML struct {
	XMLName  xml.Name `xml:"Weapon"`
	ItemInfo ItemXML  `xml:"Item"`
	Attack   int      `xml:"Attack"`
	Damage   int      `xml:"Damage"`
}

type ItemXML struct {
	XMLName     xml.Name `xml:"Item"`
	Name        string   `xml:"Name"`
	Description string   `xml:"Description"`
	ItemLevel   int      `xml:"Level"`
	ItemWorth   int      `xml:"Worth"`
	Quantity    int      `xml:"Quantity"`
}
