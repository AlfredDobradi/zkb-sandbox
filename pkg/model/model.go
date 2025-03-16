package model

import "time"

type Killmail struct {
	KillmailID        uint64          `json:"killmail_id"`
	Attackers         []CharacterInfo `json:"attackers"`
	Victim            CharacterInfo   `json:"victim"`
	OriginalTimestamp time.Time       `json:"killmail_time"`
	Zkill             struct {
		URL  string `json:"url"`
		Hash string `json:"hash"`
		NPC  bool   `json:"npc"`
	} `json:"zkb"`
}

type CharacterInfo struct {
	CharacterID   uint64 `json:"character_id"`
	CorporationID uint64 `json:"corporation_id"`
	AllianceID    uint64 `json:"alliance_id"`
}
