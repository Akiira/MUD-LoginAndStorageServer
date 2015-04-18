package main

import (
	"encoding/xml"
	"io"
)

type Item_I interface {
}

type ItemXML_I interface {
}

type CharacterXML struct {
	XMLName xml.Name `xml:"Character"`
	Name    string   `xml:"Name"`
	RoomIN  int      `xml:"RoomIN"`
	HP      int      `xml:"HitPoints"`
	Race    string   `xml:"Race"`
	Class   string   `xml:"Class"`

	Strength     int `xml:"Strength"`
	Constitution int `xml:"Constitution"`
	Dexterity    int `xml:"Dexterity"`
	Wisdom       int `xml:"Wisdom"`
	Charisma     int `xml:"Charisma"`
	Inteligence  int `xml:"Inteligence"`

	Level      int `xml:"Level"`
	Experience int `xml:"Experience"`
	Gold       int `xml:"Gold"`

	CurrentWorld string `xml:"CurrentWorld"`

	WeaponComment xml.Comment  `xml:",comment"`
	EquipedWeapon WeaponXML    `xml:"Weapon"`
	ArmSet        ArmourSetXML `xml:"ArmourSet"`
	PersInv       InventoryXML `xml:"Inventory"`
}

func (c *CharacterXML) SetStats(stats []int) {
	c.Strength = stats[0]
	c.Constitution = stats[1]
	c.Dexterity = stats[2]
	c.Wisdom = stats[3]
	c.Charisma = stats[4]
	c.Inteligence = stats[5]
}

func (c *CharacterXML) SetToDefaultValues() {
	c.RoomIN = 1001
	c.HP = 20 + c.Constitution
	c.Level = 1
	c.Gold = 50
	c.CurrentWorld = "world1"
	c.WeaponComment = xml.Comment("The equiped weapon")
	c.EquipedWeapon = NewButterKnife()
}

type ItemXML struct {
	XMLName     xml.Name `xml:"Item"`
	Name        string   `xml:"Name"`
	Description string   `xml:"Description"`
	ItemLevel   int      `xml:"Level"`
	ItemWorth   int      `xml:"Worth"`
}

type InventoryXML struct {
	XMLName xml.Name      `xml:"Inventory"`
	Items   []interface{} `xml:",any"`
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
	ItemInfo *ItemXML `xml:"Item"`
	Attack   int      `xml:"Attack"`
	MinDmg   int      `xml:"MinDmg"`
	MaxDmg   int      `xml:"MaxDmg"`
}

func NewButterKnife() WeaponXML {
	var wpn WeaponXML
	wpn.ItemInfo = new(ItemXML)
	wpn.ItemInfo.Name = "Rusty Butter Knife"
	wpn.ItemInfo.Description = "An old rusty butter knife, why would anyone use this as a weapon?"
	wpn.ItemInfo.ItemLevel = 0
	wpn.ItemInfo.ItemWorth = 0
	wpn.Attack = 1
	wpn.MinDmg = 0
	wpn.MaxDmg = 2

	return wpn
}

func (c *InventoryXML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var item ItemXML_I

	for t, err := d.Token(); err != io.EOF; {
		switch t1 := t.(type) {
		case xml.StartElement:
			if t1.Name.Local == "Armour" {
				item = new(ArmourXML)
			} else if t1.Name.Local == "Weapon" {
				item = new(WeaponXML)
			} else {
				item = new(ItemXML)
			}

			err = d.DecodeElement(item, &t1)
			checkError(err)
			c.Items = append(c.Items, item)
		}

		t, err = d.Token()
	}

	return nil
}
