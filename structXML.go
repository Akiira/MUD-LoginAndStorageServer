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

	WeaponComment xml.Comment  `xml:",comment"`
	EquipedWeapon WeaponXML    `xml:"Weapon"`
	ArmSet        ArmourSetXML `xml:"ArmourSet"`
	PersInv       InventoryXML `xml:"Inventory"`
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
