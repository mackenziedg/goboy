package main

type GPU struct {
	mmu *MMU
}

func (g *GPU) Reset(mmu *MMU) {
}

func (g *GPU) LoadTile(address uint16) {

}
