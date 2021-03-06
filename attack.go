package main

import (
	"DoomSlayeRL/routines"
	"fmt"
)

func m_meleeAttack(attacker *p_pawn, victim *p_pawn, d *dungeon) {
	if attacker.isPlayer() && victim.aiData != nil && victim.aiData.state == AI_STAGGERED {
		m_gloryKill(attacker, victim, d)
		return
	}
	damage := attacker.meleeData.damageDice.roll()
	victim.receiveDamage(damage)
	attacker.spendTurnsForAction(turnCostFor("melee_attack"))
	log.appendMessage(fmt.Sprintf("%s %s %s! (%d damage)", attacker.name, attacker.meleeData.meleeAttackString, victim.name, damage))
}

func (victim *p_pawn) receiveDamage(damage int) { //deals with armor, staggered state etc
	const (
		STAGGER_PERCENT_THRESHOLD = 30
		STAGGERED_TIME_AMOUNT     = 60
	)
	if victim.wearedArmor != nil {
		percent := victim.wearedArmor.armorData.damageConsumingPercent
		damageToArmor := damage * percent / 100
		victim.wearedArmor.armorData.currArmor -= damageToArmor
		damageToPawn := damage - damageToArmor
		if victim.wearedArmor.armorData.currArmor < 0 {
			damageToPawn -= victim.wearedArmor.armorData.currArmor
			victim.wearedArmor.armorData.currArmor = 0
		}
		victim.hp -= damageToPawn
	} else {
		victim.hp -= damage
	}
	if victim.isPlayer() == false {
		if victim.getHpPercent() < STAGGER_PERCENT_THRESHOLD {
			victim.aiData.state = AI_STAGGERED
			victim.aiData.stateTimeoutTurn = CURRENT_TURN + STAGGERED_TIME_AMOUNT
		}
	}
}

func m_rangedAttack(attacker *p_pawn, vx, vy int, dung *dungeon) {
	aw := attacker.weaponInHands
	ax, ay := attacker.getCoords()
	// vx, vy := victim.getCoords()
	if aw.weaponData.getType() == "projectile" {
		attacker.spendTurnsForAction(turnCostFor("ranged_attack"))
		proj := aw.weaponData.createProjectile(ax, ay, vx, vy)
		dung.addProjectileToList(proj)
		log.appendMessagef("%s shoots!", attacker.name)
		if aw.weaponData.maxammo > 0 {
			aw.weaponData.ammo -= 1 // TODO: investigate
		}
	}
	if aw.weaponData.getType() == "hitscan" {
		attacker.spendTurnsForAction(turnCostFor("ranged_attack"))
		shots := aw.weaponData.hitscanData.shotsPerAttack
		if shots < 1 {
			shots = 1
		}
		for i := 0; i < shots; i++ {
			m_traceHitscanShot(attacker, vx, vy, dung)
			checkDeadPawns(dung)
			if aw.weaponData.maxammo > 0 {
				aw.weaponData.ammo -= 1 // TODO: investigate
				if aw.weaponData.ammo == 0 {
					break
				}
			}
		}
	}
}

func m_traceHitscanShot(attacker *p_pawn, tox, toy int, d *dungeon) {
	const BULLET_TRACE_RANGE = 30
	aw := attacker.weaponInHands
	ax, ay := attacker.getCoords()
	pellets := aw.weaponData.hitscanData.pelletsPerShot
	if pellets < 1 {
		pellets = 1
	}
	spreadAngle := aw.weaponData.hitscanData.spreadAngle
	if spreadAngle == 0 {
		spreadAngle = 15 // TODO: think about that.
	}

	// init spread bound vectors
	leftSpreadVector := routines.CreateVectorByStartAndEndInt(ax, ay, tox, toy)
	leftSpreadVector.Rotate(spreadAngle / 2)
	rightSpreadVector := routines.CreateVectorByStartAndEndInt(ax, ay, tox, toy)
	rightSpreadVector.Rotate(-spreadAngle / 2)

	// init an array of vectors for trace
	bRealPositions := make([]*routines.Vector, 0)
	bPelletIsHit := make([]bool, pellets)
	bDirVectors := make([]*routines.Vector, 0)

	for i := 0; i < pellets; i++ {
		bRealPositions = append(bRealPositions, routines.CreateVectorByIntegers(ax, ay))
		bDirVectors = append(bDirVectors, routines.CreateRandomVectorBetweenTwo(leftSpreadVector, rightSpreadVector))
		bDirVectors[i].TransformIntoUnitVector()
	}

	// now lets trace each pellet
	totalHitPellets := 0
	for totalHitPellets < pellets {
		for i := 0; i < pellets; i++ {
			if !bPelletIsHit[i] {
				bRealPositions[i].Add(bDirVectors[i])
				bx, by := bRealPositions[i].GetRoundedCoords()
				if !areCoordinatesInRangeFrom(ax, ay, bx, by, BULLET_TRACE_RANGE) {
					bPelletIsHit[i] = true
					totalHitPellets++
					continue
				}
				victim := d.getPawnAt(bx, by)
				if victim != nil { // todo: bullets/pellets penetration
					// TODO: miss shots
					damage := aw.weaponData.hitscanData.damageDice.roll()
					victim.receiveDamage(damage)
					log.appendMessagef("The %s is hit!", victim.name)
					bPelletIsHit[i] = true
					totalHitPellets++
				}
				if d.isTileOpaque(bx, by) {
					bPelletIsHit[i] = true
					totalHitPellets++
				}
			}
		}
		renderBullets(bRealPositions, bDirVectors, d)
	}
}
