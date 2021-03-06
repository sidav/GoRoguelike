package main

import (
	"fmt"
)

func plr_playerControl(d *dungeon) {
	p := d.player
	valid_key_pressed := false
	movex := 0
	movey := 0
	for !valid_key_pressed {
		key_pressed := readKey()
		valid_key_pressed = true
		movex, movey = plr_keyToDirection(key_pressed)
		if movex == 0 && movey == 0 {
			switch key_pressed {
			case "5", ".":
				d.player.spendTurnsForAction(10) // just wait for a sec
			case "g":
				plr_doPickUpButton(d)
			case "f":
				plr_aimAndFire(d)
			case "D":
				plr_doDropButton(d)
			case "i":
				if len(p.inventory.items) > 0 {
					plr_UseItemFromInventory(p)
				} else {
					log.appendMessage("You have no items.")
				}
			case "r":
				plr_reload(p)
			case "ESCAPE":
				GAME_IS_RUNNING = false
			case "[": // debug
				RENDER_DISABLE_LOS = !RENDER_DISABLE_LOS
				log.appendMessage("Changed LOS setting.")
			case "M": // debug
				d.init_placeItemsAndEnemies()
				log.appendMessage("Spawned MOAR!!!")
			default:
				valid_key_pressed = false
				log.appendMessagef("Unknown key %s (Wrong keyboard layout?)", key_pressed)
				renderLevel(d, true)
			}
		}
	}
	// log.appendMessage(key_pressed)

	if movex != 0 || movey != 0 {
		m_moveOrMeleeAttackPawn(d.player, d, movex, movey)
	}
	plr_pickupInstantlyPickupables(d)
	plr_checkItemsOnFloor(d)
}

func plr_keyToDirection(keyPressed string) (int, int) {
	switch keyPressed {
	case "s", "2":
		return 0, 1
	case "w", "8":
		return 0, -1
	case "a", "4":
		return -1, 0
	case "d", "6":
		return 1, 0
	case "7":
		return -1, -1
	case "9":
		return 1, -1
	case "1":
		return -1, 1
	case "3":
		return 1, 1
	default:
		return 0, 0
	}
}

func plr_aimAndFire(d *dungeon) {
	p := d.player
	if p.weaponInHands == nil {
		log.appendMessage("You have nothing to fire with!")
		return
	}
	if !p.weaponInHands.weaponData.hasEnoughAmmoToShoot() {
		log.appendMessage("You are out of your ammo! Reload!")
		return
	}
	targets := d.getListOfPawnsVisibleFor(p)
	curr_target_index := 0
	aimx, aimy := p.x, p.y
	// choose target
	if len(targets) > 0 {
		aimx, aimy = targets[curr_target_index].x, targets[curr_target_index].y
	}
		log.appendMessagef("You target with your %s.", p.weaponInHands.name)
	aimLoop:
		for {
			renderTargetingLine(p.x, p.y, aimx, aimy, true, d)
			keypressed := readKey()
			aimModx, aimMody := plr_keyToDirection(keypressed)
			switch keypressed {
			case "n", "TAB":
				curr_target_index++
				if curr_target_index >= len(targets) {
					curr_target_index = 0
				}
				if len(targets) > 0 {
					aimx, aimy = targets[curr_target_index].x, targets[curr_target_index].y
				}
			case "f":
				if aimx == p.x && aimy == p.y {
					log.appendMessage("Why would you want to do that?")
					return
				}
				m_rangedAttack(p, aimx, aimy, d)
				break aimLoop
			case "ESCAPE":
				log.appendMessage("Okay, then.")
				break aimLoop
			}
			aimx += aimModx
			aimy += aimMody
		}
	}

func plr_doDropButton(d *dungeon) {
	p := d.player
	items := p.inventory.items
	if len(items) == 0 {
		log.appendMessage("You have nothing to drop.")
		return
	} else {
		item := p.inventory.selectItem()
		if item != nil {
			item.x, item.y = p.x, p.y
			p.inventory.removeItem(item)
			d.addItemToFloor(item)
			log.appendMessagef("You drop your %s.", item.name)
		} else {
			log.appendMessage("Okely-dokely.")
		}
	}
}

func plr_doPickUpButton(d *dungeon) {
	p := d.player
	items := d.getListOfItemsAt(p.x, p.y)
	for i := 0; i < len(items); i++ {
		item := items[i]
		plr_pickUpAnItem(item, d)
		return
	}
	if len(items) == 0 {
		log.appendMessage("There is nothing here.")
		return
	}
}

func plr_pickupInstantlyPickupables(d *dungeon) {
	px, py := d.player.getCoords()
	items := d.getListOfItemsAt(px, py)
	for i:=0; i<len(items);i++{
		if items[i].instantlyPickupable {
			plr_pickUpAnItem(items[i], d)
		}
	}
}

func plr_checkItemsOnFloor(d *dungeon) {
	px, py := d.player.getCoords()
	items := d.getListOfItemsAt(px, py)
	if len(items) == 1 {
		log.appendMessage(fmt.Sprintf("You see here a %s", items[0].name))
	} else if len(items) > 1 {
		log.appendMessage(fmt.Sprintf("You see here a %s and %d more items", items[0].name, len(items)-1))
	}
}

func plr_reload(p *p_pawn) {
	if !p.canShoot() {
		log.appendMessage("You have nothing to reload.")
		return
	}
	wpn := p.weaponInHands.weaponData
	ammoType := wpn.ammoType
	currInvAmmo := p.inventory.ammo[ammoType]
	currAmmo := p.weaponInHands.weaponData.ammo
	maxAmmo := p.weaponInHands.weaponData.maxammo
	ammoToRefill := maxAmmo - currAmmo
	if ammoToRefill == 0 {
		log.appendMessagef("Your %s is already loaded!", p.weaponInHands.name)
		return
	}
	if currInvAmmo == 0 {
		log.appendMessagef("You have no ammo to reload your %s.", p.weaponInHands.name)
		return
	}
	if currInvAmmo >= ammoToRefill {
		p.weaponInHands.weaponData.ammo = maxAmmo
		p.inventory.ammo[ammoType] -= ammoToRefill
	} else {
		p.weaponInHands.weaponData.ammo += currInvAmmo
		p.inventory.ammo[ammoType] = 0
	}
	p.spendTurnsForAction(turnCostFor("reload"))
	log.appendMessagef("You reload your %s.", p.weaponInHands.name)
}
