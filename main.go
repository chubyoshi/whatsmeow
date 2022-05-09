package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var client *whatsmeow.Client

func init() {
	fmt.Println("---Initiating Program---")
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:wapp.db?_foreign_keys=on", dbLog)
	if err != nil {
		fmt.Printf("[Main]DB Err: %+v\n", err)
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		fmt.Printf("[Main][GetFirstDevice] Err: %+v\n", err)
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler) //Event handler such as receive message

	if client.Store.ID == nil {
		fmt.Println("---NIL---")
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			fmt.Printf("[Main][GetQRChannel] Err: %+v\n", err)
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		fmt.Println("---CONNECT---")
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			fmt.Printf("[Main][Connect] Err: %+v\n", err)
			panic(err)
		}
	}
	fmt.Println("---EXIT LOGIN---")
}

func main() {
	defer client.Disconnect()
	go scheduler()

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	go func() {
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	}()
	<-c
	fmt.Println("---EXIT PROGRAM---")
}

//eventHandler Used to automated events such as event message
func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("---Received a message!", v.Message.GetConversation())
		// if v.Info.Sender.User == "6281386133023" {
		// 	t, err := client.SendMessage(types.JID{
		// 		User:   "6281386133023", //Agus
		// 		Server: types.DefaultUserServer,
		// 	}, "", &waProto.Message{
		// 		Conversation: proto.String("gpp ey!"),
		// 	})
		// 	if err != nil {
		// 		fmt.Printf("---[Main][SendMessage] Time: %s Err: %+v---\n", t.Format("RFC850"), err)
		// 	}
		// }
	}
}

func scheduler() {
	localTime, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		localTime = time.FixedZone("UTC+7", 7)
	}
	scheduler := gocron.NewScheduler(localTime)
	_, err = scheduler.Cron("0 0 * * 1-6").Do(BookPool) //monday - saturday at 00:00
	if err != nil {
		fmt.Printf("---[Main][NewScheduler] Err: %+v---\n", err)
		return
	}
	scheduler.StartAsync()
}

func BookPool() {
	msg := "Name: Yose\nUnit: Redwood 15D\nBooking time: 17:00 WIB for 1 hour\nUser: 1 orang"
	t, err := client.SendMessage(types.JID{
		User:   "6281585627181",
		Server: types.DefaultUserServer,
	}, "", &waProto.Message{
		Conversation: proto.String(msg),
	})
	if err != nil {
		fmt.Printf("---[Main][SendMessage] Time: %s Err: %+v---\n", t.Format("RFC850"), err)
	}
	fmt.Printf("---[Main][SendMessage] Time: %s Err: %+v---\n", t.Format("RFC850"), err)
}
