package main

import (
	"GoRoguelike/routines"
	cw "TCellConsoleWrapper/tcell_wrapper"
)

type dungeon struct {
	player      *p_pawn
	tiles       [levelsizex][levelsizey]d_tile
	pawns       []*p_pawn
	items       []*i_item
	projectiles []*projectile
}

func (dung *dungeon) spawnPawnAtRandomPosition(name string) {
	for tries := 0; tries < 1000; tries++ {
		x, y := routines.Random(levelsizex), routines.Random(levelsizey)
		if dung.isTilePassableAndNotOccupied(x, y) {
			dung.pawns = append(dung.pawns, p_createPawn(name, x, y))
			return
		}
	}
}

func (dung *dungeon) spawnItemAtRandomPosition(name string) {
	for tries := 0; tries < 1000; tries++ {
		x, y := routines.Random(levelsizex), routines.Random(levelsizey)
		if dung.isTilePassableAndNotOccupied(x, y) {
			dung.items = append(dung.items, i_createItem(name, x, y))
			return
		}
	}
}

func (dung *dungeon) isPawnPresent(ix, iy int) bool {
	x, y := dung.player.x, dung.player.y
	if ix == x && iy == y {
		return true
	}
	for i := 0; i < len(dung.pawns); i++ {
		x, y = dung.pawns[i].x, dung.pawns[i].y
		if ix == x && iy == y {
			return true
		}
	}
	return false
}

func (dung *dungeon) getPawnAt(x, y int) *p_pawn {
	px, py := dung.player.x, dung.player.y
	if px == x && py == y {
		return dung.player
	}
	for i := 0; i < len(dung.pawns); i++ {
		px, py = dung.pawns[i].x, dung.pawns[i].y
		if px == x && py == y {
			return dung.pawns[i]
		}
	}
	return nil
}

func (d *dungeon) removePawn(p *p_pawn) {
	for i := 0; i < len(d.pawns); i++ {
		if p == d.pawns[i] {
			d.pawns = append(d.pawns[:i], d.pawns[i+1:]...) // ow it's fucking... magic!
		}
	}
}

func (d *dungeon) isItemPresent(ix, iy int) bool {
	for i := 0; i < len(d.items); i++ {
		x, y := d.items[i].x, d.items[i].y
		if ix == x && iy == y {
			return true
		}
	}
	return false
}

func (d *dungeon) addItemToFloor(item *i_item) {
	d.items = append(d.items, item)
}

func (d *dungeon) getListOfItemsAt(ix, iy int) []*i_item {
	items := make([]*i_item, 0)
	for i := 0; i < len(d.items); i++ {
		x, y := d.items[i].x, d.items[i].y
		if ix == x && iy == y {
			items = append(items, d.items[i])
		}
	}
	return items
}

func (d *dungeon) removeItemFromFloor(item *i_item) {
	for i := 0; i < len(d.items); i++ {
		if item == d.items[i] {
			d.items = append(d.items[:i], d.items[i+1:]...) // ow it's fucking... magic!
		}
	}
}

func (dung *dungeon) isTilePassable(x, y int) bool {
	if x < 0 || x >= levelsizex || y < 0 || y >= levelsizey {
		log.warningf("Passability for unexistent index %d requested!", x)
		return false
	}
	return dung.tiles[x][y].IsPassable
}

func (dung *dungeon) isTileOpaque(x, y int) bool {
	if x < 0 || x >= levelsizex || y < 0 || y >= levelsizey {
		log.warningf("Opacity for unexistent index %d requested!", x)
		return true
	}
	return dung.tiles[x][y].opaque
}

func (dung *dungeon) isTilePassableAndNotOccupied(x, y int) bool {
	return dung.isTilePassable(x, y) && !dung.isPawnPresent(x, y)
}

func (d *dungeon) addProjectileToList(p *projectile) {
	d.projectiles = append(d.projectiles, p)
}

func (d *dungeon) removeProjectileFromList(p *projectile) {
	for i := 0; i < len(d.projectiles); i++ {
		if p == d.projectiles[i] {
			d.projectiles = append(d.projectiles[:i], d.projectiles[i+1:]...) // ow it's fucking... magic!
		}
	}
}

func (d *dungeon) addBloodSplats(x, y, radius int) {
	const (
		SPLAT_CHANCE   = 40
		GIB_CHANCE     = 40
		BIG_GIB_CHANCE = 30
	)
	for i := x - radius; i <= x+radius; i++ {
		for j := y - radius; j <= y+radius; j++ {
			if areCoordinatesValid(i, j) {
				if routines.RandomPercent() < SPLAT_CHANCE || (x == i && y == j) {
					d.tiles[i][j].cCell.color = cw.RED
					currApp := d.tiles[i][j].cCell.appearance
					if (currApp == '.' || currApp == ',') && routines.RandomPercent() < GIB_CHANCE {
						if routines.RandomPercent() < BIG_GIB_CHANCE {
							d.tiles[i][j].cCell.appearance = ';'
						} else {
							d.tiles[i][j].cCell.appearance = ','
						}
					}
				}
			}
		}
	}
}
