package main

const (
	SKIP = iota
	VERIFY
	REFUSE
	RX = iota - 3 + 0x11
	RX_STOP
	TX
	TX_STOP
	IMG_UPLOAD
	IMG_UPLOAD_STOP
	IMG_DOWNLOAD
	IMG_DOWNLOAD_STOP
	IMG_GET
	SET_CONF
	RESET
	FROM
	S_E_IMG_NIL = iota - 12 + 0x31
	S_S_SET_CONF
	S_E_SET_CONF
	PCM = 0x51
	IMG = 0x61
)

type dataPack struct {
	from string
	data []byte
}
