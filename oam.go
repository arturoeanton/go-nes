package main

// Sprite representa los datos de un objeto de 8x8 o 8x16
type Sprite struct {
	Y      byte
	TileID byte
	Attr   byte
	X      byte
}

// OAM gestiona los 256 bytes de memoria de objetos (64 sprites * 4 bytes)
type OAM struct {
	Data [256]byte
	Addr byte // Registro $2003
}

func (o *OAM) WriteAddr(addr byte) {
	o.Addr = addr
}

func (o *OAM) WriteData(data byte) {
	o.Data[o.Addr] = data
	o.Addr++
}

func (o *OAM) ReadData() byte {
	return o.Data[o.Addr]
}

// GetSprite devuelve un sprite estructurado para facilitar el renderizado
func (o *OAM) GetSprite(index int) Sprite {
	offset := index * 4
	return Sprite{
		Y:      o.Data[offset],
		TileID: o.Data[offset+1],
		Attr:   o.Data[offset+2],
		X:      o.Data[offset+3],
	}
}
