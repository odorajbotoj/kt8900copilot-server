package main

const (
	SKIP              = iota
	VERIFY                               // s -> c
	REFUSE                               // s -> c
	BUSY                                 // s -> c
	RX                = iota - 4 + 0x11  // c -> s
	RX_STOP                              // c -> s
	TX                                   // s -> c
	TX_STOP                              // s -> c
	IMG_UPLOAD                           // c -> s
	IMG_UPLOAD_STOP                      // c -> s
	IMG_DOWNLOAD                         // s -> c
	IMG_DOWNLOAD_STOP                    // s -> c
	IMG_GET                              // c -> s -> c
	SET_CONF                             // c -> s -> c
	RESET                                // c -> s -> c
	FROM                                 // s -> c
	ONLINE                               // s -> c
	OFFLINE                              // s -> c
	S_E_IMG_NIL       = iota - 13 + 0x31 // c -> s -> c
	S_S_SET_CONF                         // c -> s -> c
	S_E_SET_CONF                         // c -> s -> c
	PCM               = 0x51             // c -> s -> c
	IMG               = 0x61             // c -> s -> c
)

type dataPack struct {
	from string
	data []byte
}
