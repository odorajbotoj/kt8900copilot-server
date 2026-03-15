package main

const (
	SKIP              = 0x00
	VERIFY            = 0x01 // s -> c
	REFUSE            = 0x02 // s -> c
	BUSY              = 0x03 // s -> c
	VERIFIED          = 0x04 // s -> c
	RX                = 0x11 // c -> s
	RX_STOP           = 0x12 // c -> s
	TX                = 0x13 // s -> c
	TX_STOP           = 0x14 // s -> c
	IMG_UPLOAD        = 0x15 // c -> s
	IMG_UPLOAD_STOP   = 0x16 // c -> s
	IMG_DOWNLOAD      = 0x17 // s -> c
	IMG_DOWNLOAD_STOP = 0x18 // s -> c
	IMG_GET           = 0x19 // c -> s -> c
	SET_CONF          = 0x1A // c -> s -> c
	RESET             = 0x1B // c -> s -> c
	PLAY              = 0x1C // c -> s -> c
	AFSK              = 0x1D // c -> s -> c
	S_MESSAGE         = 0x31 // c -> s -> c
	S_S_SET_CONF      = 0x32 // c -> s -> c
	S_E_SET_CONF      = 0x33 // c -> s -> c
	S_E_CAM_DISABLED  = 0x34 // c -> s -> c
	S_E_IMG_NIL       = 0x35 // c -> s -> c
	S_S_PLAY          = 0x36 // c -> s -> c
	S_E_PLAY          = 0x37 // c -> s -> c
	FROM              = 0x51 // s -> c
	ONLINE            = 0x52 // s -> c
	OFFLINE           = 0x53 // s -> c
	PCM               = 0x71 // c -> s -> c
	IMG               = 0x72 // c -> s -> c
)

type dataPack struct {
	from string
	data []byte
}
