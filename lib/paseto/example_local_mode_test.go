package paseto

import (
	"encoding/hex"
	"fmt"
	"log"
)

func ExampleLocalMode() {
	//
	// In local mode, we create sender and receiver using the same key.
	//
	key, _ := hex.DecodeString("707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f")
	sender, _ := NewLocalMode(key)
	receiver, _ := NewLocalMode(key)

	token, err := sender.Pack([]byte("Hello receiver"), []byte(">>> footer"))
	if err != nil {
		log.Fatal(err)
	}

	// Sender then send the encrypted token to receiver
	// ...

	// Receiver unpack the token from sender to get the plain text and
	// footer.
	plain, footer, err := receiver.Unpack(token)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Receive data from sender: %s\n", plain)
	fmt.Printf("Receive footer from sender: %s\n", footer)
	// Output:
	// Receive data from sender: Hello receiver
	// Receive footer from sender: >>> footer
}
